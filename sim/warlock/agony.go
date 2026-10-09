package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Bane of Agony ramps in thirds: the first third of the ticks at half the base, the second at the base,
// the last at one and a half. Only the base ramps; the spell power share does not.
func (warlock *Warlock) registerBaneOfAgony() {
	rankedSpell := spellData.BaneOfAgony.ByID(11713)
	agonyTick := rankedSpell.PeriodicEffect()

	// The Amplify Curse bonus each target's Bane landed with: the buff is spent by the cast, and the ticks
	// that follow keep what it gave.
	ampMultByTarget := map[int32]float64{}

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellBaneOfAgony

	config.Dot = spelldata.DotConfig(rankedSpell, agonyTick)
	config.Dot.Aura.Tag = "Affliction"

	// Note whether Amplify Curse was up. Dot.Apply snapshots on every application, whatever applied the
	// Bane, and runs inside the cast's effects, before the cast spends the buff.
	config.Dot.OnSnapshot = func(_ *core.Simulation, target *core.Unit, _ *core.Dot) {
		ampMultByTarget[target.UnitIndex] = 1
		if warlock.AmplifyCurseAura.IsActive() {
			ampMultByTarget[target.UnitIndex] = warlock.AmplifyCurseMultiplier
		}
	}

	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		// The ramp goes on the base alone; CalcPeriodicDamage adds the spell power share on top, unramped.
		baseDamage := agonyTick.Average() * baneOfAgonyRamp(dot.TickCount(), dot.BaseTickCount)
		result := dot.Spell.CalcPeriodicDamage(sim, target, baseDamage, rankedSpell.TickOutcome(dot))

		// Amplify Curse raises the whole effect, spell power included. A dot activated without Apply has
		// no entry, and takes no bonus rather than reading the map's zero.
		ampMult, ok := ampMultByTarget[target.UnitIndex]
		if !ok {
			ampMult = 1
		}
		scaleResult(sim, dot.Spell, result, ampMult)
		dot.Spell.DealPeriodicDamage(sim, result)
	}

	// Roll to hit and apply as dot in place of any other Bane
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
		if result.Landed() {
			warlock.deactivateOtherBanes(sim, spell, target)
			spell.Dot(target).Apply(sim)
		}
		spell.DealOutcome(sim, result)
	}

	// Calculate expected damage from the average tick, which is what a tick is worth over the whole duration.
	// With Amplify Curse: the dot on the target as it landed (useSnapshot), or a cast now, amplified if the
	// buff is up, so Dot Percent Increase sees what re-applying under Amplify Curse gains.
	config.ExpectedTickDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
		result := spell.CalcPeriodicDamage(sim, target, agonyTick.Average(), spell.OutcomeExpectedMagicAlwaysHit)

		ampMult := 1.0
		if useSnapshot {
			if landed, ok := ampMultByTarget[target.UnitIndex]; ok {
				ampMult = landed
			}
		} else if warlock.AmplifyCurseAura.IsActive() {
			ampMult = warlock.AmplifyCurseMultiplier
		}
		result.Damage *= ampMult
		return result
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
