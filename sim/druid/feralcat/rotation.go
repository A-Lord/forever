package feralcat

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// Ported from https://github.com/NerdEgghead/TBC_cat_sim

const RipEndThresh = time.Second * 10

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

	return cat.probeRotation(sim)
}
