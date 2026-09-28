package buffs

import (
	"math"
	"time"

	"github.com/wowsims/forever/sim/core"
)

func ApplyFixedUptimeAura(aura *core.Aura, uptime float64, tickLength time.Duration, startTime time.Duration) {
	auraDuration := aura.Duration
	ticksPerAura := float64(auraDuration) / float64(tickLength)
	chancePerTick := core.TernaryFloat64(uptime == 1, 1, 1.0-math.Pow(1-uptime, 1/ticksPerAura))

	aura.Unit.RegisterResetEffect(func(sim *core.Simulation) {
		core.StartPeriodicAction(sim, core.PeriodicActionOptions{
			Period: tickLength,
			OnAction: func(sim *core.Simulation) {
				if sim.RandomFloat("FixedAura") < chancePerTick {
					aura.Activate(sim)
					if aura.MaxStacks > 0 {
						aura.AddStack(sim)
					}
				}
			},
		})

		// Also try once at the start.
		core.StartPeriodicAction(sim, core.PeriodicActionOptions{
			Period:   startTime,
			NumTicks: 1,
			OnAction: func(sim *core.Simulation) {
				if sim.RandomFloat("FixedAura") < uptime {
					// Use random duration to compensate for increased chance collapsed into single tick.
					randomDur := tickLength + time.Duration(float64(auraDuration-tickLength)*sim.RandomFloat("FixedAuraDur"))

					aura.Duration = randomDur
					aura.Activate(sim)
					aura.Duration = auraDuration
				}
			},
		})
	})
}

func applyFixedShoutAura(char *core.Character, aura *core.Aura, category string) {
	aura.ApplyOnInit(func(aura *core.Aura, sim *core.Simulation) {
		auras := char.GetAurasWithTag(category)
		var playerAura *core.Aura
		for _, bsAura := range auras {
			if bsAura.ActionID.Tag == 0 {
				playerAura = bsAura
				break
			}
		}

		if playerAura == nil {
			return
		}

		playerAura.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
			pa := sim.GetConsumedPendingActionFromPool()
			pa.NextActionAt = playerAura.ExpiresAt() + char.ReactionTime
			pa.OnAction = func(sim *core.Simulation) {
				aura.Activate(sim)

				core.StartPeriodicAction(sim, core.PeriodicActionOptions{
					Period:   aura.Duration + 1,
					NumTicks: 1,
					OnAction: func(sim *core.Simulation) {
						aura.Activate(sim)
					},
				})
			}
			sim.AddPendingAction(pa)
		})
	})

	ApplyFixedUptimeAura(aura, 1, aura.Duration+1, -1)
}

func ScheduledAura(aura *core.Aura, options core.PeriodicActionOptions) {
	aura.OnReset = func(aura *core.Aura, sim *core.Simulation) {
		aura.Duration = core.NeverExpires
		core.StartPeriodicAction(sim, options)
	}
}

func SlowAura(target *core.Unit) *core.Aura {
	return castSlowReductionAura(target, "Slow", 31589, 1.5, time.Second*15)
}

func castSlowReductionAura(target *core.Unit, label string, spellID int32, multiplier float64, duration time.Duration) *core.Aura {
	aura := target.GetOrRegisterAura(core.Aura{Label: label, ActionID: core.ActionID{SpellID: spellID}, Duration: duration})
	aura.NewExclusiveEffect("CastSpdReduction", false, core.ExclusiveEffect{
		// How far from 1 the applied factor is, which is the scale every member
		// of the category bids on: a 33% slow outbids a 20% one. The factor is
		// 1/multiplier, so a caller stating 1.5 is a 33.3% slow.
		Priority: 1 - 1/multiplier,
		OnGain: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.MultiplyCastSpeed(sim, 1/multiplier)
			ee.Aura.Unit.MultiplyRangedSpeed(sim, 1/multiplier)
		},
		OnExpire: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.MultiplyCastSpeed(sim, multiplier)
			ee.Aura.Unit.MultiplyRangedSpeed(sim, multiplier)
		},
	})
	return aura
}

// A slow on casts alone, in the category Slow's cast and ranged slow takes: only the strongest applies.
func CastSpeedReductionEffect(aura *core.Aura, castTimeMultiplier float64) *core.ExclusiveEffect {
	return aura.NewExclusiveEffect("CastSpdReduction", false, core.ExclusiveEffect{
		Priority: 1 - 1/castTimeMultiplier,
		OnGain: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.MultiplyCastSpeed(sim, 1/castTimeMultiplier)
		},
		OnExpire: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.MultiplyCastSpeed(sim, castTimeMultiplier)
		},
	})
}

func AtkSpeedReductionEffect(aura *core.Aura, speedMultiplier float64) *core.ExclusiveEffect {
	return aura.NewExclusiveEffect("AtkSpdReduction", false, core.ExclusiveEffect{
		// How far from 1 the applied factor is, which is the scale every member
		// of the category bids on: a 20% slow outbids a 10% one. The factor is
		// 1/speedMultiplier, so a helper stating 1.2 is a 16.67% slow.
		Priority: 1 - 1/speedMultiplier,
		OnGain: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.MultiplyAttackSpeed(sim, 1/speedMultiplier)
		},
		OnExpire: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			ee.Aura.Unit.MultiplyAttackSpeed(sim, speedMultiplier)
		},
	})
}
