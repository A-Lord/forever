package core

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

// An enemy carries the creature family its encounter target names, and a spell reads it off the
// target it lands on. A target that names none is Unknown.
func TestSpellReadsTheCreatureFamilyOfItsTarget(t *testing.T) {
	request := fakeRageSimRequest()
	request.Encounter.Targets = append(request.Encounter.Targets, &proto.Target{
		Name: "naga", Level: 63, MobType: proto.MobType_MobTypeHumanoid, CreatureFamily: proto.CreatureFamily_CreatureFamilyNaga,
		SwingSpeed: 2,
	})
	sim := setupFakeRageSimFrom(request)
	char := sim.Raid.Parties[0].Players[0].GetCharacter()

	var seen []proto.CreatureFamily
	sim.Environment.State = Constructed
	spell := char.RegisterSpell(SpellConfig{
		ActionID: ActionID{SpellID: 910101},
		ApplyEffects: func(_ *Simulation, target *Unit, _ *Spell) {
			seen = append(seen, target.CreatureFamily)
		},
	})
	spell.finalize()
	sim.Environment.State = Finalized

	for _, target := range sim.Encounter.AllTargetUnits {
		spell.SkipCastAndApplyEffects(sim, target)
	}
	want := []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyUnknown, proto.CreatureFamily_CreatureFamilyNaga}
	if !slices.Equal(seen, want) {
		t.Errorf("the spell read the creature families %v off its targets, want %v", seen, want)
	}
}
