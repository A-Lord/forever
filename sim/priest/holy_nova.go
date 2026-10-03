package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Causes an explosion of holy light around the caster, causing X to Y Holy damage to all enemy targets within 10 yards and
// healing all party members within 10 yards for A to B.
// These effects cause no threat. Each time Holy Fire deals damage, you have a 5% chance for your next Holy Nova to cost no Mana.
func (priest *Priest) registerHolyNovaSpell(rank *spelldata.Spell) {
	// panic("to be implemented")

	actionID := core.ActionID{SpellID: rank.ID}
	baseDamage := rank.DamageEffect()
	// baseheal := rank.HealEffect()

	// healthMetrics := priest.NewHealthMetrics(actionID.WithTag(1))
	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))

	config.ActionID = actionID
	config.ClassSpellMask = PriestSpellHolyNova

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, baseDamage.Roll(sim), spell.OutcomeMagicHitAndCrit)
		// priest.GainHealth(sim, baseheal.Roll(sim), healthMetrics)
	}

	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, baseDamage.Average(), spell.OutcomeExpectedMagicHitAndCrit)
	}

	priest.RegisterSpell(config)
}
