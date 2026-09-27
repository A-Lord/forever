package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Agony currently uses the classic implementation of "three thirds",
// with the first third at 0.5 power, the second at 1.0, and the last at 1.5.
// Current testing shows that only the base power of the spell scales in this manner
func (warlock *Warlock) registerBaneOfAgony() {
	rankedSpell := spellData.BaneOfAgony.ByID(11713)
	agonyTick := rankedSpell.PeriodicEffect()

	// The Amplify Curse bonus each target's Bane landed with, set on every application: the buff is spent
	// by the cast, and the ticks that follow keep what it gave.
	ampMultByTarget := map[int32]float64{}

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellBaneOfAgony

	config.Dot = spelldata.DotConfig(rankedSpell, agonyTick)
	config.Dot.Aura.Tag = "Affliction"
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		// The ramp goes on the base alone; CalcPeriodicDamage adds the spell power share on top, unramped.
		baseDamage := agonyTick.Average(dot.Spell.Unit.Level) * baneOfAgonyRamp(dot.TickCount(), dot.BaseTickCount)
		result := dot.Spell.CalcPeriodicDamage(sim, target, baseDamage, rankedSpell.TickOutcomeHitRolled(dot))
		// Amplify Curse raises the whole effect, spell power included.
		scaleResult(sim, dot.Spell, result, ampMultByTarget[target.UnitIndex])
		dot.Spell.DealPeriodicDamage(sim, result)
	}

	// Roll to hit, note whether Amplify Curse was up, and apply as dot in place of any other Bane
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
		if result.Landed() {
			ampMultByTarget[target.UnitIndex] = 1
			if warlock.AmplifyCurseAura.IsActive() {
				ampMultByTarget[target.UnitIndex] = warlock.AmplifyCurseMultiplier
			}
			warlock.deactivateOtherBanes(sim, spell, target)
			spell.Dot(target).Apply(sim)
		}
		spell.DealOutcome(sim, result)
	}

	// Calculate expected damage from the average tick, which is what a tick is worth over the whole duration.
	config.ExpectedTickDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcPeriodicDamage(sim, target, agonyTick.Average(spell.Unit.Level), spell.OutcomeExpectedMagicAlwaysHit)
	}

	warlock.BaneOfAgony = warlock.RegisterSpell(config)
	warlock.Banes = append(warlock.Banes, warlock.BaneOfAgony)
}

// Helper function to facilitate the percentage of base damage dealt per tick
func baneOfAgonyRamp(tick int32, ticks int32) float64 {
	third := max(ticks/3, 1)
	switch {
	case tick <= third:
		return 0.5
	case tick <= 2*third:
		return 1
	default:
		return 1.5
	}
}
