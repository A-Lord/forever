//go:build with_db

package feralcat

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	goproto "google.golang.org/protobuf/proto"
)

func TestCatBuildsUseNativeRaidTargetArmor(t *testing.T) {
	for _, name := range []string{"starter", "raid"} {
		t.Run(name, func(t *testing.T) {
			request := buildRequest(t, name)
			target := request.Encounter.Targets[0]
			defaultTarget := core.FreshDefaultTargetConfig()
			if target.Level != core.CharacterLevel+3 || target.Stats[stats.Armor] != defaultTarget.Stats[stats.Armor] || target.Stats[stats.Health] != defaultTarget.Stats[stats.Health] {
				t.Fatalf("Cat build overrides native raid target level, armor, or health: %v", target)
			}
			armored := core.RunRaidSim(request)
			unarmoredRequest := goproto.Clone(request).(*proto.RaidSimRequest)
			unarmoredRequest.Encounter.Targets[0].Stats[stats.Armor] = 0
			unarmored := core.RunRaidSim(unarmoredRequest)
			if armored.Error != nil || unarmored.Error != nil {
				t.Fatalf("armor comparison failed: %v / %v", armored.Error, unarmored.Error)
			}
			if armored.RaidMetrics.Dps.Avg >= unarmored.RaidMetrics.Dps.Avg {
				t.Fatalf("native armor did not reduce Cat damage: armored %.3f, unarmored %.3f", armored.RaidMetrics.Dps.Avg, unarmored.RaidMetrics.Dps.Avg)
			}
		})
	}
}

func TestMaintainFaerieFireUsesSharedDebuffRefresh(t *testing.T) {
	rotation := catRotation(&proto.FeralCatDruid_Rotation{FinishingMove: proto.FeralCatDruid_Rotation_Rip, RipMinComboPoints: 5, MaintainFaerieFire: true})
	for _, external := range []bool{false, true} {
		request := catPowershiftRequest(powershiftTalents, rotation, false)
		request.Raid.Debuffs = &proto.Debuffs{FaerieFire: external}
		player := runCat(t, request)
		casts := castsOf(player.Actions, 9907)
		if external && casts != 0 {
			t.Errorf("Cat replaced an active external Faerie Fire %d times", casts)
		}
		if !external && casts == 0 {
			t.Error("Cat did not supply the missing Faerie Fire")
		}
	}
}

func TestRaidDefaultRunsSelectedCatRotation(t *testing.T) {
	player := runCat(t, buildRequest(t, "raid"))
	for _, id := range []int32{1322605, 9830, 9896, 417141} {
		if castsOf(player.Actions, id) == 0 {
			t.Errorf("raid default never casts spell %d", id)
		}
	}
	for _, id := range []int32{768, 9634, 9904, 22829, 9907} {
		if got := castsOf(player.Actions, id); got != 0 {
			t.Errorf("raid default casts spell %d %d times", id, got)
		}
	}
}
