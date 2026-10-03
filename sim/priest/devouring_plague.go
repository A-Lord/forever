package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Devouring Plague - Undead Racial
// Shadow school DoT, 3 min cooldown, 24s duration

var DevouringPlagueRankMap = spellData.DevouringPlague

func (priest *Priest) registerDevouringPlagueSpell(rank *spelldata.Spell, cdTimer *core.Timer) {
	// panic("To be implemented")

	actionID := core.ActionID{SpellID: rank.ID}
	tick := rank.PeriodicEffect()

	healthMetrics := priest.NewHealthMetrics(actionID.WithTag(1))

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))

	config.ActionID = actionID
	config.ClassSpellMask = PriestSpellDevouringPlague

	config.Cast.CD = core.Cooldown{
		Timer:    cdTimer,
		Duration: max(rank.Cooldown(), rank.CategoryCooldown()),
	}

	config.Dot = spelldata.DotConfig(rank, tick)
	config.Dot.Aura.Tag = "Holy Fire"
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		result := dot.Spell.CalcAndDealPeriodicDamage(sim, target, tick.Roll(sim), rank.TickOutcomeHitRolled(dot))

		priest.GainHealth(sim, result.Damage, healthMetrics)
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, tick.Roll(sim), spell.OutcomeMagicHitAndCrit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
	}

	config.ExpectedTickDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
		return spell.CalcPeriodicDamage(sim, target, tick.Average(), spell.OutcomeExpectedMagicAlwaysHit)
	}

	priest.RegisterSpell(config)
}
