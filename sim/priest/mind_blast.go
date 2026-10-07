package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var MindBlastRankMap = spellData.MindBlast

func (priest *Priest) registerMindBlastSpell(_ int32, rank *spelldata.Spell) {
	baseDamage := rank.DamageEffect()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = PriestSpellMindBlast

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, baseDamage.Average(), spell.OutcomeMagicHitAndCrit)
	}

	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, baseDamage.Average(), spell.OutcomeExpectedMagicHitAndCrit)
	}

	priest.RegisterSpell(config)
}
