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
	return newBearWithTalents(t, "0000000000000000-0000000000000000000-0000000000000000")
}

func newBearWithTalents(t *testing.T, talents string) (*core.Simulation, *GuardianDruid) {
	t.Helper()
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
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 60},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*GuardianDruid)
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
	sim, bear := newBearWithTalents(t, "0000000000000000-0000000100000000000-0000000000000000")
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
