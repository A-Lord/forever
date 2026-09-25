package dps

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// A warrior fighting under Flametongue Totem strikes the target with the
// totem's fire on its main-hand hits, and an unresisted strike that does not
// crit is the client's 1363 x the main hand's speed / 100.
func TestFlametongueTotemStrikesOffAFightingWarriorsMainHand(t *testing.T) {
	const arcaniteReaper = 12784
	reaper := core.GetItemByID(arcaniteReaper)
	if reaper == nil {
		t.Fatalf("item %d is not in the database", arcaniteReaper)
	}

	equipment := &proto.EquipmentSpec{}
	for slot := proto.ItemSlot_ItemSlotHead; slot <= proto.ItemSlot_ItemSlotMainHand; slot++ {
		equipment.Items = append(equipment.Items, &proto.ItemSpec{})
	}
	equipment.Items[proto.ItemSlot_ItemSlotMainHand].Id = arcaniteReaper
	player := &proto.Player{
		Name:          "Flametongue",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarrior,
		Equipment:     equipment,
		TalentsString: DefaultFuryTalents,
		Spec:          DefaultOptions,
		Rotation:      core.GetAplRotation("../../../ui/specs/warrior/dps/apls", "fury").Rotation,
	}
	result := core.RunRaidSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(player, &proto.PartyBuffs{FlametongueTotem: true}, &proto.RaidBuffs{},
			&proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1, IsTest: true},
	})
	if result.Error != nil {
		t.Fatalf("the sim failed: %v", result.Error.Message)
	}

	var strikes *proto.TargetedActionMetrics
	for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
		if action.Id.GetSpellId() == 16389 {
			strikes = action.Targets[0]
		}
	}
	if strikes == nil || strikes.Damage <= 0 {
		t.Fatalf("the fight's metrics hold %v for spell 16389, want the totem's fire on the target", strikes)
	}

	plainHits := strikes.Hits - strikes.ResistedHits
	plainDamage := strikes.Damage - strikes.CritDamage - (strikes.ResistedDamage - strikes.ResistedCritDamage)
	if plainHits == 0 {
		t.Fatalf("none of %d strikes landed unresisted without a crit", strikes.Casts)
	}
	want := 1363 * reaper.SwingSpeed / 100
	if got := plainDamage / float64(plainHits); math.Abs(got-want) > 1e-6 {
		t.Errorf("an unresisted strike dealt %v, want %v for a %v speed main hand", got, want, reaper.SwingSpeed)
	}
}
