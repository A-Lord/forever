package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (warlock *Warlock) registerDrainLife() {
	rankedSpell := spellData.DrainLife.ByID(11700)
	drainLifeTick := rankedSpell.PeriodicEffect()

	soulSiphonMult := warlock.soulSiphonPerEffect()
	healthMetrics := warlock.NewHealthMetrics(core.ActionID{SpellID: rankedSpell.ID})

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellDrainLife

	config.Dot = spelldata.DotConfig(rankedSpell, drainLifeTick)
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		result := dot.Spell.CalcPeriodicDamage(sim, target, drainLifeTick.Average(dot.Spell.Unit.Level), rankedSpell.TickOutcomeHitRolled(dot))
		scaleResult(sim, dot.Spell, result, warlock.soulSiphonMultiplier(target, soulSiphonMult))
		dot.Spell.DealPeriodicDamage(sim, result)

		heal := result.Damage * warlock.PseudoStats.SelfHealingMultiplier * warlock.PseudoStats.HealingTakenMultiplier
		warlock.GainHealth(sim, heal, healthMetrics)
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
		spell.DealOutcome(sim, result)
	}

	config.ExpectedTickDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		result := spell.CalcPeriodicDamage(sim, target, drainLifeTick.Average(spell.Unit.Level), spell.OutcomeExpectedMagicAlwaysHit)
		result.Damage *= warlock.soulSiphonMultiplier(target, soulSiphonMult)
		return result
	}

	warlock.DrainLife = warlock.RegisterSpell(config)
}
