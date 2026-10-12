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

func TestFeralPhysicalDamageMatchesPawDamage(t *testing.T) {
	for _, name := range []string{"Auto", "Shred", "Ravage", "Claw", "Maul", "Primal Bite", "Lacerate", "Rake", "Bite"} {
		t.Run(name, func(t *testing.T) {
			damage := func(mode string) (float64, float64) {
				sim, cat := newCat(t, "050022-55000032021131212051-053", nil)
				if name == "Maul" || name == "Primal Bite" || name == "Lacerate" {
					cat.ClearForm(sim)
					cat.DireBearFormAura.Activate(sim)
				}
				cat.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
				cat.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
				cat.AddStatDynamic(sim, stats.ExpertisePercent, 100)
				cat.AddComboPoints(sim, 5, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
				const bonus = 7.0
				if mode == "stat" {
					cat.AddStatDynamic(sim, stats.PhysicalDamage, bonus)
				} else if mode == "paw" {
					cat.AutoAttacks.MH().BaseDamageMin += bonus
					cat.AutoAttacks.MH().BaseDamageMax += bonus
				}
				spell := cat.MHAutoSpell
				switch name {
				case "Shred":
					spell = cat.Shred.Spell
				case "Ravage":
					spell = cat.Ravage.Spell
				case "Claw":
					spell = cat.Claw.Spell
				case "Maul":
					spell = cat.GetSpell(cat.Maul.ActionID.WithTag(1))
				case "Primal Bite":
					spell = cat.PrimalBite.Spell
				case "Lacerate":
					spell = cat.Lacerate.Spell
					dot := spell.Dot(cat.CurrentTarget)
					dot.Apply(sim)
					dot.SetStacks(sim, 2)
				case "Rake":
					spell = cat.Rake.Spell
				case "Bite":
					spell = cat.FerociousBite.Spell
				}
				expected := 0.0
				if name == "Auto" {
					expected = spell.ExpectedInitialDamage(sim, cat.CurrentTarget)
				}
				spell.ApplyEffects(sim, cat.CurrentTarget, spell)
				return spell.SpellMetrics[cat.CurrentTarget.UnitIndex].TotalDamage, expected
			}
			plain, _ := damage("base")
			stat, statExpected := damage("stat")
			paw, pawExpected := damage("paw")
			if math.Abs(stat-paw) > 1e-8 || math.Abs(statExpected-pawExpected) > 1e-8 {
				t.Fatalf("flat physical damage gives %g (expected %g), paw damage gives %g (expected %g)", stat, statExpected, paw, pawExpected)
			}
			if name == "Rake" || name == "Bite" {
				if math.Abs(stat-plain) > 1e-8 {
					t.Fatalf("non-weapon hit gained %g damage", stat-plain)
				}
			} else if stat <= plain {
				t.Fatalf("weapon damage did not increase: %g -> %g", plain, stat)
			}
		})
	}
}

func TestFeralFlatDamageFollowsBuffsFormsAndReset(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	const bonus = 9.0
	check := func(inPaw bool) {
		t.Helper()
		// Remove weapon variance so this measures only the flat damage path.
		cat.AutoAttacks.MH().BaseDamageMin = 0
		cat.AutoAttacks.MH().BaseDamageMax = 0
		wantWeapon, wantSpell := 0.0, 0.0
		if inPaw {
			wantWeapon = cat.GetStat(stats.PhysicalDamage)
		} else {
			wantSpell = cat.GetStat(stats.PhysicalDamage)
		}
		if got := cat.MHWeaponDamage(sim, 0); math.Abs(got-wantWeapon) > 1e-8 {
			t.Fatalf("weapon flat damage = %g, want %g", got, wantWeapon)
		}
		if got := cat.MHNormalizedWeaponDamage(sim, 0); math.Abs(got-wantWeapon) > 1e-8 {
			t.Fatalf("normalized weapon flat damage = %g, want %g", got, wantWeapon)
		}
		if got := cat.MHAutoSpell.BonusDamage(cat.AttackTables[cat.CurrentTarget.UnitIndex]); math.Abs(got-wantSpell) > 1e-8 {
			t.Fatalf("spell flat damage = %g, want %g", got, wantSpell)
		}
	}
	cat.AddStatDynamic(sim, stats.PhysicalDamage, bonus)
	check(true)
	for _, aura := range []*core.Aura{cat.BearFormAura, cat.DireBearFormAura, cat.CatFormAura} {
		cat.ClearForm(sim)
		check(false)
		aura.Activate(sim)
		check(true)
	}
	cat.AddStatDynamic(sim, stats.PhysicalDamage, -bonus)
	check(true)
	cat.AddStatDynamic(sim, stats.PhysicalDamage, bonus)
	sim.Cleanup()
	sim.Reset()
	if got := cat.GetStat(stats.PhysicalDamage); got != 0 {
		t.Fatalf("temporary flat damage survives reset: %g", got)
	}
	check(true)
}
