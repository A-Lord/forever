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

// rebaseWithoutEnchants makes the baseline the current gear with every candidate slot's enchant
// cleared.
func (o *reforgeOptimizer) rebaseWithoutEnchants(candidates []enchantCandidate) error {
	raid := googleProto.Clone(o.baseRaidProto).(*proto.Raid)
	items := raid.Parties[0].Players[0].Equipment.Items
	for _, candidate := range candidates {
		items[candidate.slot].Enchant = 0
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
		for idx, value := range pseudoStats {
			if value != 0 {
				objCoeffs[pseudoStatCoeffKey(proto.PseudoStat(idx))] += value
			}
		}

		key := enchantVariableKey(candidate.slot, candidate.enchant.EffectID)
		variables.set(key, capCoeffs)
		variables.setObj(key, objCoeffs)
		constraints.set(enchantSlotKey(candidate.slot), lessEq(1))
	}
	return variables, constraints
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
