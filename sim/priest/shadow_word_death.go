package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var ShadowWordDeathRankMap = spellData.ShadowWordDeath

func (priest *Priest) registerShadowWordDeathSpell(_ int32, rank *spelldata.Spell) {
	baseDamage := rank.DamageEffect()
	backlash := rank.EffectN(3).Percent()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = PriestSpellShadowWordDeath

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, baseDamage.Roll(sim), spell.OutcomeMagicHitAndCrit)
		priest.RemoveHealth(sim, priest.MaxHealth()*backlash)
	}

	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, baseDamage.Average(), spell.OutcomeExpectedMagicHitAndCrit)
	}

	priest.RegisterSpell(config)
}
