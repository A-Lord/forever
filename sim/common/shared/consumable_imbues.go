package shared

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// A temporary enchant consumable, applied to each hand core put its enchant on: the enchant's stats
// once per hand, its flat weapon damage on that hand's weapon, and each combat spell cast off that
// hand's hits at the chance its slot states.
func NewTemporaryEnchantConsumable(itemID int32) {
	core.NewDataDrivenConsumableEffect(itemID, func(agent core.Agent) {
		applyTemporaryEnchant(agent.GetCharacter(), core.GetConsumableByID(itemID))
	})
}

func applyTemporaryEnchant(character *core.Character, consumable core.Consumable) {
	enchant := consumable.TemporaryEnchant
	if enchant == nil {
		return
	}

	enchanted := false
	for _, slot := range []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand} {
		if character.Equipment[slot].TempEnchant != enchant.EffectID {
			continue
		}
		enchanted = true
		character.AddStats(enchant.Stats.Add(stats.FromPseudoStatsProto(enchant.PseudoStats)))
		if weapon := handWeapon(character, slot); weapon != nil && enchant.WeaponDamage != 0 {
			weapon.BaseDamageMin += enchant.WeaponDamage
			weapon.BaseDamageMax += enchant.WeaponDamage
		}
	}
	if !enchanted {
		return
	}

	for _, effect := range enchant.EnchantEffects {
		registerTemporaryEnchantCombatSpell(character, consumable, effect)
	}
}

func handWeapon(character *core.Character, slot proto.ItemSlot) *core.Weapon {
	if slot == proto.ItemSlot_ItemSlotMainHand {
		return character.AutoAttacks.MH()
	}
	return character.AutoAttacks.OH()
}

// The combat spell rolls on the hits of the weapons carrying the enchant, at their own speeds.
func registerTemporaryEnchantCombatSpell(character *core.Character, consumable core.Consumable, effect *proto.ItemEffect) {
	damage := spelldata.MustFind(effect.BuffId)
	damageSpell := character.RegisterSpell(spellDataProcDamageSpell(character, damage))
	enchantID := consumable.TemporaryEnchant.EffectID
	chance := effect.GetProc().GetProcChance()

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
