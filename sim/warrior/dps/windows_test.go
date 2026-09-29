package dps

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

func newWindowSim(t *testing.T) (*core.Simulation, *DpsWarrior, *core.Unit) {
	t.Helper()

	player := &proto.Player{
		Name:          "Window",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarrior,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: DefaultFuryTalents,
		Spec:          DefaultOptions,
		Rotation:      &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	return sim, sim.Raid.Parties[0].Players[0].(*DpsWarrior), sim.Encounter.AllTargetUnits[0]
}

// Offensive State 1282735 hears the warrior's melee swings and abilities, so a dodged melee hit
// opens the Overpower window and a dodged hit of any other kind does not.
func TestOverpowerOpensOnADodgedMeleeHit(t *testing.T) {
	cases := []struct {
		name     string
		procMask core.ProcMask
		opens    bool
	}{
		{"a dodged main-hand swing", core.ProcMaskMeleeMHAuto, true},
		{"a dodged melee ability", core.ProcMaskMeleeMHSpecial, true},
		{"a dodged spell", core.ProcMaskSpellDamage, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sim, war, target := newWindowSim(t)

			war.OnSpellHitDealt(sim, &core.Spell{ProcMask: c.procMask, Unit: &war.Unit},
				&core.SpellResult{Target: target, Outcome: core.OutcomeDodge})

			if got := war.OverpowerAura.IsActive(); got != c.opens {
				t.Errorf("the Overpower window is up: %v, want %v", got, c.opens)
			}
		})
	}
}

// Defensive State 5301 hears melee and ranged hits taken, so a block, dodge or parry of one opens
// the Revenge window, a full block that deals no damage included, and an avoided spell does not.
func TestRevengeOpensOnAnAvoidedMeleeOrRangedHit(t *testing.T) {
	cases := []struct {
		name     string
		procMask core.ProcMask
		outcome  core.HitOutcome
		opens    bool
	}{
		{"a dodged swing", core.ProcMaskMeleeMHAuto, core.OutcomeDodge, true},
		{"a parried swing", core.ProcMaskMeleeMHAuto, core.OutcomeParry, true},
		{"a fully blocked swing", core.ProcMaskMeleeMHAuto, core.OutcomeBlock, true},
		{"a dodged ranged attack", core.ProcMaskRangedAuto, core.OutcomeDodge, true},
		{"a landed swing", core.ProcMaskMeleeMHAuto, core.OutcomeHit, false},
		{"a dodged spell", core.ProcMaskSpellDamage, core.OutcomeDodge, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sim, war, target := newWindowSim(t)
			window := war.GetAuraByID(core.ActionID{SpellID: 25288})
			if window == nil {
				t.Fatal("the warrior has no Revenge window")
			}

			war.OnSpellHitTaken(sim, &core.Spell{ProcMask: c.procMask, Unit: target},
				&core.SpellResult{Target: &war.Unit, Outcome: c.outcome})

			if got := window.IsActive(); got != c.opens {
				t.Errorf("the Revenge window is up: %v, want %v", got, c.opens)
			}
		})
	}
}
