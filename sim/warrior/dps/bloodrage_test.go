package dps

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

// Bloodrage 2687 restores 10 rage and casts 29131, a 10 s aura that restores 1 rage a second: the
// aura is up for its duration and the cast is worth 20 rage in all.
func TestBloodrageRestoresRageOverItsAura(t *testing.T) {
	player := &proto.Player{
		Name:      "Bloodrage",
		Race:      proto.Race_RaceOrc,
		Class:     proto.Class_ClassWarrior,
		Equipment: &proto.EquipmentSpec{},
		Spec:      DefaultOptions,
		Rotation:  &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	war := sim.Raid.Parties[0].Players[0].(*DpsWarrior)
	target := sim.Encounter.AllTargetUnits[0]
	war.AutoAttacks.CancelAutoSwing(sim)
	target.AutoAttacks.CancelAutoSwing(sim)

	bloodrage := war.GetSpell(core.ActionID{SpellID: 2687})
	aura := war.GetAuraByID(core.ActionID{SpellID: 29131})
	if bloodrage == nil || aura == nil {
		t.Fatalf("Bloodrage %v, its aura %v; want both registered", bloodrage, aura)
	}

	before := war.CurrentRage()
	bloodrage.SkipCastAndApplyEffects(sim, &war.Unit)
	if got := war.CurrentRage() - before; got != 10 {
		t.Errorf("the cast restored %v rage at once, want 10", got)
	}
	if !aura.IsActive() || aura.Duration != 10*time.Second {
		t.Fatalf("the aura is up: %v for %v, want up for 10s", aura.IsActive(), aura.Duration)
	}

	end := sim.CurrentTime + aura.Duration + time.Second/2
	sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{DoAt: end, OnAction: func(*core.Simulation) {}}))
	for steps := 0; sim.CurrentTime < end; steps++ {
		if steps > 10000 {
			t.Fatalf("the sim did not reach %v", end)
		}
		sim.Step()
	}

	if aura.IsActive() {
		t.Error("the aura is still up after its 10 s")
	}
	if got := war.CurrentRage() - before; got != 20 {
		t.Errorf("the cast restored %v rage in all, want 20", got)
	}
}
