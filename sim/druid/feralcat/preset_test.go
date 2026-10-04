//go:build with_db

package feralcat

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func TestFeralPreset(t *testing.T) {
	for _, race := range []proto.Race{proto.Race_RaceNightElf, proto.Race_RaceTauren} {
		t.Run(race.String(), func(t *testing.T) {
			rotation := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "default").Rotation
			request := coverageRequest(false, rotation)
			player := request.Raid.Parties[0].Players[0]
			player.Race = race
			player.TalentsString = ""

			result := core.RunRaidSim(request)

			if result.Error != nil {
				t.Fatal(result.Error.Message)
			}
			metrics := result.RaidMetrics.Parties[0].Players[0]
			if dps := metrics.Dps.Avg; dps <= 0 || math.IsNaN(dps) || math.IsInf(dps, 0) {
				t.Fatalf("minimal cat produces invalid damage rate %v", dps)
			}
			casts := make(map[int32]int32)
			for _, action := range metrics.Actions {
				for _, target := range action.Targets {
					casts[action.Id.GetSpellId()] += target.Casts
				}
			}
			for _, id := range []int32{9850, 9896} {
				if casts[id] == 0 {
					t.Errorf("minimal cat never casts spell %d", id)
				}
			}
			t.Logf("Claw casts=%d, Rip casts=%d, DPS=%g", casts[9850], casts[9896], metrics.Dps.Avg)
		})
	}
}
