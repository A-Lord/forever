package dps

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
)

// The golden suites are skipped while the class is stubbed, so this is what
// checks that the warrior's own enemy debuff is the generated aura. Only
// Demoralizing Shout is registered so far: registerSunderArmor and
// registerThunderClap are still commented out in Warrior.Initialize.
func TestWarriorAppliesTheGeneratedDemoralizingShout(t *testing.T) {
	player := &proto.Player{
		Name:          "Debuffs",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarrior,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: DefaultArmsTalents,
		Spec:          DefaultOptions,
	}

	env, _, _ := core.NewEnvironment(
		core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		core.MakeSingleTargetEncounter(0), false, true)

	war := env.Raid.Parties[0].Players[0].(*DpsWarrior)
	target := env.GetTargetUnitByIndex(0)

	aura := target.GetAura("Demoralizing Shout (Player)")
	if aura == nil {
		t.Fatal("the enemy has no aura labelled \"Demoralizing Shout (Player)\"")
	}
	if want := (core.ActionID{SpellID: 11556, Tag: 0}); aura.ActionID != want {
		t.Errorf("the warrior's own copy is %v, want %v", aura.ActionID, want)
	}
	if aura.Duration != buffs.DemoralizingShoutDuration(0) {
		t.Errorf("the warrior's own copy lasts %v, want the client's 45 seconds", aura.Duration)
	}
	if got := aura.ExclusiveEffects[0].Priority; got != -buffs.DemoralizingShoutValue(0) {
		t.Errorf("the warrior's own copy bids %v, want the magnitude of %v",
			got, buffs.DemoralizingShoutValue(0))
	}
	if war.DemoralizingShoutAuras.Get(target) != aura {
		t.Error("the spell's aura array holds an aura other than the warrior's own shout")
	}
}
