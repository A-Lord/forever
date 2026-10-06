package priest

import (
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var SmiteRankMap = spellData.Smite

func (priest *Priest) registerSmiteSpell(_ int32, rank *spelldata.Spell) {
	baseDamage := rank.DamageEffect()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = PriestSpellSmite

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		priest.calcAndDealPowerInLightDamage(sim, target, spell, baseDamage.Roll(sim))
	}

	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, baseDamage.Average(), spell.OutcomeExpectedMagicHitAndCrit)
	}

	priest.RegisterSpell(config)
}

func (priest *Priest) calcAndDealPowerInLightDamage(sim *core.Simulation, target *core.Unit, spell *core.Spell, baseDamage float64) {
	multiplier := 1.0
	burning := slices.ContainsFunc(priest.holyFires, func(holyFire *core.Spell) bool {
		return holyFire.Dot(target).IsActive()
	})
	if burning {
		multiplier = priest.powerInLightMultiplier
	}
	spell.DamageMultiplier *= multiplier
	spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
	spell.DamageMultiplier /= multiplier
}
