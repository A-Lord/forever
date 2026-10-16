//go:build with_db

package shaman_test

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Windfury Weapon's strikes are white in name only: Flametongue Totem, which hears auto attacks
// alone, procs off every one that lands, yet they roll the special attack table and never glance.
func TestWindfuryWeaponStrikesProcFlametongueTotemAndNeverGlance(t *testing.T) {
	player := windfuryShaman(&proto.APLRotation{})
	result := core.RunRaidSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(player, &proto.PartyBuffs{},
			&proto.RaidBuffs{FlametongueTotem: true}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{Iterations: 20, RandomSeed: 1},
	})
	if result.Error != nil {
		t.Fatalf("the sim failed: %s", result.Error.Message)
	}

	// A proc logs no casts, so each count is the outcomes it logged.
	strikeTag := spelldata.MustFind(16362).Refs()[0].ID
	flametongueTotem := spelldata.MustFind(16387).ID
	var mainHand, strikes, flametongue proto.TargetedActionMetrics
	for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
		var into *proto.TargetedActionMetrics
		switch {
		case action.Id.GetOtherId() == proto.OtherAction_OtherActionAttack && action.Id.Tag == 1:
			into = &mainHand
		case action.Id.GetOtherId() == proto.OtherAction_OtherActionAttack && action.Id.Tag == strikeTag:
			into = &strikes
		case action.Id.GetSpellId() == flametongueTotem && action.Id.Tag == 1:
			into = &flametongue
		default:
			continue
		}
		for _, target := range action.Targets {
			into.Hits += target.Hits + target.Blocks
			into.Crits += target.Crits + target.BlockedCrits
			into.Glances += target.Glances + target.GlanceBlocks
			into.Misses += target.Misses + target.Dodges + target.Parries
		}
	}
	landed := func(m *proto.TargetedActionMetrics) int32 { return m.Hits + m.Crits + m.Glances }

	if landed(&strikes) == 0 {
		t.Fatal("no Windfury Weapon strike landed over the run")
	}
	if strikes.Glances != 0 {
		t.Errorf("Windfury Weapon's strikes glanced %d times, want none", strikes.Glances)
	}
	if mainHand.Glances == 0 {
		t.Error("the main hand's swings never glanced, so the run cannot tell the strikes from them")
	}
	if got, want := landed(&flametongue)+flametongue.Misses, landed(&mainHand)+landed(&strikes); got != want {
		t.Errorf("Flametongue Totem procced %d times, want one for each of the %d swings and %d Windfury Weapon strikes that landed",
			got, landed(&mainHand), landed(&strikes))
	}
}
