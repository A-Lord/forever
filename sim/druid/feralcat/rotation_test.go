//go:build with_db

package feralcat

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

// Same tree as the cat UI default: shifting power, improved shifting power,
// shredding attacks, and natural shapeshifter.
const powershiftTalents = "0000000000000000-50000032121032212051-0500300000000000"

func TestPowershiftUsesShiftingPower(t *testing.T) {
	result := core.RunSim(catPowershiftRequest(), nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatalf("sim error: %s", result.Error.Message)
	}
	player := result.RaidMetrics.Parties[0].Players[0]
	if player.Dps.Avg <= 0 {
		t.Fatalf("dps = %v, want a fight that deals damage", player.Dps.Avg)
	}
	if got := castsOf(player.Actions, 1322605); got == 0 {
		t.Fatal("shifting power was never cast")
	}
	if got := castsOf(player.Actions, 9830); got == 0 {
		t.Fatal("shred was never cast")
	}
	if got := castsOf(player.Actions, 9896); got == 0 {
		t.Fatal("rip was never cast")
	}
	if got := castsOf(player.Actions, 768); got != 0 {
		t.Fatalf("cat form casts = %d, the energy button left cat form", got)
	}
}

func catPowershiftRequest() *proto.RaidSimRequest {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 866
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1, IsTest: true},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Cat", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
				TalentsString:  powershiftTalents,
				ReactionTimeMs: 100,
				Equipment:      &proto.EquipmentSpec{Items: items},
				Consumables:    &proto.ConsumesSpec{},
				Buffs:          &proto.IndividualBuffs{},
				Spec: &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{
					Options: &proto.FeralCatDruid_Options{},
				}},
				Rotation: &proto.APLRotation{
					PriorityList: []*proto.APLListItem{{
						Action: &proto.APLAction{
							Action: &proto.APLAction_CatOptimalRotationAction{
								CatOptimalRotationAction: &proto.APLActionCatOptimalRotationAction{
									FinishingMove:      proto.FeralCatDruid_Rotation_Rip,
									Biteweave:          true,
									RipMinComboPoints:  5,
									BiteMinComboPoints: 5,
									MaintainFaerieFire: true,
								},
							},
						},
					}},
				},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 90},
	}
}

func castsOf(actions []*proto.ActionMetrics, spellID int32) int32 {
	var n int32
	for _, action := range actions {
		if action.GetId().GetSpellId() != spellID {
			continue
		}
		for _, target := range action.Targets {
			n += target.Casts
		}
	}
	return n
}
