package reforgeoptimizer

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// Turns the request's enchant options into LP variables: one binary per (slot,
// enchant) and a per-slot row allowing at most one of them.

type enchantCandidate struct {
	slot     proto.ItemSlot
	enchant  core.Enchant
	equipped bool // The candidate is the slot's enchant in the input gear.
}

// The objective bonus that breaks an EP tie toward giving a slot an enchant; the slot's equipped
// enchant gets twice as much, so a tie keeps it. It sits far below any real EP difference and above
// HiGHS's default absolute MIP gap (1e-6); the reported score leaves it out.
const enchantTieBreak = 1e-4

func enchantVariableKey(slot proto.ItemSlot, effectID int32) string {
	return fmt.Sprintf("Enchant_%d_%d", int(slot), effectID)
}

func enchantSlotKey(slot proto.ItemSlot) string {
	return fmt.Sprintf("EnchantSlot_%d", int(slot))
}

// The static enchants that fit each listed, unfrozen, occupied slot, in slot then effect id order.
// The fitting row gives the stats, since an effect id can carry rows for different items. A slot
// holding a proc enchant the sim implements is simulated, not optimized, so it lists none.
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
		equipped := gear.GetItems()[slot].GetEnchant()
		if row, ok := core.FittingEnchantRow(equipped, *item); ok && row.ProcKind != proto.EnchantProcKind_EnchantProcKindNone && core.HasEnchantEffect(equipped) {
			continue
		}
		ids := slices.Clone(option.GetEffectIds())
		slices.Sort(ids)
		for _, id := range slices.Compact(ids) {
			if row, ok := core.FittingEnchantRow(id, *item); ok && row.IsStatic() {
				candidates = append(candidates, enchantCandidate{slot: slot, enchant: row, equipped: id == equipped})
			}
		}
	}
	return candidates
}

// Clears every candidate slot's static enchant, which makes the gear the
// baseline. An enchant that is not static gives no stats, so it stays and a slot with no pick
// keeps it.
func clearStaticEnchants(gear *proto.EquipmentSpec, candidates []enchantCandidate) {
	slots := make([]proto.ItemSlot, 0, len(candidates))
	for _, candidate := range candidates {
		slots = append(slots, candidate.slot)
	}
	for _, slot := range slices.Compact(slots) {
		item := gear.Items[slot]
		if row, ok := core.FittingEnchantRow(item.GetEnchant(), *core.GetItemByID(item.GetId())); ok && !row.IsStatic() {
			continue
		}
		item.Enchant = 0
	}
}

// The candidate's share of enchantTieBreak. A slot that keeps an enchant which is not static
// already has one, so a candidate there gets no bonus for filling it.
func (o *reforgeOptimizer) tieBreak(candidate enchantCandidate) float64 {
	switch {
	case candidate.equipped:
		return 2 * enchantTieBreak
	case o.baseGear.GetItems()[candidate.slot].GetEnchant() == 0:
		return enchantTieBreak
	default:
		return 0
	}
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
		capCoeffs := o.resolveCapCoeffs(capDelta)
		capCoeffs[enchantSlotKey(candidate.slot)] = 1

		objCoeffs := map[string]float64{tieBreakCoeffKey: o.tieBreak(candidate)}
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

// Adds pseudo-stats to an objective coefficient map the way Stats.computeEP in
// ui/sim/proto/stats.ts values them:
//   - The part of a rating's melee, ranged and spell percents that all three channels share counts
//     once under the rating when the rating carries EP, otherwise under each channel.
//   - A channel's percent beyond that shared part, and any other percent with a root rating, counts
//     under its own key when that key carries EP. Otherwise it falls back to EP through the rating:
//     for a channel, what the rating is worth beyond the channels' weights; for any other percent,
//     the rating's full worth.
//   - One rating's fallbacks count once, as the largest, under the rating's key.
//   - Any other pseudo-stat counts under its own key.
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

// The rating a percent outside the rating channels belongs to, as
// UnitStat.getRootStat gives it: hit and crit rating for the off hand's percents, and the expertise,
// dodge, parry and block ratings for theirs.
func percentRootRating(pseudoStat proto.PseudoStat) (stats.Stat, bool) {
	for _, entry := range ratingPercents {
		if entry.root && entry.percent == pseudoStat {
			return entry.rating, true
		}
	}
	return 0, false
}

var weaponDpsPseudoStat = map[proto.ItemSlot]proto.PseudoStat{
	proto.ItemSlot_ItemSlotMainHand: proto.PseudoStat_PseudoStatMainHandDps,
	proto.ItemSlot_ItemSlotOffHand:  proto.PseudoStat_PseudoStatOffHandDps,
	proto.ItemSlot_ItemSlotRanged:   proto.PseudoStat_PseudoStatRangedDps,
}

// The enchant's weapon damage as bonus DPS on the weapon it sits on. A swing deals
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
