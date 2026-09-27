package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (warlock *Warlock) registerDrainSoul() {
	rankedSpell := spellData.DrainSoul.ByID(11675)
	drainSoulTick := rankedSpell.PeriodicEffect()

	soulSiphonMult := warlock.soulSiphonPerEffect()

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellDrainSoul

	config.Dot = spelldata.DotConfig(rankedSpell, drainSoulTick)
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		result := dot.Spell.CalcPeriodicDamage(sim, target, drainSoulTick.Average(dot.Spell.Unit.Level), rankedSpell.TickOutcomeHitRolled(dot))
		scaleResult(sim, dot.Spell, result, warlock.soulSiphonMultiplier(target, soulSiphonMult))
		dot.Spell.DealPeriodicDamage(sim, result)
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
		spell.DealOutcome(sim, result)
	}

	config.ExpectedTickDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		result := spell.CalcPeriodicDamage(sim, target, drainSoulTick.Average(spell.Unit.Level), spell.OutcomeExpectedMagicAlwaysHit)
		result.Damage *= warlock.soulSiphonMultiplier(target, soulSiphonMult)
		return result
	}

	warlock.DrainSoul = warlock.RegisterSpell(config)
}
