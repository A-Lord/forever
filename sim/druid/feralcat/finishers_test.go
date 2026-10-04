//go:build with_db

package feralcat

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/stats"
)

func prepareFeralDamageCase(sim *core.Simulation, cat *FeralDruid) {
	target := cat.CurrentTarget
	target.AddStatDynamic(sim, stats.Armor, -target.GetStat(stats.Armor))
	cat.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
	cat.AddStatDynamic(sim, stats.ExpertisePercent, 100)
	cat.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
	fillResource(sim, cat)
}

func TestRipSnapshotsComboPointsOneThroughFive(t *testing.T) {
	for comboPoints := int32(1); comboPoints <= 5; comboPoints++ {
		t.Run(string(rune('0'+comboPoints)), func(t *testing.T) {
			// Given a cat with a fixed number of combo points.
			sim, cat := newCat(t, "", nil)
			prepareFeralDamageCase(sim, cat)
			cat.AddComboPoints(sim, comboPoints, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))

			// When Rip is applied.
			dot := cat.Rip.Dot(cat.CurrentTarget)
			dot.Apply(sim)

			// Then the row's 15 base and 25.5 per point are captured at application.
			want := 15 + 25.5*float64(comboPoints)
			if dot.SnapshotBaseDamage != want {
				t.Errorf("%d-point Rip snapshot = %v, want %v", comboPoints, dot.SnapshotBaseDamage, want)
			}
		})
	}
}

func TestRipReadsAttackPowerAtEachTickAndCapsItsPlaceholderAtFourPoints(t *testing.T) {
	tickDamage := func(attackPower float64) float64 {
		sim, cat := newCat(t, "", nil)
		prepareFeralDamageCase(sim, cat)
		cat.AddComboPoints(sim, 5, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
		dot := cat.Rip.Dot(cat.CurrentTarget)
		dot.Apply(sim)
		cat.AddStatDynamic(sim, stats.AttackPower, attackPower)
		dot.TickOnce(sim)
		return cat.Rip.SpellMetrics[cat.CurrentTarget.UnitIndex].TotalDamage
	}

	// Given two identical five-point Rips, with attack power added after application.
	withoutAP := tickDamage(0)
	withAP := tickDamage(100)

	// When each Rip ticks, then the assumed 0.01 coefficient uses current attack power and four points.
	if got, want := withAP-withoutAP, 4.0; math.Abs(got-want) > 1e-9 {
		t.Errorf("five-point Rip gained %v tick damage from 100 attack power, want %v", got, want)
	}
}

func TestRakeHasNoAttackPowerScalingWithoutAGameFilesCoefficient(t *testing.T) {
	totalDamage := func(attackPower float64) float64 {
		sim, cat := newCat(t, "", nil)
		prepareFeralDamageCase(sim, cat)
		cat.AddStatDynamic(sim, stats.AttackPower, attackPower)
		if !cat.Rake.Cast(sim, cat.CurrentTarget) {
			t.Fatal("Rake did not cast")
		}
		cat.Rake.Dot(cat.CurrentTarget).TickOnce(sim)
		return cat.Rake.SpellMetrics[cat.CurrentTarget.UnitIndex].TotalDamage
	}

	// Given equal Rake casts whose only difference is 1000 attack power.
	withoutAP := totalDamage(0)
	withAP := totalDamage(1000)

	// When the direct hit and one tick land, then neither invents a coefficient absent from both rows.
	if withAP != withoutAP {
		t.Errorf("Rake damage with 1000 attack power = %v, want unchanged %v", withAP, withoutAP)
	}
}

func TestGenesisReachesTheTriggeredPounceBleed(t *testing.T) {
	tickDamage := func(talents string) float64 {
		sim, cat := newCat(t, talents, nil)
		prepareFeralDamageCase(sim, cat)
		dot := cat.PounceBleed.Dot(cat.CurrentTarget)
		dot.Apply(sim)
		dot.TickOnce(sim)
		return cat.PounceBleed.SpellMetrics[cat.CurrentTarget.UnitIndex].TotalDamage
	}

	// Given Pounce's triggered bleed with and without Genesis 5/5.
	base := tickDamage("")
	genesis := tickDamage("0500000000000000-00000000000000000000-0000000000000000")

	// When the bleed ticks, then the talent's parent-Pounce mask raises it by 5%.
	if got, want := genesis/base, 1.05; math.Abs(got-want) > 1e-9 {
		t.Errorf("Genesis Pounce bleed multiplier = %v, want %v", got, want)
	}
}

func TestRendAndTearReadsBleedsPerTarget(t *testing.T) {
	biteDamage := func(targetIndex int) float64 {
		sim, cat := newCatWearingTargets(t, "0000000000000000-00000000000000000050-0000000000000000", nil, nil, 2)
		target := sim.Environment.GetTargetUnitByIndex(int32(targetIndex))
		target.AddStatDynamic(sim, stats.Armor, -target.GetStat(stats.Armor))
		cat.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
		cat.AddStatDynamic(sim, stats.ExpertisePercent, 100)
		cat.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
		fillResource(sim, cat)
		cat.Rake.Dot(sim.Environment.GetTargetUnitByIndex(0)).Apply(sim)
		cat.AddComboPoints(sim, 1, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
		if !cat.FerociousBite.Cast(sim, target) {
			t.Fatal("Ferocious Bite did not cast")
		}
		return cat.FerociousBite.SpellMetrics[target.UnitIndex].TotalDamage
	}

	// Given two targets where only the first has a tagged Druid bleed.
	bleeding := biteDamage(0)
	unbled := biteDamage(1)

	// When Rend and Tear prices Bite, then its 10% bonus applies only to the bleeding target.
	if got, want := bleeding/unbled, 1.1; math.Abs(got-want) > 1e-9 {
		t.Errorf("Rend and Tear multiplier = %v, want %v", got, want)
	}
}

func TestFerociousBiteScalesWithComboPointsAndSpendsAllExtraEnergy(t *testing.T) {
	biteDamage := func(comboPoints int32, attackPower float64, clearcasting bool) (float64, float64, bool) {
		sim, cat := newCat(t, "", nil)
		prepareFeralDamageCase(sim, cat)
		cat.AddComboPoints(sim, comboPoints, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
		cat.AddStatDynamic(sim, stats.AttackPower, attackPower)
		if clearcasting {
			cat.ClearcastingAura.Activate(sim)
			cat.ClearcastingAura.SetStacks(sim, cat.ClearcastingAura.MaxStacks)
		}
		if !cat.FerociousBite.Cast(sim, cat.CurrentTarget) {
			t.Fatal("Ferocious Bite did not cast")
		}
		energy := cat.CurrentEnergy()
		for sim.CurrentTime <= core.SpellBatchWindow && !sim.Step() {
		}
		return cat.FerociousBite.SpellMetrics[cat.CurrentTarget.UnitIndex].TotalDamage,
			energy, cat.ClearcastingAura.IsActive()
	}

	for comboPoints := int32(1); comboPoints <= 5; comboPoints++ {
		// Given equal Bites whose only damage input difference is 100 attack power.
		withoutAP, energy, _ := biteDamage(comboPoints, 0, false)
		withAP, _, _ := biteDamage(comboPoints, 100, false)

		// When they land, then the assumed 0.03 coefficient scales once per combo point.
		if got, want := withAP-withoutAP, 3*float64(comboPoints); math.Abs(got-want) > 1e-9 {
			t.Errorf("%d-point Bite gained %v damage from 100 attack power, want %v", comboPoints, got, want)
		}
		if energy != 0 {
			t.Errorf("%d-point Bite left %v energy, want 0", comboPoints, energy)
		}
	}

	// Given Clearcasting at full energy, the base cost is free and all 100 energy remains extra.
	regular, _, _ := biteDamage(5, 0, false)
	clear, energy, active := biteDamage(5, 0, true)

	// When Bite lands, then the extra 35 energy adds 87.5 damage and the cast consumes the charge.
	if got, want := clear-regular, 87.5; math.Abs(got-want) > 1e-9 {
		t.Errorf("Clearcasting Bite bonus = %v, want %v", got, want)
	}
	if energy != 0 || active {
		t.Errorf("Clearcasting Bite left energy=%v active=%v, want 0 and false", energy, active)
	}
}
