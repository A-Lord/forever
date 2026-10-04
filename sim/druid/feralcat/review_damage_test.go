//go:build with_db

package feralcat

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/stats"
)

func TestBiteDodgeConsumesEnergyAndKeepsComboPoints(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	target := cat.CurrentTarget
	target.PseudoStats.BaseDodgeChance = 1
	cat.AddComboPoints(sim, 5, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
	if !cat.FerociousBite.Cast(sim, target) {
		t.Fatal("bite did not cast")
	}
	if cat.FerociousBite.SpellMetrics[target.UnitIndex].Dodges != 1 {
		t.Fatal("bite was not dodged")
	}
	if cat.CurrentEnergy() != 0 || cat.ComboPoints() != 5 {
		t.Fatalf("dodged bite leaves energy=%g CP=%d, want 0 and 5", cat.CurrentEnergy(), cat.ComboPoints())
	}
	t.Logf("dodged bite: energy=%g CP=%d", cat.CurrentEnergy(), cat.ComboPoints())
}

func TestRakeTickScalesWithAttackPower(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	cat.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
	dot := cat.Rake.Dot(cat.CurrentTarget)
	dot.Apply(sim)
	metric := &cat.Rake.SpellMetrics[cat.CurrentTarget.UnitIndex]
	before := metric.TotalDamage
	dot.TickOnce(sim)
	low := metric.TotalDamage - before
	cat.AddStatDynamic(sim, stats.AttackPower, 100)
	before = metric.TotalDamage
	dot.TickOnce(sim)
	high := metric.TotalDamage - before
	if math.Abs((high-low)-5.5) > 1e-5 {
		t.Fatalf("100 AP tick gain=%g, want 5.5", high-low)
	}
	t.Logf("Rake 100 AP gain per tick=%g", high-low)
}

func TestRakeInitialDamageIgnoresArmor(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	target := cat.CurrentTarget
	hit := cat.Rake.CalcDamage(sim, target, 100, cat.Rake.OutcomeAlwaysHit).Damage
	target.AddStatDynamic(sim, stats.Armor, 10000)
	armored := cat.Rake.CalcDamage(sim, target, 100, cat.Rake.OutcomeAlwaysHit).Damage
	if math.Abs(hit-armored) > 1e-6 {
		t.Fatalf("Rake initial hit %g -> %g after armor", hit, armored)
	}
	t.Logf("Rake initial hit before/after 10000 armor: %g/%g", hit, armored)
}

func TestFurorLeavingAtZeroTimeWithZeroEnergy(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	cat.SpendEnergy(sim, cat.CurrentEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	if !cat.DireBearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("bear cast failed")
	}
	sim.CurrentTime = 3 * time.Second
	fillResource(sim, cat)
	if !cat.CatForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("cat cast failed")
	}
	if cat.CurrentEnergy() != 0 {
		t.Fatalf("zero energy stored at t=0 restored %g", cat.CurrentEnergy())
	}
}

func TestRipAttackPowerCapsAtFourComboPoints(t *testing.T) {
	for cp := int32(1); cp <= 5; cp++ {
		sim, cat := newCat(t, "", nil)
		cat.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
		cat.AddComboPoints(sim, cp, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
		dot := cat.Rip.Dot(cat.CurrentTarget)
		dot.Apply(sim)
		metric := &cat.Rip.SpellMetrics[cat.CurrentTarget.UnitIndex]
		before := metric.TotalDamage
		dot.TickOnce(sim)
		low := metric.TotalDamage - before
		cat.AddStatDynamic(sim, stats.AttackPower, 100)
		before = metric.TotalDamage
		dot.TickOnce(sim)
		gain := metric.TotalDamage - before - low
		if math.Abs(gain-float64(min(cp, 4))) > 1e-5 {
			t.Fatalf("cp=%d tick gain=%g", cp, gain)
		}
		t.Logf("Rip %d CP: 100 AP gains %g per tick / %g total", cp, gain, gain*6)
	}
}

func TestCatShippedPresetRuns(t *testing.T) {
	req := catPowershiftRequest(powershiftTalents, core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "default").Rotation, false)
	req.Raid.Parties[0].Players[0].Equipment = core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "default").GearSet
	player := runCat(t, req)
	for _, id := range []int32{9830, 9904, 9896, 1322605} {
		if castsOf(player.Actions, id) == 0 {
			t.Fatalf("shipped preset never casts %d", id)
		}
	}
}

func TestCatStealthPresetsRun(t *testing.T) {
	for name, opener := range map[string]int32{"ravage": 9867, "pounce": 9827} {
		t.Run(name, func(t *testing.T) {
			req := catPowershiftRequest(powershiftTalents, core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", name).Rotation, false)
			req.Raid.Parties[0].Players[0].Equipment = core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "default").GearSet
			player := runCat(t, req)
			if castsOf(player.Actions, opener) != 1 || castsOf(player.Actions, 9830) == 0 {
				t.Fatalf("%s must open once then continue with Shred", name)
			}
		})
	}
}
