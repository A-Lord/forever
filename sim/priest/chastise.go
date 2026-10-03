package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// TODO: To be implemented. Chastise is content Forever added; the sim has never
// modelled it, so this is a stub rather than a port.
//
// Spell IDs: 1277331, 1277332, 1277333, 1277334, 1277335
// SpellData: spellData.Chastise (5 ranks)
// SpellMask: PriestSpellChastise
//
// Not wired into Initialize() yet -- wiring an unimplemented ability would panic
// on every character of this class. Wire it when it is implemented.

var ChastiseRankMap = spellData.Chastise

func (priest *Priest) registerChastise(rank *spelldata.Spell, chastiseTimer *core.Timer) {
	// panic("To be implemented")

	chastiseDamage := rank.DamageEffect()
	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))

	config.ActionID = core.ActionID{SpellID: rank.ID}
	config.ClassSpellMask = PriestSpellChastise
	config.SpellSchool = rank.SpellSchool()
	config.DefenseType = core.DefenseTypeMagic
	config.ProcMask = core.ProcMaskSpellDamage
	config.Flags = core.SpellFlagAPL
	config.Rank = rank.RankNumber()

	config.Cast.CD = core.Cooldown{
		Timer:    chastiseTimer,
		Duration: max(rank.Cooldown(), rank.CategoryCooldown()),
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcDamage(sim, target, chastiseDamage.Roll(sim), spell.OutcomeMagicHitAndCrit)
	}

	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, chastiseDamage.Average(), spell.OutcomeExpectedMagicHitAndCrit)
	}

	priest.RegisterSpell(config)
}
