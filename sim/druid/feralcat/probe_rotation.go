package feralcat

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/druid"
)

// probe is a rotation sweep. The shipped rotation leaves it disabled.
type rotationProbe struct {
	enabled bool
	// shiftBelow casts Shifting Power at this energy or lower. Negative never casts it.
	shiftBelow float64
	// shiftBelowShred casts it whenever energy is short of Shred, whatever Shred costs.
	shiftBelowShred bool
	// spBeforeBuilder casts it even when Shred is affordable.
	spBeforeBuilder bool
	// rakeFirst refreshes Rake before Shred so Rend and Tear stays up.
	rakeFirst bool
	bearweave bool
	maulWeave bool
}

var probe rotationProbe

func (cat *FeralDruid) probeRotation(sim *core.Simulation) bool {
	// Shifting Power when energy is short of Shred, and Rake before Shred when that
	// option is on. A sweep can replace the policy for one test.
	policy := probe
	if !policy.enabled {
		policy = rotationProbe{shiftBelow: -1, shiftBelowShred: true, rakeFirst: true}
	}
	return cat.act(sim, policy)
}

func (cat *FeralDruid) act(sim *core.Simulation, policy rotationProbe) bool {
	if !cat.GCD.IsReady(sim) {
		return false
	}
	if cat.bearweave != 0 {
		return cat.continueBearweave(sim, policy)
	}
	if !cat.InForm(druid.Cat) {
		return cat.CatForm.Cast(sim, nil)
	}
	if cat.readyToShift && cat.shift(sim) {
		return true
	}

	rotation := &cat.Rotation
	if rotation.MaintainFaerieFire && cat.FaerieFire != nil &&
		cat.FaerieFireAuras != nil &&
		!cat.FaerieFireAuras.Get(cat.CurrentTarget).IsActive() &&
		cat.FaerieFire.CanCast(sim, cat.CurrentTarget) {
		return cat.FaerieFire.Cast(sim, cat.CurrentTarget)
	}

	energy := cat.CurrentEnergy()
	cp := cat.ComboPoints()
	ripUp := cat.Rip.CurDot().IsActive()
	remaining := sim.GetRemainingDuration()
	omen := cat.ClearcastingAura.IsActive()
	shredCost := cat.CurrentShredCost()
	ripCost := cat.CurrentRipCost()
	biteCost := cat.CurrentFerociousBiteCost()

	ripNow := !rotation.BiteOverRip && cp >= rotation.RipCP && !ripUp && remaining >= RipEndThresh
	biteNow := rotation.UseBite && cp >= rotation.BiteCP && remaining >= time.Second &&
		(ripUp || rotation.BiteOverRip)

	if ripNow && (energy >= ripCost || omen) && cat.Rip.CanCast(sim, cat.CurrentTarget) {
		return cat.Rip.Cast(sim, cat.CurrentTarget)
	}
	if biteNow && (energy >= biteCost || omen) && cat.FerociousBite.CanCast(sim, cat.CurrentTarget) {
		return cat.FerociousBite.Cast(sim, cat.CurrentTarget)
	}

	canShred := (energy >= shredCost || omen) && cat.Shred.CanCast(sim, cat.CurrentTarget)
	rakeDown := !cat.Rake.CurDot().IsActive()
	canRake := rotation.UseRakeTrick && rakeDown &&
		(energy >= cat.Rake.Cost.GetCurrentCost() || omen) &&
		cat.Rake.CanCast(sim, cat.CurrentTarget)
	spNow := cat.ShiftingPower.CanCast(sim, cat.CurrentTarget) &&
		(energy <= policy.shiftBelow || (policy.shiftBelowShred && energy < shredCost))
	if policy.rakeFirst && canRake {
		return cat.Rake.Cast(sim, cat.CurrentTarget)
	}
	if policy.spBeforeBuilder && spNow {
		return cat.queueShift(sim)
	}
	if canShred {
		return cat.Shred.Cast(sim, cat.CurrentTarget)
	}
	if canRake {
		return cat.Rake.Cast(sim, cat.CurrentTarget)
	}
	if spNow {
		return cat.queueShift(sim)
	}
	if policy.bearweave && energy < shredCost && cat.PrimalBite.IsReady(sim) && cat.Talents.PrimalBite {
		cat.bearweave = 1
		cat.bearweaveAt = sim.CurrentTime
		if cat.DireBearForm.Cast(sim, nil) {
			return true
		}
		cat.bearweave = 0
	}

	cat.waitForProbe(sim, cat.NextEnergyTickAt()+cat.ReactionTime)
	return false
}

func (cat *FeralDruid) queueShift(sim *core.Simulation) bool {
	if cat.shift(sim) {
		return true
	}
	if cat.readyToShift {
		cat.waitForProbe(sim, sim.CurrentTime+cat.ReactionTime)
	}
	return false
}

func (cat *FeralDruid) waitForProbe(sim *core.Simulation, at time.Duration) {
	if at < cat.GCD.ReadyAt() {
		at = cat.GCD.ReadyAt()
	}
	cat.WaitUntil(sim, at)
}

func (cat *FeralDruid) continueBearweave(sim *core.Simulation, policy rotationProbe) bool {
	if !cat.InForm(druid.Bear) {
		cat.bearweave = 0
		return cat.CatForm.Cast(sim, nil)
	}
	switch cat.bearweave {
	case 1:
		if cat.PrimalBite.CanCast(sim, cat.CurrentTarget) {
			cast := cat.PrimalBite.Cast(sim, cat.CurrentTarget)
			if policy.maulWeave && cat.Maul.CanCast(sim, cat.CurrentTarget) {
				cat.Maul.Cast(sim, cat.CurrentTarget)
				cat.bearweave = 2
				wake := cat.AutoAttacks.NextAttackAt() + time.Millisecond
				if wake < sim.CurrentTime+100*time.Millisecond {
					wake = sim.CurrentTime + 2500*time.Millisecond
				}
				cat.waitForProbe(sim, wake)
				return cast
			}
			// The bite took the GCD. The next wake shifts back to cat.
			cat.bearweave = 2
			return cast
		}
		if sim.CurrentTime-cat.bearweaveAt > 6*time.Second || !cat.PrimalBite.IsReady(sim) {
			cat.bearweave = 0
			return cat.CatForm.Cast(sim, nil)
		}
		cat.waitForProbe(sim, sim.CurrentTime+200*time.Millisecond)
		return false
	default:
		cat.bearweave = 0
		return cat.CatForm.Cast(sim, nil)
	}
}
