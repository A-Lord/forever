package shared

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// A temporary enchant consumable, applied to each hand core put its enchant on: the enchant's stats
// once per hand and each combat spell cast off that hand's hits at the chance its row states. Its flat
// weapon damage is core's, on every weapon built from that hand's item. The equip spells named are
// the enchant's that grant no stat of the enchant's own, such as Consecrated Sharpening Stone's 28893:
// their auras are parsed onto one permanent aura counted once per hand, since each weapon carries its
// own copy of the equip aura.
func NewSpellDataTemporaryEnchant(itemID int32, equipSpellIDs ...int32) {
	core.NewSpellDataConsumableEffect(itemID, func(agent core.Agent) {
		character := agent.GetCharacter()
		hands := applyTemporaryEnchant(character, core.GetConsumableByID(itemID))
		if hands == 0 {
			return
		}
		for _, spellID := range equipSpellIDs {
			row := spelldata.MustFind(spellID)
			aura := character.GetOrRegisterAura(spelldata.AuraConfig(row, spelldata.Permanent()))
			spelldata.ParseEffects(character, aura, row, spelldata.Count(float64(hands)))
		}
	})
}

// Returns the number of hands carrying the enchant.
func applyTemporaryEnchant(character *core.Character, consumable core.Consumable) int {
	enchant := consumable.TemporaryEnchant
	if enchant == nil {
		return 0
	}

	hands := 0
	for _, slot := range []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand} {
		if character.Equipment[slot].TempEnchant != enchant.EffectID {
			continue
		}
		hands++
		character.AddStats(enchant.Stats.Add(stats.FromPseudoStatsProto(enchant.PseudoStats)))
	}
	if hands == 0 {
		return 0
	}

	for _, effect := range enchant.EnchantEffects {
		registerTemporaryEnchantCombatSpell(character, consumable, effect)
	}
	return hands
}

// The combat spell rolls on the hits of the weapons carrying the enchant, at their own speeds.
func registerTemporaryEnchantCombatSpell(character *core.Character, consumable core.Consumable, effect *proto.ItemEffect) {
	damage := spelldata.MustFind(effect.BuffId)
	damageSpell := character.RegisterSpell(spellDataProcDamageSpell(character, damage, true))
	enchantID := consumable.TemporaryEnchant.EffectID
	chance := damage.StatedChance()

	config := spelldata.ProcTrigger(character, damage, nil, spelldata.WeaponProc(), func(character *core.Character, trigger *core.ProcTrigger) {
		trigger.ProcChance = 0
		trigger.DPM = character.NewDynamicLegacyProcForTempEnchant(enchantID, 0, func(core.ProcMask) float64 { return chance })
	})
	config.Name = consumable.Name
	config.ActionID = core.ActionID{ItemID: consumable.Id}
	config.Handler = procDamageHandler(character, damageSpell, config.Callback)
	config.TriggerImmediately = true

	character.MakeProcTriggerAura(config)
}
