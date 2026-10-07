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
	FaerieFireForOmen  bool
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
		FaerieFireForOmen:  rotation.FaerieFireForOmen,
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

	target := cat.CurrentTarget
	// The throw leaves Cat Form and needs no GCD, so Cat Form goes back up in the same decision.
	// It goes before every GCD of the rotation, and Berserk waits for it: on the pull the throw and
	// Cat Form's GCD come first, and Berserk opens on the GCD after.
	if explosive := cat.wantExplosive(sim, target); explosive != nil {
		explosive.Cast(sim, target)
		return cat.CatForm.Cast(sim, nil)
	}
	if cat.faerieFireNow(sim, target) {
		return cat.FaerieFire.Cast(sim, target)
	}
	// Rip is first when both bleeds are down. Refreshing either one early, to keep them apart,
	// was behind letting the fallen one wait a GCD.
	if cat.ripNow(sim, target) {
		return cat.Rip.Cast(sim, target)
	}
	if cat.biteNow(sim, target) {
		return cat.FerociousBite.Cast(sim, target)
	}
	if cat.rakeNow(sim, target) {
		return cat.Rake.Cast(sim, target)
	}

	builder := cat.builder()
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
	// Nothing else to cast: Faerie Fire is free in Cat Form and its landed debuff rolls Omen of Clarity.
	if cat.wantOmenFishing(sim, target) {
		return cat.FaerieFire.Cast(sim, target)
	}
	return false
}

// wantOmenFishing is true when the rotation fishes for Omen of Clarity, Omen can proc again and is not
// already up, and Faerie Fire is ready. Its GCD must not hold anything back: no spell the rotation
// wants is affordable before the GCD ends, and Shifting Power does not come off cooldown inside it.
func (cat *FeralDruid) wantOmenFishing(sim *core.Simulation, target *core.Unit) bool {
	omen := cat.OmenOfClarityAura
	if !cat.Rotation.FaerieFireForOmen || omen == nil || omen.Icd == nil || !omen.Icd.IsReady(sim) || cat.ClearcastingAura.IsActive() {
		return false
	}
	gcd := cat.FaerieFire.DefaultCast.GCD
	if gcd <= 0 {
		gcd = time.Second
	}
	regen := cat.CurrentEnergyRegenMultiplier() * 10 * gcd.Seconds()
	if cat.CurrentEnergy()+regen >= cat.cheapestWanted(sim, gcd) {
		return false
	}
	if cat.Talents.ShiftingPower && cat.ShiftingPower.TimeToReady(sim) < gcd {
		return false
	}
	return cat.FaerieFire.CanCast(sim, target)
}

// cheapestWanted is the lowest energy cost among the spells the rotation casts once it can pay, within
// the next window: Rip and Bite on the points they need, Rip and Rake when they are down or fall off
// inside the window, and the builder.
func (cat *FeralDruid) cheapestWanted(sim *core.Simulation, window time.Duration) float64 {
	rotation := &cat.Rotation
	cost := cat.builder().Cost.GetCurrentCost()
	cp := cat.ComboPoints()
	ripDot := cat.Rip.CurDot()
	ripUp := ripDot.IsActive()
	if cp >= rotation.RipCP && (!ripUp || ripDot.RemainingDuration(sim) < window) && sim.GetRemainingDuration() >= RipEndThresh {
		cost = min(cost, cat.Rip.Cost.GetCurrentCost())
	}
	if rotation.UseBite && cp >= rotation.BiteCP && (rotation.BiteOverRip || ripUp) {
		cost = min(cost, cat.FerociousBite.Cost.GetCurrentCost())
	}
	if rake := cat.Rake.CurDot(); rotation.UseRake && (!rake.IsActive() || rake.RemainingDuration(sim) < window) {
		cost = min(cost, cat.Rake.Cost.GetCurrentCost())
	}
	return cost
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
// Shifting Power does not spend Clearcasting, so with it up the next cast is free and the bar
// keeps filling through a second GCD: the ceiling leaves room for both, 40 and 20.
func (cat *FeralDruid) shiftCeiling() float64 {
	gcd := cat.ShiftingPower.DefaultCast.GCD
	if gcd <= 0 {
		gcd = time.Second
	}
	regen := cat.CurrentEnergyRegenMultiplier() * 10 * gcd.Seconds()
	if cat.ClearcastingAura.IsActive() {
		regen *= 2
	}
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

// wantExplosive is an explosive that cannot be used in a shapeshift form, ready to throw, with the
// mana to cast Cat Form again afterwards. Major cooldowns never leave the form on their own, so the
// rotation throws it. An APL that casts the explosive itself takes it out of the major cooldowns.
// 5/5 Furor brings back all the energy left in Cat Form; a lower rank keeps only part of it, so the
// throw then waits for the bar to be as low as a shift.
func (cat *FeralDruid) wantExplosive(sim *core.Simulation, target *core.Unit) *core.Spell {
	if cat.CatForm.Cost != nil && cat.CurrentMana() < cat.CatForm.Cost.GetCurrentCost() {
		return nil
	}
	if cat.Talents.Furor < 5 && cat.CurrentEnergy() >= cat.shiftCeiling() {
		return nil
	}
	for _, mcd := range cat.GetMajorCooldowns() {
		spell := mcd.Spell
		if spell.Flags.Matches(core.SpellFlagExplosive) && spell.CastRequirement.NotShapeshifted && spell.CanCast(sim, target) {
			return spell
		}
	}
	return nil
}

// builder is Shred from behind the target, Claw otherwise.
func (cat *FeralDruid) builder() *druid.DruidSpell {
	if cat.BehindTarget() {
		return cat.Shred
	}
	return cat.Claw
}

// The spells a decision casts ahead of the builder, in that order, each true when it is cast now.
func (cat *FeralDruid) faerieFireNow(sim *core.Simulation, target *core.Unit) bool {
	return cat.Rotation.MaintainFaerieFire && !cat.FaerieFireAuras.Get(target).IsActive() && cat.FaerieFire.CanCast(sim, target)
}

func (cat *FeralDruid) ripNow(sim *core.Simulation, target *core.Unit) bool {
	return cat.ComboPoints() >= cat.Rotation.RipCP && !cat.Rip.CurDot().IsActive() &&
		sim.GetRemainingDuration() >= RipEndThresh && cat.Rip.CanCast(sim, target)
}

func (cat *FeralDruid) biteNow(sim *core.Simulation, target *core.Unit) bool {
	rotation := &cat.Rotation
	biteWindow := rotation.BiteOverRip || cat.Rip.CurDot().IsActive()
	return rotation.UseBite && cat.ComboPoints() >= rotation.BiteCP && biteWindow && cat.FerociousBite.CanCast(sim, target)
}

func (cat *FeralDruid) rakeNow(sim *core.Simulation, target *core.Unit) bool {
	return cat.Rotation.UseRake && !cat.Rake.CurDot().IsActive() && cat.Rake.CanCast(sim, target)
}

// berserkReady holds Berserk until the GCD is free and the decision casts a spell its crit covers,
// Rake or the builder. Rip, Bite, Shifting Power, Faerie Fire and the Cat Form cast after a throw
// would each take Berserk's first GCD instead, and pooling for Rip casts nothing.
func (cat *FeralDruid) berserkReady(sim *core.Simulation) bool {
	target := cat.CurrentTarget
	if !cat.GCD.IsReady(sim) || !cat.InForm(druid.Cat) || cat.wantExplosive(sim, target) != nil {
		return false
	}
	if cat.faerieFireNow(sim, target) || cat.ripNow(sim, target) || cat.biteNow(sim, target) {
		return false
	}
	if cat.rakeNow(sim, target) {
		return true
	}
	builder := cat.builder()
	return !cat.pooling(sim, builder) && !cat.wantShift(sim, target) && builder.CanCast(sim, target)
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
