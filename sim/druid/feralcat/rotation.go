package feralcat

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/druid"
)

// A Rip cast with less of the fight left than this does not tick out.
const RipEndThresh = time.Second * 10

// With 5 combo points and Rip inside this window, a builder that would leave Rip unpaid waits.
const ripPoolWindow = 2 * time.Second

// Thistle Tea restores 100 energy, and it never fires on its own: the bar refills continuously,
// so there is never room for the whole 100. Drinking at 20 lands it once a fight.
const teaEnergy = 20
const thistleTeaID int32 = 7676

type FeralDruidRotation struct {
	RipCP              int32
	BiteCP             int32
	UseBite            bool
	BiteOverRip        bool
	UseRake            bool
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
		UseRake:            rotation.RakeTrick,
		MaintainFaerieFire: rotation.MaintainFaerieFire,
	}
}

func (cat *FeralDruid) OnGCDReady(_ *core.Simulation) {}

// One decision of the cat rotation. Finishers first, then Rake while it is down, then the builder.
// Shifting Power is spent whenever the gain and the energy regenerated during its GCD still fit.
// Bite-weaving bites whenever Rip is up. A bleed is refreshed when it has fallen, not before.
// When nothing is castable it returns without waiting, so the APL looks again one reaction time later.
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
	if rotation.MaintainFaerieFire && !cat.FaerieFireAuras.Get(target).IsActive() && cat.FaerieFire.CanCast(sim, target) {
		return cat.FaerieFire.Cast(sim, target)
	}

	cp := cat.ComboPoints()
	ripDot := cat.Rip.CurDot()
	ripUp := ripDot.IsActive()
	// Rip is first when both bleeds are down. Refreshing either one early, to keep them apart,
	// was behind letting the fallen one wait a GCD.
	if cp >= rotation.RipCP && !ripUp && sim.GetRemainingDuration() >= RipEndThresh &&
		cat.Rip.CanCast(sim, target) {
		return cat.Rip.Cast(sim, target)
	}
	biteWindow := rotation.BiteOverRip || ripUp
	if rotation.UseBite && cp >= rotation.BiteCP && biteWindow && cat.FerociousBite.CanCast(sim, target) {
		return cat.FerociousBite.Cast(sim, target)
	}
	if rotation.UseRake && !cat.Rake.CurDot().IsActive() && cat.Rake.CanCast(sim, target) {
		return cat.Rake.Cast(sim, target)
	}

	builder := cat.Claw
	if cat.BehindTarget() {
		builder = cat.Shred
	}
	// Tea does not start the GCD, so the same decision can still spend the builder or the shift.
	if cat.wantTea(sim) {
		cat.thistleTea().Cast(sim, &cat.Unit)
	}
	if cat.pooling(sim, builder) {
		if cat.wantShift(sim, target) {
			return cat.castShift(sim, target, shiftDecided)
		}
		return false
	}
	if cat.wantShift(sim, target) {
		return cat.castShift(sim, target, shiftDecided)
	}
	if builder.CanCast(sim, target) {
		return builder.Cast(sim, target)
	}
	return false
}

func (cat *FeralDruid) castShift(sim *core.Simulation, target *core.Unit, alreadyDecided bool) bool {
	if alreadyDecided {
		return cat.ShiftingPower.Cast(sim, target)
	}
	// The next decision, one reaction time later, spends it.
	cat.readyToShift = true
	return false
}

// shiftingPowerEnergy is the row's 40, plus Wolfshead when that helm is worn.
func (cat *FeralDruid) shiftingPowerEnergy() float64 {
	return 40 + cat.WolfsheadEnergyBonus
}

// shiftCeiling is the energy at which the gain plus regen over the GCD still fits under 100.
// 50 for the row's 40. Wolfshead's extra 20 lowers it to 30. The bar regens 10 per second.
func (cat *FeralDruid) shiftCeiling() float64 {
	gcd := cat.ShiftingPower.DefaultCast.GCD
	if gcd <= 0 {
		gcd = time.Second
	}
	regen := cat.CurrentEnergyRegenMultiplier() * 10 * gcd.Seconds()
	return cat.MaximumEnergy() - cat.shiftingPowerEnergy() - regen
}

// wantShift spends Shifting Power once it is ready and the energy is under shiftCeiling, including
// while the builder could still be paid. Waiting for an empty bar does not bring the cooldown back
// any sooner, and casting with no regard for the cap wastes the gain.
func (cat *FeralDruid) wantShift(sim *core.Simulation, target *core.Unit) bool {
	if !cat.ShiftingPower.CanCast(sim, target) {
		return false
	}
	return cat.CurrentEnergy() < cat.shiftCeiling()
}

// pooling keeps a Rip's energy together when Rip will fall inside the window and the points are ready.
// Spending the builder would leave the refresh short, so the decision waits or shifts instead.
func (cat *FeralDruid) pooling(sim *core.Simulation, builder *druid.DruidSpell) bool {
	if cat.ComboPoints() < cat.Rotation.RipCP {
		return false
	}
	ripDot := cat.Rip.CurDot()
	if !ripDot.IsActive() || sim.GetRemainingDuration() < RipEndThresh {
		return false
	}
	left := ripDot.RemainingDuration(sim)
	if left <= 0 || left > ripPoolWindow {
		return false
	}
	gcd := builder.DefaultCast.GCD
	if gcd <= 0 {
		gcd = time.Second
	}
	afterBuilder := cat.CurrentEnergy() - builder.Cost.GetCurrentCost() + cat.CurrentEnergyRegenMultiplier()*10*gcd.Seconds()
	return afterBuilder < cat.Rip.Cost.GetCurrentCost()
}

func (cat *FeralDruid) thistleTea() *druid.DruidSpell {
	spell := cat.GetSpell(core.ActionID{ItemID: 7676})
	if spell == nil {
		return nil
	}
	return &druid.DruidSpell{Spell: spell}
}

func (cat *FeralDruid) wantTea(sim *core.Simulation) bool {
	tea := cat.thistleTea()
	if tea == nil || cat.CurrentEnergy() > teaEnergy {
		return false
	}
	return tea.CanCast(sim, &cat.Unit)
}
