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
	return math.Floor(float64(e.scaled(level)))
}

func (e *Effect) scaled(level int32) float32 {
	// MaxLevel 0 is the game files' "no cap", which a spell with no SpellLevels row carries.
	lvl := level
	if e.MaxLevel > 0 && int32(e.MaxLevel) < lvl {
		lvl = int32(e.MaxLevel)
	}

	delta := lvl - int32(e.BaseLevel)
	if delta < 0 {
		delta = 0
	}

	return float32(e.BasePoints) + float32(float32(delta)*float32(e.PPL))
}

// The low end of the roll. EffectVariance spreads the base points only, so the spread keeps its width
// as the amount scales, and it sits around the amount before it is floored, with each end rounded
// outward. At level 14 Shadow Oil's 350026 is 64.6 ± 4, shown as 60 to 69, and Frost Oil's 350025 is
// 41.4 ± 2, shown as 39 to 44. An effect that does not roll answers the average at both ends.
func (e *Effect) Min(level int32) float64 {
	if e.Variance == 0 {
		return e.Average(level)
	}
	return e.spreadEnd(level, -1)
}

func (e *Effect) Max(level int32) float64 {
	if e.Variance == 0 {
		return e.Average(level)
	}
	return e.spreadEnd(level, 1)
}

// EffectVariance is a float32 widened into the DB, so an end that is whole in game lands a hair to
// either side of it here (47.99999985 for 350026's 48); it is snapped before it is rounded.
func (e *Effect) spreadEnd(level int32, side float64) float64 {
	end := float64(e.scaled(level)) + side*math.Abs(e.BasePoints)*e.Variance/2
	if whole := math.Round(end); math.Abs(end-whole) < 1e-4 {
		return whole
	}
	if side < 0 {
		return math.Floor(end)
	}
	return math.Ceil(end)
}

// The amount for one cast: rolled where the game files state a spread, the average where they do not.
func (e *Effect) Roll(sim *core.Simulation, level int32) float64 {
	if e.Variance == 0 {
		return e.Average(level)
	}
	return sim.Roll(e.Min(level), e.Max(level))
}
