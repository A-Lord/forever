package feralcat

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/druid"
)

// Ported from https://github.com/NerdEgghead/TBC_cat_sim

const BiteTrickCP = int32(2)
const BiteTrickMax = 39.0
const BiteTime = time.Duration(0)
const RipTrickMin = 52.0
const RipEndThresh = time.Second * 10
const MaxWaitTime = time.Second * 1

type FeralDruidRotation struct {
	RipCP        int32
	BiteCP       int32
	RipTrickCP   int32
	UseBite      bool
	BiteOverRip  bool
	UseRipTrick  bool
	UseRakeTrick bool
	Wolfshead    bool

	MaintainFaerieFire bool
}

func (cat *FeralDruid) setupRotation(rotation *proto.FeralCatDruid_Rotation) {
	useBite := (rotation.Biteweave && rotation.FinishingMove == proto.FeralCatDruid_Rotation_Rip) ||
		rotation.FinishingMove == proto.FeralCatDruid_Rotation_Bite
	ripCP := rotation.RipMinComboPoints

	if rotation.FinishingMove != proto.FeralCatDruid_Rotation_Rip {
		ripCP = 6
	}

	cat.Rotation = FeralDruidRotation{
		RipCP:        ripCP,
		BiteCP:       rotation.BiteMinComboPoints,
		RipTrickCP:   rotation.RipMinComboPoints,
		UseBite:      useBite,
		BiteOverRip:  useBite && rotation.FinishingMove != proto.FeralCatDruid_Rotation_Rip,
		UseRipTrick:  rotation.Ripweave,
		UseRakeTrick: rotation.RakeTrick,
		Wolfshead:    cat.HasItemEquipped(8345, []proto.ItemSlot{proto.ItemSlot_ItemSlotHead}),

		MaintainFaerieFire: rotation.MaintainFaerieFire,
	}
}

func (cat *FeralDruid) OnGCDReady(_ *core.Simulation) {}

func (cat *FeralDruid) shift(sim *core.Simulation) bool {
	// Shifting Power spends mana for energy and stays in cat form. A cast that
	// cannot happen (cooldown, mana, missing talent, wrong form) is not a shift.
	if !cat.ShiftingPower.CanCast(sim, cat.CurrentTarget) {
		cat.readyToShift = false
		return false
	}

	cat.waitingForTick = false

	// If we have just now decided to shift, then we do not execute the
	// shift immediately, but instead trigger an input delay for realism.
	if !cat.readyToShift {
		cat.readyToShift = true
		return false
	}

	cat.readyToShift = false

	// Ready cooldowns fire first. Core refuses one that would leave cat form.
	// One that takes the GCD is left for the next decision.
	for _, mcd := range cat.GetMajorCooldowns() {
		if !mcd.IsReady(sim) {
			continue
		}
		mcd.TryActivate(sim, &cat.Character)
	}

	if !cat.GCD.IsReady(sim) {
		return true
	}
	return cat.ShiftingPower.Cast(sim, cat.CurrentTarget)
}

func (cat *FeralDruid) doRotation(sim *core.Simulation) bool {
	if !cat.GCD.IsReady(sim) {
		return false
	}

	// Shifting Power is the mana check that opens an OOM event. This rotation
	// bails out before casting, so drive the event pair explicitly to keep the
	// measured OOM time accurate. It has to sit above the early returns below:
	// an event opened here must be able to close on the same evaluation paths,
	// or a readyToShift tick or a Faerie Fire refresh leaves it open and
	// inflates the measured OOM time.
	shiftCost := cat.ShiftingPower.Cost.GetCurrentCost()
	oom := cat.CurrentMana() < shiftCost
	if oom {
		if !cat.IsOOM() {
			cat.StartOOMEvent(sim, shiftCost, false)
		}
	} else if cat.IsOOM() {
		cat.EndOOMEvent(sim, false)
	}

	// If we're out of form always shift back in.
	if !cat.InForm(druid.Cat) {
		return cat.CatForm.Cast(sim, nil)
	}

	// If we previously decided to shift, execute now after input delay.
	// A failed cast falls through and chooses another button this evaluation.
	if cat.readyToShift && cat.shift(sim) {
		return true
	}

	rotation := &cat.Rotation

	// Maintain Faerie Fire before other decisions.
	if rotation.MaintainFaerieFire && cat.FaerieFire != nil &&
		cat.FaerieFireAuras != nil &&
		!cat.FaerieFireAuras.Get(cat.CurrentTarget).IsActive() &&
		cat.FaerieFire.CanCast(sim, cat.CurrentTarget) {
		return cat.FaerieFire.Cast(sim, cat.CurrentTarget)
	}

	energy := cat.CurrentEnergy()
	cp := cat.ComboPoints()
	ripDot := cat.Rip.CurDot()
	ripDebuff := ripDot.IsActive()
	ripEnd := core.NeverExpires
	if ripDebuff {
		ripEnd = ripDot.ExpiresAt()
	}
	rakeDebuff := cat.Rake.CurDot().IsActive()
	nextTick := cat.NextEnergyTickAt()
	omenProc := cat.ClearcastingAura.IsActive()

	ripCost := cat.CurrentRipCost()
	biteCost := cat.CurrentFerociousBiteCost()
	shredCost := cat.CurrentShredCost()
	// TODO: The energy thresholds below assume a 45-energy builder, a fixed constant rather than a
	// cost read from the rows. Re-tune them with the Forever cat rotation.
	builderCost := 45.0

	ripNow := cp >= rotation.RipCP && !ripDebuff
	ripweaveNow := rotation.UseRipTrick &&
		cp >= rotation.RipTrickCP &&
		!ripDebuff &&
		energy >= RipTrickMin

	remainingDuration := sim.GetRemainingDuration()
	ripNow = (ripNow || ripweaveNow) && remainingDuration >= RipEndThresh

	biteAtEnd := cp >= rotation.BiteCP &&
		(remainingDuration < RipEndThresh ||
			(ripDebuff && sim.Duration-ripEnd < RipEndThresh))

	biteBeforeRip := ripDebuff && rotation.UseBite &&
		ripEnd-sim.CurrentTime >= BiteTime

	biteNow := (biteBeforeRip || rotation.BiteOverRip) &&
		cp >= rotation.BiteCP

	ripNext := (ripNow || (cp >= rotation.RipCP && ripEnd <= nextTick)) &&
		sim.Duration-nextTick >= RipEndThresh

	biteBeforeRipNext := biteBeforeRip && ripEnd-nextTick >= BiteTime

	timeToNextTick := nextTick - sim.CurrentTime
	cat.waitingForTick = true

	if oom {
		// No-shift rotation when OOM.
		if ripNow && (energy >= ripCost || omenProc) {
			return cat.Rip.Cast(sim, cat.CurrentTarget)
		} else if biteNow && (energy >= biteCost || omenProc) {
			return cat.FerociousBite.Cast(sim, cat.CurrentTarget)
		} else if energy >= shredCost || omenProc {
			return cat.Shred.Cast(sim, cat.CurrentTarget)
		}
	} else if energy < 10 {
		cat.shift(sim)
	} else if ripNow {
		if energy >= ripCost || omenProc {
			cat.Rip.Cast(sim, cat.CurrentTarget)
			cat.waitingForTick = false
		} else if timeToNextTick > MaxWaitTime {
			cat.shift(sim)
		}
	} else if biteNow || biteAtEnd {
		cutoffMod := 20.0
		if timeToNextTick <= time.Second {
			cutoffMod = 0.0
		}
		if energy >= shredCost+15.0+cutoffMod || (energy >= 15+cutoffMod && omenProc) {
			return cat.Shred.Cast(sim, cat.CurrentTarget)
		}
		if energy >= biteCost {
			return cat.FerociousBite.Cast(sim, cat.CurrentTarget)
		}
		wait := false
		if energy >= 22 && biteBeforeRip && !biteBeforeRipNext {
			wait = true
		} else if energy >= 15 && (!biteBeforeRip || biteBeforeRipNext || biteAtEnd) {
			wait = true
		} else if !ripNext {
			wait = false
			cat.shift(sim)
		} else {
			wait = true
		}
		if wait && timeToNextTick > MaxWaitTime {
			cat.shift(sim)
		}
	} else if energy >= biteCost && energy <= BiteTrickMax &&
		rotation.UseRakeTrick &&
		timeToNextTick > cat.ReactionTime &&
		!omenProc &&
		cp >= BiteTrickCP {
		return cat.FerociousBite.Cast(sim, cat.CurrentTarget)
	} else if energy >= biteCost && energy < builderCost &&
		rotation.UseRakeTrick &&
		timeToNextTick > time.Second+cat.ReactionTime &&
		!rakeDebuff &&
		!omenProc {
		return cat.Rake.Cast(sim, cat.CurrentTarget)
	} else if energy >= 22 {
		if omenProc {
			return cat.Shred.Cast(sim, cat.CurrentTarget)
		}
		if energy >= shredCost {
			return cat.Shred.Cast(sim, cat.CurrentTarget)
		}
		if timeToNextTick > MaxWaitTime {
			cat.shift(sim)
		}
	} else if !ripNext {
		cat.shift(sim)
	} else if timeToNextTick > MaxWaitTime {
		cat.shift(sim)
	}

	// Model input latency: delay GCD trigger after a power shift or energy tick.
	if cat.readyToShift {
		cat.WaitUntil(sim, sim.CurrentTime+cat.ReactionTime)
	} else if cat.waitingForTick {
		cat.WaitUntil(sim, sim.CurrentTime+timeToNextTick+cat.ReactionTime)
	}

	return false
}
