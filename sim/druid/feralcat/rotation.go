package feralcat

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/druid"
)

// A Rip cast with less of the fight left than this does not tick out.
const RipEndThresh = time.Second * 10

// Bite beside an active Rip only while that Rip still has longer than this left, so the combo
// points are not spent on a Rip that is about to fall off. Rip lasts 12s, so this is the first 4s.
const biteWithRipLeft = 8 * time.Second

type FeralDruidRotation struct {
	RipCP              int32
	BiteCP             int32
	UseBite            bool
	BiteOverRip        bool
	UseRakeTrick       bool
	MaintainFaerieFire bool
}

func (cat *FeralDruid) setupRotation(rotation *proto.FeralCatDruid_Rotation) {
	ripFinisher := rotation.FinishingMove == proto.FeralCatDruid_Rotation_Rip
	biteFinisher := rotation.FinishingMove == proto.FeralCatDruid_Rotation_Bite
	// A Rip needs more combo points than the bar holds when the rotation never casts it.
	ripCP := int32(6)
	if ripFinisher || (biteFinisher && rotation.Ripweave) {
		ripCP = rotation.RipMinComboPoints
	}

	cat.Rotation = FeralDruidRotation{
		RipCP:              ripCP,
		BiteCP:             rotation.BiteMinComboPoints,
		UseBite:            biteFinisher || (ripFinisher && rotation.Biteweave),
		BiteOverRip:        biteFinisher,
		UseRakeTrick:       rotation.RakeTrick,
		MaintainFaerieFire: rotation.MaintainFaerieFire,
	}
}

func (cat *FeralDruid) OnGCDReady(_ *core.Simulation) {}

// One decision of the cat rotation. Finishers first, then Rake while it is down, then the builder, and
// Shifting Power when the builder is unaffordable. Bite beside Rip only while Rip has more than
// 8s left. When nothing is castable it returns without waiting, so the APL looks again one reaction
// time later.
func (cat *FeralDruid) doRotation(sim *core.Simulation) bool {
	if !cat.GCD.IsReady(sim) {
		return false
	}

	// Shifting Power is the mana check that opens an OOM event. This rotation bails out before
	// casting, so it drives the event pair itself to keep the measured OOM time accurate.
	shiftCost := cat.ShiftingPower.Cost.GetCurrentCost()
	if cat.Talents.ShiftingPower && cat.CurrentMana() < shiftCost {
		if !cat.IsOOM() {
			cat.StartOOMEvent(sim, shiftCost, false)
		}
	} else if cat.IsOOM() {
		cat.EndOOMEvent(sim, false)
	}

	shiftDecided := cat.readyToShift
	cat.readyToShift = false

	if !cat.InForm(druid.Cat) {
		return cat.CatForm.Cast(sim, nil)
	}

	rotation := &cat.Rotation
	target := cat.CurrentTarget
	if rotation.MaintainFaerieFire && cat.FaerieFire.ShouldRefreshExclusiveEffects(sim, target, 0) && cat.FaerieFire.CanCast(sim, target) {
		return cat.FaerieFire.Cast(sim, target)
	}

	cp := cat.ComboPoints()
	ripDot := cat.Rip.CurDot()
	ripUp := ripDot.IsActive()
	if cp >= rotation.RipCP && !ripUp && sim.GetRemainingDuration() >= RipEndThresh &&
		cat.Rip.CanCast(sim, target) {
		return cat.Rip.Cast(sim, target)
	}
	// Bite as the finisher does not wait on Rip. Bite beside Rip waits until Rip has time left.
	biteWindow := rotation.BiteOverRip || (ripUp && ripDot.RemainingDuration(sim) > biteWithRipLeft)
	if rotation.UseBite && cp >= rotation.BiteCP && biteWindow && cat.FerociousBite.CanCast(sim, target) {
		return cat.FerociousBite.Cast(sim, target)
	}
	if rotation.UseRakeTrick && !cat.Rake.CurDot().IsActive() && cat.Rake.CanCast(sim, target) {
		return cat.Rake.Cast(sim, target)
	}

	builder := cat.Claw
	if cat.BehindTarget() {
		builder = cat.Shred
	}
	if cat.Consumables.GetConjuredId() == 7676 {
		if tea := cat.GetSpell(core.ActionID{ItemID: 7676}); tea != nil && cat.CurrentEnergy() <= 20 && tea.CanCast(sim, &cat.Unit) {
			tea.Cast(sim, &cat.Unit)
		}
	}
	if cat.wantShift(sim, target, builder) {
		if shiftDecided {
			return cat.ShiftingPower.Cast(sim, target)
		}
		// The next decision, one reaction time later, spends it.
		cat.readyToShift = true
		return false
	}
	if builder.CanCast(sim, target) {
		return builder.Cast(sim, target)
	}
	return false
}

// Shifting Power restores energy after the builder becomes unaffordable.
func (cat *FeralDruid) wantShift(sim *core.Simulation, target *core.Unit, builder *druid.DruidSpell) bool {
	if !cat.ShiftingPower.CanCast(sim, target) {
		return false
	}
	return cat.CurrentEnergy() < builder.Cost.GetCurrentCost() && !builder.CanCast(sim, target)
}
