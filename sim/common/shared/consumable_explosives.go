package shared

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// An explosive used from the bar: the damage its on-use spell deals, on the item's cooldown and the
// category cooldown every explosive shares, and the damage the spell deals to the thrower where the
// row states one. Only a character with the profession the item requires can use it.
func NewSpellDataExplosive(itemID int32) {
	core.NewDataDrivenConsumableEffect(itemID, func(agent core.Agent) {
		character := agent.GetCharacter()
		consumable := core.GetConsumableByID(itemID)
		if consumable.RequiredProfession != proto.Profession_ProfessionUnknown && !character.HasProfession(consumable.RequiredProfession) {
			return
		}

		for _, itemEffect := range consumable.ItemEffects {
			if itemEffect.GetOnUse() == nil {
				continue
			}
			config := onUseSpellConfig(character, itemID, itemEffect, spellDataOnUseDamageSpell)
			config.Flags |= core.SpellFlagExplosive
			if self := spelldata.MustFind(itemEffect.BuffId).CasterDamageEffect(); self != spelldata.NilEffect {
				config.ApplyEffects = withCasterDamage(character, config, self)
			}

			character.AddMajorCooldown(core.MajorCooldown{
				Spell:    character.RegisterSpell(config),
				Type:     core.CooldownTypeDPS | core.CooldownTypeExplosive,
				Priority: core.CooldownPriorityLow,
			})
		}
	})
}

// The spell's effects, then the roll of its caster effect on the thrower. That half is a spell of its
// own, tagged 1, so that it reaches the character as a harmful spell hit, which is what a listener on
// spell damage taken hears.
func withCasterDamage(character *core.Character, config core.SpellConfig, self *spelldata.Effect) core.ApplySpellResults {
	selfSpell := character.RegisterSpell(core.SpellConfig{
		ActionID:    config.ActionID.WithTag(1),
		SpellSchool: config.SpellSchool,
		DefenseType: config.DefenseType,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       core.SpellFlagExplosive,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
	})
	outcome := GetOutcome(selfSpell, damageOutcome(config.DefenseType, spelldata.Find(self.SpellID).CannotCrit(), OutcomeDefault))

	applyEffects := config.ApplyEffects
	return func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		applyEffects(sim, target, spell)
		selfSpell.CalcAndDealDamage(sim, &character.Unit, self.Roll(sim, character.Level), outcome)
	}
}
