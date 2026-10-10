//go:build with_db

package feralcat

import (
	"github.com/wowsims/forever/sim/core/proto"
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

// Ravage 9867 and Pounce 9827 state No Active Defense: a target that dodges everything does not dodge them.
func TestRavageAndPounceCannotBeDodged(t *testing.T) {
	for _, opener := range []string{"Ravage", "Pounce"} {
		sim, cat := newCat(t, "", nil)
		target := cat.CurrentTarget
		target.PseudoStats.BaseDodgeChance = 1
		cat.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
		cat.ProwlAura.Activate(sim)
		spell := cat.Ravage
		if opener == "Pounce" {
			spell = cat.Pounce
		}
		if !spell.Cast(sim, target) {
			t.Fatalf("%s did not cast", opener)
		}
		if m := spell.SpellMetrics[target.UnitIndex]; m.Dodges != 0 || m.Hits+m.Crits != 1 {
			t.Errorf("%s against a target with 100%% dodge: %d dodges, %d hits and crits", opener, m.Dodges, m.Hits+m.Crits)
		}
	}
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

// Neither Rake's initial hit nor its ticks are reduced by armor in game.
func TestRakeIgnoresArmor(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	target := cat.CurrentTarget
	hit := cat.Rake.CalcDamage(sim, target, 100, cat.Rake.OutcomeAlwaysHit).Damage
	tick := cat.Rake.CalcPeriodicDamage(sim, target, 100, cat.Rake.OutcomeAlwaysHit).Damage
	target.AddStatDynamic(sim, stats.Armor, 10000)
	armoredHit := cat.Rake.CalcDamage(sim, target, 100, cat.Rake.OutcomeAlwaysHit).Damage
	armoredTick := cat.Rake.CalcPeriodicDamage(sim, target, 100, cat.Rake.OutcomeAlwaysHit).Damage
	if math.Abs(hit-armoredHit) > 1e-6 || math.Abs(tick-armoredTick) > 1e-6 {
		t.Errorf("Rake hit %g -> %g and tick %g -> %g after 10000 armor, want unchanged", hit, armoredHit, tick, armoredTick)
	}
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
	if n := castsOf(player.Actions, 22812); n != 0 {
		t.Fatalf("shipped preset casts Barkskin %d times", n)
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

// Rend and Tear raises Shred on a bleeding target. With the bleed assumed, a target with no bleed of
// the cat's own takes the same raised Shred.
func TestAssumedBleedRaisesShredLikeOwnBleed(t *testing.T) {
	const rendAndTear5 = "0000000000000000-00000000000000000050-0000000000000000"
	shred := func(assume, rake bool) float64 {
		sim, cat := newCat(t, rendAndTear5, nil)
		cat.AssumeBleedActive = assume
		target := cat.CurrentTarget
		cat.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
		cat.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
		target.PseudoStats.BaseDodgeChance = 0
		if rake {
			cat.Rake.Dot(target).Apply(sim)
		}
		fillResource(sim, cat)
		if !cat.Shred.Cast(sim, target) {
			t.Fatal("shred did not cast")
		}
		return cat.Shred.SpellMetrics[target.UnitIndex].TotalDamage
	}
	plain, bleeding, assumed := shred(false, false), shred(false, true), shred(true, false)
	if plain <= 0 || bleeding <= plain*1.01 {
		t.Fatalf("shred %g without a bleed, %g with Rake up: Rend and Tear did not apply", plain, bleeding)
	}
	if math.Abs(assumed-bleeding) > 1e-6 {
		t.Fatalf("shred with the bleed assumed = %g, with Rake up = %g", assumed, bleeding)
	}
}

// Rend and Tear raises bleed ticks. A Rake or Rip tick counts its own bleed as the bleed on the target,
// and the talent is the only difference between the two builds.
func TestRendAndTearRaisesBleedTicks(t *testing.T) {
	const rendAndTear5 = "0000000000000000-00000000000000000050-0000000000000000"
	tick := func(talents string, rip bool) float64 {
		sim, cat := newCat(t, talents, nil)
		cat.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
		spell := cat.Rake
		if rip {
			cat.AddComboPoints(sim, 5, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
			spell = cat.Rip
		}
		dot := spell.Dot(cat.CurrentTarget)
		dot.Apply(sim)
		metric := &spell.SpellMetrics[cat.CurrentTarget.UnitIndex]
		before := metric.TotalDamage
		dot.TickOnce(sim)
		return metric.TotalDamage - before
	}
	for _, rip := range []bool{false, true} {
		plain, raised := tick("", rip), tick(rendAndTear5, rip)
		if plain <= 0 || math.Abs(raised/plain-1.10) > 1e-6 {
			t.Errorf("rip=%v: tick %g without the talent, %g with 5/5, want x1.10", rip, plain, raised)
		}
	}
}

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

func TestRakeUsesAttackPowerForTicksOnly(t *testing.T) {
	damage := func(attackPowerBeforeCast float64, attackPowerAfterCast float64) (float64, float64) {
		sim, cat := newCat(t, "", nil)
		prepareFeralDamageCase(sim, cat)
		cat.AddStatDynamic(sim, stats.AttackPower, attackPowerBeforeCast)
		if !cat.Rake.Cast(sim, cat.CurrentTarget) {
			t.Fatal("Rake did not cast")
		}
		metric := &cat.Rake.SpellMetrics[cat.CurrentTarget.UnitIndex]
		direct := metric.TotalDamage
		cat.AddStatDynamic(sim, stats.AttackPower, attackPowerAfterCast)
		cat.Rake.Dot(cat.CurrentTarget).TickOnce(sim)
		return direct, metric.TotalDamage - direct
	}

	baseDirect, baseTick := damage(0, 0)
	preCastDirect, preCastTick := damage(1000, 0)
	postCastDirect, postCastTick := damage(0, 1000)

	if preCastDirect != baseDirect || postCastDirect != baseDirect {
		t.Errorf("Rake direct damage with attack power before/after cast = %v/%v, want unchanged %v", preCastDirect, postCastDirect, baseDirect)
	}
	// Rank 1 observations fit 0.055 attack power per tick. Higher ranks keep that coefficient as an
	// assumption, and whether Rake should snapshot attack power remains open.
	for _, result := range []struct {
		timing string
		tick   float64
	}{
		{"before cast", preCastTick},
		{"after application", postCastTick},
	} {
		if got, want := result.tick-baseTick, 55.0; math.Abs(got-want) > 1e-9 {
			t.Errorf("Rake tick gained %v damage from 1000 attack power added %s, want %v", got, result.timing, want)
		}
	}
}

func TestPounceTalentsModifyOnlyTheirOwnEffects(t *testing.T) {
	const brutalImpactTalents = "0000000000000000-00002000000000000000-0000000000000000"

	castPounce := func(talents string) (float64, time.Duration, time.Duration, time.Duration) {
		sim, cat := newCat(t, talents, nil)
		if talents == brutalImpactTalents && cat.Talents.BrutalImpact != 2 {
			t.Fatalf("Brutal Impact rank = %d, want 2", cat.Talents.BrutalImpact)
		}
		prepareFeralDamageCase(sim, cat)
		sim.CurrentTime = -time.Second
		if !cat.Prowl.Cast(sim, cat.CurrentTarget) {
			t.Fatal("Prowl did not cast before the pull")
		}
		sim.CurrentTime = 0
		fillResource(sim, cat)
		if !cat.Pounce.Cast(sim, cat.CurrentTarget) {
			t.Fatal("Pounce did not cast from Prowl")
		}
		bleed := cat.PounceBleed.Dot(cat.CurrentTarget)
		bleed.TickOnce(sim)
		stun := cat.CurrentTarget.GetAura("Pounce Stun")
		if stun == nil || !stun.IsActive() {
			t.Fatal("Pounce stun did not activate")
		}
		return cat.PounceBleed.SpellMetrics[cat.CurrentTarget.UnitIndex].TotalDamage,
			stun.RemainingDuration(sim), bleed.RemainingDuration(sim), bleed.BaseDurationFlat
	}

	baseDamage, baseStun, baseBleed, _ := castPounce("")
	genesisDamage, _, _, _ := castPounce("0500000000000000-00000000000000000000-0000000000000000")
	_, brutalStun, brutalBleed, brutalBleedExtension := castPounce(brutalImpactTalents)

	if got, want := genesisDamage/baseDamage, 1.05; math.Abs(got-want) > 1e-9 {
		t.Errorf("Genesis Pounce bleed multiplier = %v, want %v", got, want)
	}
	if got, want := brutalStun-baseStun, time.Second; got != want {
		t.Errorf("Brutal Impact Pounce stun extension = %v, want %v", got, want)
	}
	if brutalBleed != baseBleed {
		t.Errorf("Pounce bleed duration with Brutal Impact = %v, want unchanged %v", brutalBleed, baseBleed)
	}
	if brutalBleedExtension != 0 {
		t.Errorf("Pounce bleed duration extension with Brutal Impact = %v, want 0", brutalBleedExtension)
	}
}

func TestRendAndTearReadsBleedsPerTarget(t *testing.T) {
	biteDamage := func(targetIndex int) float64 {
		sim, cat := newCatWearing(t, "0000000000000000-00000000000000000050-0000000000000000", nil, nil, &proto.Target{Name: "target", Level: 60}, &proto.Target{Name: "second target", Level: 60})
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
