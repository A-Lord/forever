//go:build with_db

package feralbear

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

// Heart of the wild, feral instinct, thick hide, feral charge, primal bite,
// leader of the pack, natural reaction.
const bearTalents = "0000000000000000-05030300100100010500-0000000000000000"

func TestBearSimpleRotationDealsDamage(t *testing.T) {
	result := core.RunSim(bearSimpleRequest(), nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatalf("sim error: %s", result.Error.Message)
	}
	player := result.RaidMetrics.Parties[0].Players[0]
	if player.Dps.Avg <= 0 {
		t.Fatalf("dps = %v, want a fight that deals damage", player.Dps.Avg)
	}
	var casts int32
	for _, id := range []int32{1238073, 1235827, 9881, 9908} {
		casts += bearCastsOf(player.Actions, id)
	}
	if casts == 0 {
		t.Fatal("primal bite, lacerate, maul and swipe were never cast")
	}
}

func bearSimpleRequest() *proto.RaidSimRequest {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 866
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1, IsTest: true},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Bear", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
				TalentsString: bearTalents,
				Equipment:     &proto.EquipmentSpec{Items: items},
				Consumables:   &proto.ConsumesSpec{},
				Buffs:         &proto.IndividualBuffs{},
				Spec: &proto.Player_FeralBearDruid{FeralBearDruid: &proto.FeralBearDruid{
					Options: &proto.FeralBearDruid_Options{},
				}},
				Rotation: &proto.APLRotation{
					PriorityList: []*proto.APLListItem{{
						Action: &proto.APLAction{
							Action: &proto.APLAction_BearOptimalRotationAction{
								BearOptimalRotationAction: &proto.APLActionBearOptimalRotationAction{
									MaintainFaerieFire:       true,
									MaintainDemoralizingRoar: true,
									MaulRageThreshold:        50,
									SwipeUsage:               proto.FeralBearDruid_Rotation_SwipeUsage_WithEnoughAP,
									SwipeApThreshold:         2700,
								},
							},
						},
					}},
				},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 60},
	}
}

func bearCastsOf(actions []*proto.ActionMetrics, spellID int32) int32 {
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
