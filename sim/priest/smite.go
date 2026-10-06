package priest

import (
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var SmiteRankMap = spellData.Smite

func (priest *Priest) registerSmiteSpell(rank *spelldata.Spell) {
	baseDamage := rank.DamageEffect()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = PriestSpellSmite

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		multiplier := priest.powerInLightMultiplier(target)
		spell.DamageMultiplier *= multiplier
		spell.CalcAndDealDamage(sim, target, baseDamage.Roll(sim), spell.OutcomeMagicHitAndCrit)
		spell.DamageMultiplier /= multiplier
	}

	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, baseDamage.Average(), spell.OutcomeExpectedMagicHitAndCrit)
	}

	priest.RegisterSpell(config)
}

func (priest *Priest) powerInLightMultiplier(target *core.Unit) float64 {
	if priest.Talents.PowerInLight == 0 {
		return 1
	}
	burning := slices.ContainsFunc(priest.holyFires, func(holyFire *core.Spell) bool {
		return holyFire.Dot(target).IsActive()
	})
	if !burning {
		return 1
	}
	return spellData.PowerInLight.MultiplierAt(priest.Talents.PowerInLight)
}
