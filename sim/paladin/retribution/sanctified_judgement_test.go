package retribution

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/paladin"
)

// Sanctified Judgement at 3/3 returns 60% of the judged seal's mana cost on every Judgement.
func TestSanctifiedJudgementReturnsSealMana(t *testing.T) {
	player := &proto.Player{
		Name:          "Sanctified Judgement",
		Race:          proto.Race_RaceHuman,
		Class:         proto.Class_ClassPaladin,
		Equipment:     core.GetGearSet("../../../ui/specs/paladin/retribution/gear_sets", "twist_of_light").GearSet,
		TalentsString: TwistOfLightTalents,
		Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
			Options: &proto.RetributionPaladin_Options{ClassOptions: &proto.PaladinOptions{}},
		}},
		Rotation: core.GetAplRotation("../../../ui/specs/paladin/retribution/apls", "twist_of_light").Rotation,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	pal := sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin()

	if pal.Talents.SanctifiedJudgement != 3 {
		t.Fatalf("the talents take Sanctified Judgement %d/3, want 3", pal.Talents.SanctifiedJudgement)
	}

	target := pal.CurrentTarget
	seal := pal.GetSpell(core.ActionID{SpellID: paladin.SealOfRighteousnessRankMap.Highest().ID})
	if seal == nil {
		t.Fatal("Seal of Righteousness is not registered")
	}
	if !seal.Cast(sim, target) {
		t.Fatal("Seal of Righteousness did not cast")
	}

	before := pal.CurrentMana()
	if !pal.Judgement.Cast(sim, target) {
		t.Fatal("Judgement did not cast")
	}

	want := before - pal.Judgement.CurCast.Cost + seal.CurCast.Cost*0.6
	if got := pal.CurrentMana(); math.Abs(got-want) > 1e-6 {
		t.Errorf("mana after Judgement is %.3f, want %.3f (seal cost %.3f, Judgement cost %.3f)",
			got, want, seal.CurCast.Cost, pal.Judgement.CurCast.Cost)
	}
}
