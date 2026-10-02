package shared

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Target resistance a temporary enchant ignores on one school's spells.
type SchoolPiercing struct {
	School stats.SchoolIndex
	Amount float64
}

// A temporary enchant consumable, applied to each hand core put its enchant on: the enchant's stats
// once per hand and each combat spell cast off that hand's hits at the chance its row states. Its flat
// weapon damage is core's, on every weapon built from that hand's item. The equip spells named are
// the enchant's that grant no stat of the enchant's own, such as Consecrated Sharpening Stone's 28893:
// their auras are parsed onto one permanent aura once per hand, since each weapon carries its own copy
// of the equip aura, and an aura bound to a matching weapon reaches that hand's attacks alone. An
// equip spell that procs, such as Scroll of Imbue Manablade's 1308913, is a listener built from its
// row instead. The piercing is a stat slot of the enchant that has no place in the stat array, added
// once per hand too.
func NewSpellDataTemporaryEnchant(itemID int32, piercing SchoolPiercing, equipSpellIDs ...int32) {
	core.NewSpellDataConsumableEffect(itemID, func(agent core.Agent) {
		character := agent.GetCharacter()
		consumable := core.GetConsumableByID(itemID)
		hands := applyTemporaryEnchant(character, consumable)
		if len(hands) == 0 {
			return
		}
		if piercing.Amount != 0 {
			character.PseudoStats.SchoolSpellPiercing[piercing.School] += piercing.Amount * float64(len(hands))
		}
		for _, spellID := range equipSpellIDs {
			row := spelldata.MustFind(spellID)
			if row.FirstAura(dbcenums.A_PROC_TRIGGER_SPELL) != spelldata.NilEffect {
				registerTemporaryEnchantProc(character, consumable, row, hands)
				continue
			}
			aura := character.GetOrRegisterAura(spelldata.AuraConfig(row, spelldata.Permanent()))
			for _, hand := range hands {
				spelldata.ParseEffects(character, aura, row, spelldata.OnHand(hand), spelldata.Count(1))
			}
		}
	})
}

// Returns the melee hands carrying the enchant, as proc masks.
func applyTemporaryEnchant(character *core.Character, consumable core.Consumable) []core.ProcMask {
	enchant := consumable.TemporaryEnchant
	if enchant == nil {
		return nil
	}

	var hands []core.ProcMask
	for _, slot := range []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand} {
		if character.Equipment[slot].TempEnchant != enchant.EffectID {
			continue
		}
		hands = append(hands, character.GetProcMaskForWeaponSlot(slot))
		character.AddStats(enchant.Stats.Add(stats.FromPseudoStatsProto(enchant.PseudoStats)))
	}
	if len(hands) == 0 {
		return nil
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

// Why a temporary enchant's equip spell that procs cannot be built, or nil where it can: its row has
// to state the hits it hears and a rate, and the spell it fires has to put mana, rage or energy into
// the wearer's bar at once or change the cooldown of one of the wearer's spell categories.
func TemporaryEnchantProcUnsupported(row *spelldata.Spell) []string {
	unsupported := spelldata.EnchantAuraUnsupported(row)
	if len(row.Effects) > 1 {
		unsupported = append(unsupported, fmt.Sprintf("%d states effects beside its proc, which the listener leaves out", row.ID))
	}
	fired := row.FirstAura(dbcenums.A_PROC_TRIGGER_SPELL).Trigger()
	gain, cooldown := fired.EnergizeEffect(), categoryCooldownEffect(fired)

	switch {
	case fired == spelldata.Nil:
		unsupported = append(unsupported, fmt.Sprintf("%d fires no spell the store carries", row.ID))
	case gain != spelldata.NilEffect:
		if gain.Target[0] != dbcenums.TARGET_UNIT_CASTER {
			unsupported = append(unsupported, fmt.Sprintf("%d's gain lands on implicit target %d, not the wearer", fired.ID, gain.Target[0]))
		}
		switch dbcenums.PowerType(gain.Misc) {
		case dbcenums.POWER_MANA, dbcenums.POWER_RAGE, dbcenums.POWER_ENERGY:
		default:
			unsupported = append(unsupported, fmt.Sprintf("%d fills power type %d, not mana, rage or energy", fired.ID, gain.Misc))
		}
	case cooldown != spelldata.NilEffect:
		if cooldown.Target[0] != dbcenums.TARGET_UNIT_CASTER {
			unsupported = append(unsupported, fmt.Sprintf("%d changes the cooldowns of implicit target %d, not the wearer", fired.ID, cooldown.Target[0]))
		}
	default:
		unsupported = append(unsupported, fmt.Sprintf("%d fires %d, which neither restores a resource nor changes a spell category's cooldown", row.ID, fired.ID))
	}
	return unsupported
}

// An equip spell that procs, built from its row: the hits it hears, its chance and its internal
// cooldown. A listener on the wearer's own hits hears the hands carrying the enchant alone; one on
// the hits the wearer takes hears every one of them.
func registerTemporaryEnchantProc(character *core.Character, consumable core.Consumable, row *spelldata.Spell, hands []core.ProcMask) {
	actionID := core.ActionID{ItemID: consumable.Id}
	handler := temporaryEnchantProcHandler(character, actionID, row.FirstAura(dbcenums.A_PROC_TRIGGER_SPELL).Trigger())
	if handler == nil {
		return
	}

	config := spelldata.ProcTrigger(character, row, handler)
	if config.Callback.Matches(core.CallbackOnSpellHitDealt) {
		var enchanted core.ProcMask
		for _, hand := range hands {
			enchanted |= hand
		}
		config.ProcMask &= enchanted
	}
	config.Name = consumable.Name
	config.ActionID = actionID

	character.MakeProcTriggerAura(config)
}

// What the fired spell does, or nil where the wearer has no bar its gain fills. A cooldown change
// moves the whole category's timer, which is where a spell whose only recovery is its category's
// keeps its cooldown; the game files state a reduction as a negative time, Spark's 1306610 -100000.
func temporaryEnchantProcHandler(character *core.Character, actionID core.ActionID, fired *spelldata.Spell) core.ProcHandler {
	if gain := fired.EnergizeEffect(); gain != spelldata.NilEffect {
		power := dbcenums.PowerType(gain.Misc)
		bar, ok := energizedBarOf(character, power, actionID)
		if !ok {
			return nil
		}
		return func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			amount := gain.Roll(sim)
			if power.InTenths() {
				amount /= 10
			}
			bar.add(sim, amount, bar.metrics)
		}
	}

	cooldown := categoryCooldownEffect(fired)
	timer := character.CategoryTimer(cooldown.Misc)
	return func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		timer.Set(max(sim.CurrentTime, timer.ReadyAt()+cooldown.TimeValue()))
	}
}

func categoryCooldownEffect(s *spelldata.Spell) *spelldata.Effect {
	for i := range s.Effects {
		if s.Effects[i].Type == dbcenums.E_MODIFY_COOLDOWNS_BY_CATEGORY {
			return &s.Effects[i]
		}
	}
	return spelldata.NilEffect
}
