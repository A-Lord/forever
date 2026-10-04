package reforgeoptimizer

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	googleProto "google.golang.org/protobuf/proto"
)

// enchants.go turns the request's enchant options into LP variables: one binary per (slot,
// enchant) and a per-slot row allowing at most one of them.

type enchantCandidate struct {
	slot    proto.ItemSlot
	enchant core.Enchant
}

func enchantVariableKey(slot proto.ItemSlot, effectID int32) string {
	return fmt.Sprintf("Enchant_%d_%d", int(slot), effectID)
}

func enchantSlotKey(slot proto.ItemSlot) string {
	return fmt.Sprintf("EnchantSlot_%d", int(slot))
}

// The static enchants that fit each listed, unfrozen, occupied slot, in slot then effect id order.
// The fitting row gives the stats, since an effect id can carry rows for different items.
func (o *reforgeOptimizer) enchantCandidates(gear *proto.EquipmentSpec) []enchantCandidate {
	var candidates []enchantCandidate
	options := slices.Clone(o.request.GetEnchantOptions())
	slices.SortFunc(options, func(a, b *proto.BulkEnchantSlotOptions) int { return cmp.Compare(a.GetSlot(), b.GetSlot()) })
	for _, option := range options {
		slot := option.GetSlot()
		if o.frozenSlots[slot] || int(slot) >= len(gear.GetItems()) {
			continue
		}
		item := core.GetItemByID(gear.GetItems()[slot].GetId())
		if item == nil {
			continue
		}
		ids := slices.Clone(option.GetEffectIds())
		slices.Sort(ids)
		for _, id := range slices.Compact(ids) {
			if row, ok := core.FittingEnchantRow(id, *item); ok && row.IsStatic() {
				candidates = append(candidates, enchantCandidate{slot: slot, enchant: row})
			}
		}
	}
	return candidates
}

// rebaseWithoutEnchants makes the baseline the current gear with every candidate slot's static
// enchant cleared. An enchant that is not static gives no stats, so it stays and a slot with no
// pick keeps it.
func (o *reforgeOptimizer) rebaseWithoutEnchants(candidates []enchantCandidate) error {
	raid := googleProto.Clone(o.baseRaidProto).(*proto.Raid)
	items := raid.Parties[0].Players[0].Equipment.Items
	for _, candidate := range candidates {
		item := items[candidate.slot]
		if row, ok := core.FittingEnchantRow(item.GetEnchant(), *core.GetItemByID(item.GetId())); ok && !row.IsStatic() {
			continue
		}
		item.Enchant = 0
	}
	return o.setBaseRaid(raid)
}

func (o *reforgeOptimizer) buildEnchantModel(candidates []enchantCandidate, weights core.UnitStats) (*lpVariables, *lpConstraints) {
	variables := newLPVariables()
	constraints := newLPConstraints()
	for _, candidate := range candidates {
		pseudoStats := make([]float64, stats.PseudoStatsLen)
		copy(pseudoStats, candidate.enchant.PseudoStats)
		if dpsStat, ok := weaponDpsPseudoStat[candidate.slot]; ok {
			pseudoStats[dpsStat] += o.weaponDamageDps(candidate)
		}

		capDelta := core.NewUnitStats()
		capDelta.Stats = candidate.enchant.Stats.Add(stats.FromPseudoStatsProto(pseudoStats))
		copy(capDelta.PseudoStats, pseudoStats)
		capCoeffs := o.resolveCapCoeffsUnit(capDelta)
		capCoeffs[enchantSlotKey(candidate.slot)] = 1

		objCoeffs := map[string]float64{}
		for statIdx := 0; statIdx < int(stats.ProtoStatsLen); statIdx++ {
			if value := candidate.enchant.Stats[statIdx]; value != 0 {
				o.applyReforgeStat(objCoeffs, stats.Stat(statIdx), value, weights)
			}
		}
		applyPseudoStats(objCoeffs, pseudoStats, weights)

		key := enchantVariableKey(candidate.slot, candidate.enchant.EffectID)
		variables.set(key, capCoeffs)
		variables.setObj(key, objCoeffs)
		constraints.set(enchantSlotKey(candidate.slot), lessEq(1))
	}
	return variables, constraints
}

// applyPseudoStats adds pseudo-stats to an objective coefficient map the way Stats.computeEP in
// ui/sim/proto/stats.ts values them. The part of a rating's melee, ranged and spell percents every
// channel holds counts under the rating when the rating carries EP, otherwise under each channel.
// What a channel holds beyond it, and any other percent a rating raises, counts under its own key
// when that carries EP; otherwise it falls back to what the rating is worth past the channels' weights,
// and the fallbacks of one rating count once, as the largest, under the rating's key. Any other
// pseudo-stat counts under its own key.
func applyPseudoStats(coeffs map[string]float64, pseudoStats []float64, weights core.UnitStats) {
	fallbacks := map[stats.Stat]float64{}
	addPercent := func(rating stats.Stat, pseudoStat proto.PseudoStat, value float64, fallbackPerPercent float64) {
		if value == 0 {
			return
		}
		if getUnitStat(weights, stats.UnitStatFromPseudoStat(pseudoStat)) != 0 {
			coeffs[pseudoStatCoeffKey(pseudoStat)] += value
			return
		}
		ep := value * fallbackPerPercent
		if counted, ok := fallbacks[rating]; !ok || ep > counted {
			fallbacks[rating] = ep
		}
	}

	for _, channels := range core.RatingChannelTable {
		channelStats := channels.Channels()
		values := make([]float64, len(channelStats))
		weighted := 0.0
		for i, pseudoStat := range channelStats {
			values[i] = stats.PseudoStatValue(pseudoStats, pseudoStat)
			weighted += getUnitStat(weights, stats.UnitStatFromPseudoStat(pseudoStat))
		}
		perPercent := channels.RatingPerPercent * weights.Stats[channels.Rating]
		common := max(0, slices.Min(values))
		if common != 0 && perPercent != 0 {
			coeffs[statCoeffKey(proto.Stat(channels.Rating))] += common * channels.RatingPerPercent
		} else if common != 0 {
			for _, pseudoStat := range channelStats {
				coeffs[pseudoStatCoeffKey(pseudoStat)] += common
			}
		}
		residual := max(0, perPercent-weighted)
		for i, pseudoStat := range channelStats {
			addPercent(channels.Rating, pseudoStat, values[i]-common, residual)
		}
	}

	for idx, value := range pseudoStats {
		pseudoStat := proto.PseudoStat(idx)
		if value == 0 || isChannelPseudoStat(pseudoStat) {
			continue
		}
		if rating, ok := percentRootRating(pseudoStat); ok {
			addPercent(rating, pseudoStat, value, ratingPerPseudoStatPercent(pseudoStat, rating)*weights.Stats[rating])
		} else {
			coeffs[pseudoStatCoeffKey(pseudoStat)] += value
		}
	}

	// A fallback is EP at these weights, kept under the rating's key in rating units so it moves with
	// the rating's weight.
	for rating, ep := range fallbacks {
		if ep != 0 && weights.Stats[rating] != 0 {
			coeffs[statCoeffKey(proto.Stat(rating))] += ep / weights.Stats[rating]
		}
	}
}

func isChannelPseudoStat(pseudoStat proto.PseudoStat) bool {
	return slices.ContainsFunc(core.RatingChannelTable, func(channels core.RatingChannels) bool {
		return slices.Contains(channels.Channels(), pseudoStat)
	})
}

// percentRootRating is the rating a percent outside the rating channels belongs to, as
// UnitStat.getRootStat gives it: hit and crit rating for the off hand's percents, and the expertise,
// dodge, parry and block ratings for theirs. Defense and the per-school hit percents have none.
func percentRootRating(pseudoStat proto.PseudoStat) (stats.Stat, bool) {
	if slices.Contains(schoolHitPercents, pseudoStat) {
		return 0, false
	}
	for _, conversion := range core.RatingConversions {
		if conversion.Rating != stats.DefenseRating && slices.Contains(sheetPseudoStats(conversion.Percent), pseudoStat) {
			return conversion.Rating, true
		}
	}
	return 0, false
}

var weaponDpsPseudoStat = map[proto.ItemSlot]proto.PseudoStat{
	proto.ItemSlot_ItemSlotMainHand: proto.PseudoStat_PseudoStatMainHandDps,
	proto.ItemSlot_ItemSlotOffHand:  proto.PseudoStat_PseudoStatOffHandDps,
	proto.ItemSlot_ItemSlotRanged:   proto.PseudoStat_PseudoStatRangedDps,
}

// weaponDamageDps is the enchant's weapon damage as bonus DPS on the weapon it sits on. A swing deals
// weapon damage plus bonus DPS times the weapon's speed, so the two convert by the speed.
func (o *reforgeOptimizer) weaponDamageDps(candidate enchantCandidate) float64 {
	if candidate.enchant.WeaponDamage == 0 {
		return 0
	}
	item := core.GetItemByID(o.baseGear.GetItems()[candidate.slot].GetId())
	if item == nil || item.SwingSpeed == 0 {
		return 0
	}
	return candidate.enchant.WeaponDamage / item.SwingSpeed
}
