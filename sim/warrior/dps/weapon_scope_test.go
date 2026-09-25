package dps

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

func orcWarrior(t *testing.T, gearSet string, talents string) *DpsWarrior {
	t.Helper()
	player := &proto.Player{
		Name:          "Weapon Scope",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarrior,
		Equipment:     core.GetGearSet("../../../ui/specs/warrior/dps/gear_sets", gearSet).GearSet,
		TalentsString: talents,
		Spec:          DefaultOptions,
		Rotation:      core.GetAplRotation("../../../ui/specs/warrior/dps/apls", "fury").Rotation,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim.Raid.Parties[0].Players[0].(*DpsWarrior)
}

// Dual Wield Specialization 5/5 gives the off-hand axe's attacks 10% hit and the main hand's none.
// Cruelty 5/5 and Axe Specialization give each axe 6% crit.
func TestFuryTalentsReachTheirWeapons(t *testing.T) {
	warrior := orcWarrior(t, "p1_fury", DefaultFuryTalents)
	mh, oh := warrior.AutoAttacks.MHAuto(), warrior.AutoAttacks.OHAuto()

	if mh.BonusHitPercent != 0 || oh.BonusHitPercent != 10 {
		t.Errorf("main-hand and off-hand hit = %0.2f%% and %0.2f%%, want 0%% and 10%%", mh.BonusHitPercent, oh.BonusHitPercent)
	}
	if mh.BonusCritPercent != 6 || oh.BonusCritPercent != 6 {
		t.Errorf("main-hand and off-hand crit = %0.2f%% and %0.2f%%, want 6%% each", mh.BonusCritPercent, oh.BonusCritPercent)
	}
}

// A two-handed axe takes Cruelty's 5%, Axe Specialization's 1% and Weaponmaster's 5% crit.
func TestArmsTalentsReachTheTwoHandedAxe(t *testing.T) {
	warrior := orcWarrior(t, "p1_arms", DefaultArmsTalents)

	if crit := warrior.AutoAttacks.MHAuto().BonusCritPercent; crit != 11 {
		t.Errorf("two-handed axe crit = %0.2f%%, want 11%%", crit)
	}
}
