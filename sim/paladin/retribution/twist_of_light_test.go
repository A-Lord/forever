package retribution

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/paladin"
)

// The white hit that lands consumes Echo of Righteousness at once, and the replayed Seal of
// Righteousness hit resolves a batch window later, as a live seal's does.
func TestEchoResolvesABatchAfterTheHitThatConsumesIt(t *testing.T) {
	const sealOfRighteousness, sealOfCommand, echoOfRighteousness, sealOfRighteousnessHit = 20293, 20920, 1311704, 25713

	player := &proto.Player{
		Name:          "Twist",
		Race:          proto.Race_RaceHuman,
		Class:         proto.Class_ClassPaladin,
		Equipment:     core.GetGearSet("../../../ui/specs/paladin/retribution/gear_sets", "default").GearSet,
		TalentsString: "35003-503-05225031011330321",
		Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
			Options: &proto.RetributionPaladin_Options{ClassOptions: &proto.PaladinOptions{}},
		}},
		Rotation: &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	pal := sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin()
	target := sim.Encounter.AllTargetUnits[0]

	if !pal.Talents.TwistOfLight {
		t.Fatal("the talents do not take Twist of Light")
	}
	pal.GetSpell(core.ActionID{SpellID: sealOfRighteousness}).SkipCastAndApplyEffects(sim, target)
	pal.GetSpell(core.ActionID{SpellID: sealOfCommand}).SkipCastAndApplyEffects(sim, target)
	echo := pal.GetAuraByID(core.ActionID{SpellID: echoOfRighteousness})
	if echo == nil || !echo.IsActive() {
		t.Fatal("replacing Seal of Righteousness left no Echo of Righteousness")
	}

	hit := pal.GetSpell(core.ActionID{SpellID: sealOfRighteousnessHit})
	casts := func() int32 { return hit.SpellMetrics[target.UnitIndex].Casts }

	start := sim.CurrentTime
	pal.OnSpellHitDealt(sim, pal.AutoAttacks.MHAuto(), &core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: 100})
	if echo.IsActive() {
		t.Error("the Echo is still up after the white hit that landed")
	}
	if got := casts(); got != 0 {
		t.Errorf("the Echo cast Seal of Righteousness %d times in the hit's own tick, want none", got)
	}

	for casts() == 0 && sim.CurrentTime < start+time.Second {
		sim.Step()
	}
	if got := casts(); got != 1 || sim.CurrentTime != start+core.SpellBatchWindow {
		t.Errorf("the Echo cast Seal of Righteousness %d times by %v, want once a batch window after the hit",
			got, sim.CurrentTime-start)
	}
}
