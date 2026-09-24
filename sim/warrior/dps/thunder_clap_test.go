package dps

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

// With five pieces of Conqueror's Battlegear the warrior's clap bids its whole
// 30% slow in the attack-speed category. The raid's plain 20% clap loses to it,
// a slow bidding past it keeps the category, and the target never carries two
// slows at once.
func TestThunderClapBidsItsWholeSlowInTheAttackSpeedCategory(t *testing.T) {
	setup := func(t *testing.T) (*core.Simulation, *core.Unit, *core.Spell, *core.Aura, *core.Aura) {
		t.Helper()

		player := &proto.Player{
			Name:  "Conqueror",
			Race:  proto.Race_RaceOrc,
			Class: proto.Class_ClassWarrior,
			Equipment: &proto.EquipmentSpec{Items: []*proto.ItemSpec{
				{Id: 21329}, {}, {Id: 21330}, {}, {Id: 21331}, {}, {}, {}, {Id: 21332}, {Id: 21333},
			}},
			TalentsString: DefaultArmsTalents,
			Spec:          DefaultOptions,
		}
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid: core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{},
				&proto.Debuffs{ThunderClap: true}),
			Encounter:  core.MakeSingleTargetEncounter(0),
			SimOptions: &proto.SimOptions{RandomSeed: 1},
		}, simsignals.CreateSignals())
		sim.Reset()

		war := sim.Raid.Parties[0].Players[0].(*DpsWarrior)
		target := sim.Encounter.AllTargetUnits[0]
		clap := war.GetSpell(core.ActionID{SpellID: 11581})
		own := target.GetAura("Thunder Clap (Player)")
		external := target.GetAura("Thunder Clap (External)")
		if clap == nil || own == nil || external == nil {
			t.Fatalf("clap %v, own copy %v, raid copy %v; want all three registered", clap, own, external)
		}
		if !external.IsActive() {
			t.Fatal("the raid's clap is not up at the pull")
		}
		return sim, target, clap, own, external
	}

	landClap := func(t *testing.T, sim *core.Simulation, target *core.Unit, clap *core.Spell, own *core.Aura) {
		t.Helper()
		for i := 0; i < 50 && !own.IsActive(); i++ {
			clap.SkipCastAndApplyEffects(sim, target)
		}
		if !own.IsActive() {
			t.Fatal("no clap landed in 50 tries")
		}
	}

	speedIs := func(t *testing.T, target *core.Unit, want float64, context string) {
		t.Helper()
		if got := target.PseudoStats.MeleeSpeedMultiplier; math.Abs(got-want) > 1e-9 {
			t.Errorf("%s: the target swings at %v of its speed, want %v", context, got, want)
		}
	}

	t.Run("the raid's weaker clap loses to the set", func(t *testing.T) {
		sim, target, clap, own, external := setup(t)
		speedIs(t, target, 0.8, "the raid's clap alone")

		landClap(t, sim, target, clap, own)
		if got := own.ExclusiveEffects[0].Priority; math.Abs(got-0.3) > 1e-9 {
			t.Errorf("the warrior's clap bids %v, want its whole 0.3", got)
		}
		if !own.ExclusiveEffects[0].IsActive() || external.ExclusiveEffects[0].IsActive() {
			t.Errorf("the warrior's clap holds the category %v and the raid's %v, want the warrior's alone",
				own.ExclusiveEffects[0].IsActive(), external.ExclusiveEffects[0].IsActive())
		}
		speedIs(t, target, 0.7, "the warrior's clap over the raid's")
	})

	t.Run("a stronger slow keeps the category", func(t *testing.T) {
		sim, target, clap, own, external := setup(t)
		external.ExclusiveEffects[0].SetPriority(sim, 0.5)

		landClap(t, sim, target, clap, own)
		if own.ExclusiveEffects[0].IsActive() {
			t.Error("the warrior's clap took the category from a slow bidding past it")
		}
		speedIs(t, target, 0.8, "the stronger slow over the warrior's clap")

		external.Deactivate(sim)
		if !own.ExclusiveEffects[0].IsActive() {
			t.Error("the warrior's clap did not take the category once the stronger slow dropped")
		}
		speedIs(t, target, 0.7, "the warrior's clap after the stronger slow dropped")
	})
}
