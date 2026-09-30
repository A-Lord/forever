package spelldata

import (
	"math"
	"time"

	"github.com/wowsims/forever/sim/core"
)

// The game files' own number, before any unit the aura gives it.
func (e *Effect) BaseValue() float64 {
	return e.BasePoints
}

// The game files state a percentage as an integer: Improved Righteous Fury reads 16, not 0.16.
func (e *Effect) Percent() float64 {
	return e.BasePoints / 100
}

// The game files state rage on a 0-1000 bar: a -30 cost modifier is 3 rage, Charge's energize of
// 150 is 15.
func (e *Effect) Tenths() float64 {
	return e.BasePoints / 10
}

// The value read as a time, which is the unit a duration or delay modifier states it in.
func (e *Effect) TimeValue() time.Duration {
	return core.DurationFromMillis(e.BasePoints)
}

// EffectAuraPeriod: how long one tick of the aura lasts.
func (e *Effect) Period() time.Duration {
	return core.DurationFromMillis(e.PeriodMs)
}

func (e *Effect) Coeff() float64 {
	return e.SPCoef
}

func (e *Effect) APCoeff() float64 {
	return e.APCoef
}

// The spell this effect fires, or Nil where it fires nothing.
func (e *Effect) Trigger() *Spell {
	return Find(e.TriggerID)
}

// The amount at a caster level: the base points plus the per-level gain over the spell's BaseLevel,
// stopping at the level the spell stops scaling at. Both levels are the owning spell's, stamped onto
// the effect at generation, so an effect answers the same amount however the caller reached it.
//
// The base keeps the game files' own EffectBasePointsF, which is fractional on about one effect in 55,
// and the per-level gain is added on top of the fraction before the whole is floored.
//
// float32 is load-bearing: EffectRealPointsPerLevel is a float32 widened into the DB
// (3.79999995231628), and multiplying in float64 moves the result off the tooltip on six rows. The
// game resolves the amount to a whole number, which is the floor below.
func (e *Effect) Average(level int32) float64 {
	return math.Floor(float64(float32(e.BasePoints) + e.levelGain(level)))
}

func (e *Effect) levelGain(level int32) float32 {
	// MaxLevel 0 is the game files' "no cap", which a spell with no SpellLevels row carries.
	lvl := level
	if e.MaxLevel > 0 && int32(e.MaxLevel) < lvl {
		lvl = int32(e.MaxLevel)
	}

	delta := lvl - int32(e.BaseLevel)
	if delta < 0 {
		delta = 0
	}

	return float32(float32(delta) * float32(e.PPL))
}

// The low end of the roll. EffectVariance spreads the base points, whose ends are rounded to whole
// numbers; the per-level gain is added to both, and each end is rounded outward. Starfire rank 1's
// 81 ± 8.18 is 73 to 89, and Wrath rank 2's 23 ± 1.70 is 21 to 25, which is 24 to 29 with its 3.6
// gain. An effect that does not roll answers the average at both ends.
func (e *Effect) Min(level int32) float64 {
	if e.Variance == 0 {
		return e.Average(level)
	}
	low, _ := e.spread(level)
	return low
}

func (e *Effect) Max(level int32) float64 {
	if e.Variance == 0 {
		return e.Average(level)
	}
	_, high := e.spread(level)
	return high
}

func (e *Effect) spread(level int32) (float64, float64) {
	half := math.Abs(e.BasePoints) * e.Variance / 2
	gain := float64(e.levelGain(level))
	return roundOutward(math.Round(e.BasePoints-half)+gain, math.Floor),
		roundOutward(math.Round(e.BasePoints+half)+gain, math.Ceil)
}

// The per-level gain is a float32 widened here, so an end that is whole in game can land a hair to
// either side of it (Exorcism 415068 at 25 reads 114.0000019); it is snapped before it is rounded.
func roundOutward(end float64, round func(float64) float64) float64 {
	if whole := math.Round(end); core.WithinToleranceFloat64(whole, end, 1e-4) {
		return whole
	}
	return round(end)
}

// The amount for one cast: rolled where the game files state a spread, the average where they do not.
func (e *Effect) Roll(sim *core.Simulation, level int32) float64 {
	if e.Variance == 0 {
		return e.Average(level)
	}
	return sim.Roll(e.spread(level))
}
