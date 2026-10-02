//go:build with_db

package feralbear

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/druid"
)

func newBear(t *testing.T) (*core.Simulation, *GuardianDruid) {
	return newBearWithTalents(t, "0000000000000000-00000000000000000000-0000000000000000")
}

func newBearWithTalents(t *testing.T, talents string) (*core.Simulation, *GuardianDruid) {
	t.Helper()
	return newBearAgainst(t, talents, 1)
}

func newBearAgainst(t *testing.T, talents string, numTargets int) (*core.Simulation, *GuardianDruid) {
	t.Helper()
	targets := make([]*proto.Target, numTargets)
	for i := range targets {
		targets[i] = &proto.Target{Name: "target", Level: 60}
	}
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 866
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Bear", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
				TalentsString: talents,
				Spec: &proto.Player_FeralBearDruid{FeralBearDruid: &proto.FeralBearDruid{
					Options: &proto.FeralBearDruid_Options{StartingRage: 100},
				}},
				Equipment:   &proto.EquipmentSpec{Items: items},
				Consumables: &proto.ConsumesSpec{},
				Buffs:       &proto.IndividualBuffs{},
				Rotation:    &proto.APLRotation{},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: targets, Duration: 60},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*GuardianDruid)
}

func TestDemoralizingRoarReachesEveryTarget(t *testing.T) {
	sim, bear := newBearAgainst(t, "0000000000000000-00000000000000000000-0000000000000000", 3)
	targets := sim.Encounter.ActiveTargetUnits
	for range 20 {
		bear.AddRage(sim, 50, bear.NewRageMetrics(core.ActionID{SpellID: 1}))
		bear.DemoralizingRoar.Cast(sim, bear.CurrentTarget)
		sim.CurrentTime += 2 * time.Second
		debuffed := 0
		for _, target := range targets {
			if bear.DemoralizingRoarAuras.Get(target).IsActive() {
				debuffed++
			}
		}
		if debuffed == len(targets) {
			return
		}
	}
	t.Error("demoralizing roar never debuffed every target")
}

func TestDemoralizingRoarIgnoresDodge(t *testing.T) {
	sim, bear := newBear(t)
	target := sim.Encounter.ActiveTargetUnits[0]
	target.PseudoStats.BaseDodgeChance = 1
	for range 20 {
		bear.AddRage(sim, 50, bear.NewRageMetrics(core.ActionID{SpellID: 1}))
		bear.DemoralizingRoar.Cast(sim, bear.CurrentTarget)
		sim.CurrentTime += 2 * time.Second
		if m := bear.DemoralizingRoar.SpellMetrics[target.UnitIndex]; m.Dodges != 0 {
			t.Fatalf("demoralizing roar was dodged %d times", m.Dodges)
		}
		if bear.DemoralizingRoarAuras.Get(target).IsActive() {
			return
		}
	}
	t.Error("demoralizing roar never landed against a target that dodges every melee attack")
}

func TestBerserkPrimalBiteStrikesThreeTargets(t *testing.T) {
	sim, bear := newBearAgainst(t, "0000000000000000-00000000000100000001-0000000000000000", 5)
	metrics := bear.NewRageMetrics(core.ActionID{SpellID: 1})
	bear.BerserkAura.Activate(sim)
	bear.AddRage(sim, 100, metrics)
	if !bear.PrimalBite.Cast(sim, bear.CurrentTarget) {
		t.Fatal("primal bite did not cast under berserk")
	}
	struck := 0
	for _, target := range sim.Encounter.ActiveTargetUnits {
		if m := bear.PrimalBite.SpellMetrics[target.UnitIndex]; m.Hits+m.Crits+m.Misses+m.Dodges+m.Parries+m.Glances+m.Blocks > 0 {
			struck++
		}
	}
	if struck != 3 {
		t.Errorf("primal bite under berserk struck %d targets, want 3", struck)
	}
}

// Bash 8983 states a melee defense type and a miss refund: a dodged Bash stuns nothing and gives rage back.
func TestDodgedBashStunsNothingAndRefunds(t *testing.T) {
	sim, bear := newBear(t)
	target := sim.Encounter.ActiveTargetUnits[0]
	target.PseudoStats.BaseDodgeChance = 1
	before := bear.CurrentRage()
	if !bear.Bash.Cast(sim, bear.CurrentTarget) {
		t.Fatal("bash did not cast")
	}
	if m := bear.Bash.SpellMetrics[target.UnitIndex]; m.Dodges != 1 {
		t.Fatalf("bash dodges = %d, want 1", m.Dodges)
	}
	for _, aura := range bear.Bash.RelatedAuraArrays {
		if aura.Get(target).IsActive() {
			t.Error("a dodged bash stunned the target")
		}
	}
	if spent := before - bear.CurrentRage(); spent >= bear.Bash.DefaultCast.Cost {
		t.Errorf("a dodged bash spent %v of its %v rage", spent, bear.Bash.DefaultCast.Cost)
	}
}

// Berserk effect 2 is -100% on Primal Bite's 6 s cooldown, while Berserk is up.
func TestBerserkLiftsPrimalBiteCooldownOnlyWhileUp(t *testing.T) {
	sim, bear := newBearWithTalents(t, "0000000000000000-00000000000100000001-0000000000000000")
	metrics := bear.NewRageMetrics(core.ActionID{SpellID: 1})
	cast := func(when string) {
		t.Helper()
		bear.AddRage(sim, 100, metrics)
		if !bear.PrimalBite.Cast(sim, bear.CurrentTarget) {
			t.Fatalf("primal bite did not cast %s", when)
		}
		sim.CurrentTime += 2 * time.Second
	}

	bear.BerserkAura.Activate(sim)
	cast("under berserk")
	if !bear.PrimalBite.IsReady(sim) {
		t.Errorf("primal bite is on cooldown until %v under berserk", bear.PrimalBite.ReadyAt())
	}
	bear.BerserkAura.Deactivate(sim)
	cast("after berserk")
	if got, want := bear.PrimalBite.ReadyAt(), sim.CurrentTime+4*time.Second; got != want {
		t.Errorf("primal bite is ready at %v after berserk, want %v", got, want)
	}
}

func TestFrenziedRegenerationSpendsRageForHealth(t *testing.T) {
	sim, bear := newBear(t)
	bear.AddRage(sim, 100, bear.NewRageMetrics(core.ActionID{SpellID: 1}))
	before := bear.CurrentRage()
	if !bear.FrenziedRegeneration.Cast(sim, bear.CurrentTarget) {
		t.Fatal("frenzied regeneration did not cast")
	}
	for sim.CurrentTime < 1100*time.Millisecond {
		if sim.Step() {
			break
		}
	}
	if spent := before - bear.CurrentRage(); spent < 10 {
		t.Errorf("frenzied regeneration spent %v rage in its first tick, want 10", spent)
	}
	healed := 0.0
	for _, m := range bear.FrenziedRegeneration.SpellMetrics {
		healed += m.TotalHealing
	}
	if healed <= 0 {
		t.Error("frenzied regeneration healed nothing")
	}
}

func TestFirstLacerateHitsForNothing(t *testing.T) {
	sim, bear := newBear(t)
	target := bear.CurrentTarget
	metrics := bear.NewRageMetrics(core.ActionID{SpellID: 1})
	for range 20 {
		bear.AddRage(sim, 50, metrics)
		bear.Lacerate.Cast(sim, target)
		if bear.Lacerate.Dot(target).IsActive() {
			break
		}
		sim.CurrentTime += 2 * time.Second
	}
	if !bear.Lacerate.Dot(target).IsActive() {
		t.Fatal("lacerate never landed")
	}
	if got := bear.Lacerate.SpellMetrics[target.UnitIndex].TotalDamage; got != 0 {
		t.Errorf("the first lacerate dealt %v, want 0 with no stack already on the target", got)
	}
}

func TestLacerateStacksToFive(t *testing.T) {
	sim, bear := newBear(t)
	if !bear.InForm(druid.Bear) {
		t.Fatal("the bear did not start in bear form")
	}
	target := bear.CurrentTarget
	metrics := bear.NewRageMetrics(core.ActionID{SpellID: 1})
	for range 40 {
		bear.AddRage(sim, 50, metrics)
		bear.Lacerate.Cast(sim, target)
		sim.CurrentTime += 2 * time.Second
		if bear.Lacerate.Dot(target).GetStacks() >= 5 {
			break
		}
	}
	if got := bear.Lacerate.Dot(target).GetStacks(); got != 5 {
		t.Errorf("lacerate stacks = %d, want 5", got)
	}
}

func TestMaulReplacesTheNextSwing(t *testing.T) {
	sim, bear := newBear(t)
	if bear.TryMaul(sim, nil) != nil {
		t.Fatal("maul replaced a swing before it was queued")
	}
	if !bear.Maul.Cast(sim, bear.CurrentTarget) {
		t.Fatal("maul did not queue")
	}
	for sim.CurrentTime < 100*time.Millisecond {
		if sim.Step() {
			break
		}
	}
	replaced := bear.TryMaul(sim, nil)
	if replaced == nil || replaced.ActionID.SpellID != 9881 {
		t.Fatalf("the queued swing was %v, want maul 9881", replaced)
	}
}

func TestFeralChargeCastsOutsideMelee(t *testing.T) {
	sim, bear := newBearWithTalents(t, "0000000000000000-00000000100000000000-0000000000000000")
	bear.DistanceFromTarget = 10
	if !bear.FeralCharge.Cast(sim, bear.CurrentTarget) {
		t.Fatal("feral charge did not cast at 10 yards")
	}
	if bear.FeralCharge.CD.Timer != nil {
		bear.FeralCharge.CD.Reset()
	}
	bear.DistanceFromTarget = 0
	if bear.FeralCharge.CanCast(sim, bear.CurrentTarget) {
		t.Error("feral charge could be cast in melee")
	}
}

func TestSwipeDoesNotScaleWithAttackPower(t *testing.T) {
	sim, bear := newBear(t)
	target := bear.CurrentTarget
	metrics := bear.NewRageMetrics(core.ActionID{SpellID: 1})
	for range 20 {
		bear.AddRage(sim, 50, metrics)
		bear.Swipe.Cast(sim, target)
		sim.CurrentTime += 2 * time.Second
		if bear.Swipe.SpellMetrics[target.UnitIndex].Hits > 0 {
			break
		}
	}
	hits := bear.Swipe.SpellMetrics[target.UnitIndex]
	if hits.Hits == 0 {
		t.Fatal("swipe never hit")
	}
	perHit := hits.TotalDamage / float64(hits.Hits)
	// The row's damage is 83. Attack power on a bear is far larger than the gap this allows.
	if perHit < 40 || perHit > 160 {
		t.Errorf("swipe hit for %.1f, want the row's damage without attack power", perHit)
	}
}
