package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (warlock *Warlock) registerSearingPain() {
	rankedSpell := spellData.SearingPain.ByID(17923)
	searingPainHit := rankedSpell.DamageEffect()

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellSearingPain

	// Roll the damage and to hit
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, searingPainHit.Average(), spell.OutcomeMagicHitAndCrit)
	}

	// Calculate expected damage from the row's average hit
	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, searingPainHit.Average(), spell.OutcomeExpectedMagicHitAndCrit)
	}

	warlock.SearingPain = warlock.RegisterSpell(config)
}
