package dps

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

const arcaniteReaper = 12784

func fightWithArcaniteReaper(t *testing.T, partyBuffs *proto.PartyBuffs) map[*proto.ActionID]*proto.TargetedActionMetrics {
	t.Helper()
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
		Raid:       core.SinglePlayerRaidProto(player, partyBuffs, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1, IsTest: true},
	})
	if result.Error != nil {
		t.Fatalf("the sim failed: %v", result.Error.Message)
	}

	actions := map[*proto.ActionID]*proto.TargetedActionMetrics{}
	for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
		actions[action.Id] = action.Targets[0]
	}
	return actions
}

func flametongueStrikes(actions map[*proto.ActionID]*proto.TargetedActionMetrics) *proto.TargetedActionMetrics {
	for id, metrics := range actions {
		if id.GetSpellId() == 16389 && id.GetTag() == -1 {
			return metrics
		}
	}
	return nil
}

// A warrior fighting under Flametongue Totem strikes the target with the
// totem's fire on its main-hand hits, and an unresisted strike that does not
// crit is the game files' 1363 x the main hand's speed / 100.
func TestFlametongueTotemStrikesOffAFightingWarriorsMainHand(t *testing.T) {
	reaper := core.GetItemByID(arcaniteReaper)
	if reaper == nil {
		t.Fatalf("item %d is not in the database", arcaniteReaper)
	}

	strikes := flametongueStrikes(fightWithArcaniteReaper(t, &proto.PartyBuffs{FlametongueTotem: true}))
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

// Under both totems the main hand carries Windfury's enchant: the fight has
// Windfury's extra attacks and no totem fire.
func TestWindfuryTotemTakesTheMainHandFromFlametongueTotem(t *testing.T) {
	actions := fightWithArcaniteReaper(t, &proto.PartyBuffs{WindfuryTotem: true, FlametongueTotem: true})

	if strikes := flametongueStrikes(actions); strikes != nil && strikes.Damage != 0 {
		t.Errorf("Flametongue Totem struck for %v next to Windfury Totem, want nothing", strikes.Damage)
	}

	var extraAttackDamage float64
	for id, metrics := range actions {
		if id.GetOtherId() == proto.OtherAction_OtherActionAttack && id.GetTag() == 25584 {
			extraAttackDamage += metrics.Damage
		}
	}
	if extraAttackDamage <= 0 {
		t.Error("the fight has no Windfury extra attack damage, want Windfury Totem to keep proccing")
	}
}
