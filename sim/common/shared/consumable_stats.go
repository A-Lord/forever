package shared

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// A consumable whose on-use spell puts auras on the user that the parser attaches: an elixir, a
// flask, a scroll, food's Well Fed. Each spell that carries such auras, the on-use spell's own or one
// it triggers on the user, becomes an aura of its own.
//
// A buff that outlasts the item's cooldown is up from the pull: for the whole iteration where the row
// persists through death, and for the row's duration otherwise. A scroll is up for the whole
// iteration too, and its stat is exclusive with the raid buffs that grant the same stat. A buff its
// cooldown outlasts is used in combat, on that cooldown.
func NewSpellDataStatConsumable(itemID int32) {
	core.NewDataDrivenConsumableEffect(itemID, func(agent core.Agent) {
		character := agent.GetCharacter()
		consumable := core.GetConsumableByID(itemID)
		if len(consumable.ItemEffects) == 0 {
			panic(fmt.Sprintf("consumable %d states no on-use spell", itemID))
		}
		itemEffect := consumable.ItemEffects[0]

		buffs := consumableBuffs(spelldata.MustFind(itemEffect.BuffId))
		if len(buffs) == 0 {
			panic(fmt.Sprintf("consumable %d (%s): spell %d puts no aura the parser attaches on the user",
				itemID, consumable.Name, itemEffect.BuffId))
		}

		inCombat := slices.ContainsFunc(buffs, func(buff *spelldata.Spell) bool {
			return buff.DurationMs > 0 && buff.Duration() < onUseCooldown(itemEffect)
		})

		auras := make([]*core.Aura, len(buffs))
		for i, buff := range buffs {
			auras[i] = consumableAura(character, consumable, buff, i, len(buffs), inCombat)
		}

		if inCombat {
			registerConsumableOnUse(character, consumable, itemEffect, auras)
		}
	})
}

// The spells whose auras the consumable puts on its user, which is the on-use spell itself and each
// spell it triggers on the user, less the eating and drinking that only last while seated. A
// periodic trigger with a value hands its amount to every effect of the spell it triggers: Nutritious
// Food states Well Fed's Strength that way.
func consumableBuffs(row *spelldata.Spell) []*spelldata.Spell {
	var buffs []*spelldata.Spell
	if putsAurasOnTheUser(row) {
		buffs = append(buffs, row)
	}

	for i := range row.Effects {
		e := &row.Effects[i]
		if e.TriggerID == 0 || !landsOnTheUser(e) {
			continue
		}

		var trigger *spelldata.Spell
		switch {
		case e.Aura == dbcenums.A_PERIODIC_TRIGGER_SPELL_WITH_VALUE:
			trigger = withValue(e.Trigger(), e.BasePoints)
		case e.Type == dbcenums.E_TRIGGER_SPELL, e.Aura == dbcenums.A_PERIODIC_TRIGGER_SPELL:
			trigger = e.Trigger()
		default:
			continue
		}

		if putsAurasOnTheUser(trigger) {
			buffs = append(buffs, trigger)
		}
	}
	return buffs
}

func putsAurasOnTheUser(s *spelldata.Spell) bool {
	if s == spelldata.Nil || s.WhileSeated() {
		return false
	}
	return slices.ContainsFunc(s.Effects, func(e spelldata.Effect) bool {
		return e.ParsedAura() && landsOnTheUser(&e)
	})
}

// The user, or the one friendly unit the spell is cast on, which in the sim is the user: a scroll
// states the latter.
func landsOnTheUser(e *spelldata.Effect) bool {
	return slices.Contains([]dbcenums.ImplicitTarget{dbcenums.TARGET_UNIT_CASTER, dbcenums.TARGET_UNIT_TARGET_ALLY,
		dbcenums.TARGET_UNIT_TARGET_ANY}, e.Target[0])
}

func withValue(s *spelldata.Spell, value float64) *spelldata.Spell {
	valued := *s
	valued.Effects = slices.Clone(s.Effects)
	for i := range valued.Effects {
		valued.Effects[i].BasePoints = value
	}
	return &valued
}

func onUseCooldown(itemEffect *proto.ItemEffect) time.Duration {
	onUse := itemEffect.GetOnUse()
	return time.Duration(max(onUse.GetCooldownMs(), onUse.GetCategoryCooldownMs())) * time.Millisecond
}

func consumableAura(character *core.Character, consumable core.Consumable, buff *spelldata.Spell, index int, count int, inCombat bool) *core.Aura {
	config := spelldata.AuraConfig(buff, spelldata.Label(consumable.Name))
	config.ActionID = core.ActionID{ItemID: consumable.Id}
	if count > 1 {
		config.Label = consumable.Name + " - " + buff.Name
		config.ActionID.Tag = int32(index + 1)
	}

	scroll := consumable.Type == proto.ConsumableType_ConsumableTypeScroll
	if !inCombat {
		config.BuildPhase = core.CharacterBuildPhaseConsumes
		if buff.PersistsThroughDeath() || scroll || buff.DurationMs < 0 {
			spelldata.Permanent()(&config)
		} else {
			config.OnReset = func(aura *core.Aura, sim *core.Simulation) {
				aura.Activate(sim)
			}
		}
	}

	aura := character.GetOrRegisterAura(config)
	if !scroll {
		spelldata.ParseEffects(character, aura, buff)
		return aura
	}

	var exclusive []int32
	for i := range buff.Effects {
		e := &buff.Effects[i]
		for _, stat := range e.FlatStats() {
			core.MakeExclusiveFlatStatBuff(aura, stat, e.Average(character.Level), core.StatBuffCategory)
		}
		if len(e.FlatStats()) > 0 {
			exclusive = append(exclusive, int32(i+1))
		}
	}
	spelldata.ParseEffects(character, aura, buff, spelldata.SkipEffects(exclusive...))
	return aura
}

// The item used from the bar on its own cooldown and the category cooldown it shares, with the cast
// time and global cooldown its spell states.
func registerConsumableOnUse(character *core.Character, consumable core.Consumable, itemEffect *proto.ItemEffect, auras []*core.Aura) {
	row := spelldata.MustFind(itemEffect.BuffId)
	onUse := onUseCast(character, itemEffect)
	cast := spelldata.Cast(row)
	if onUse.CD.Duration > 0 {
		cast.CD = onUse.CD
	}
	cast.SharedCD = onUse.SharedCD

	spell := character.RegisterSpell(core.SpellConfig{
		ActionID: core.ActionID{ItemID: consumable.Id},
		Flags:    core.SpellFlagNoOnCastComplete,
		Cast:     cast,
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			for _, aura := range auras {
				aura.Activate(sim)
			}
		},
	})
	spell.RelatedSelfBuff = auras[0]

	character.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeDPS,
	})
}
