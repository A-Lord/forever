//go:build with_db

package feralcat

import (
	"os"
	"testing"

	"github.com/wowsims/forever/assets/database"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestCatBuildGearMatchesFaction(t *testing.T) {
	items := make(map[int32]*proto.UIItem)
	for _, item := range database.Load().Items {
		items[item.Id] = item
	}
	for _, name := range []string{"starter", "raid"} {
		t.Run(name, func(t *testing.T) {
			player := buildRequest(t, name).Raid.Parties[0].Players[0]
			for _, equipped := range player.Equipment.Items {
				if equipped.Id == 0 {
					continue
				}
				item := items[equipped.Id]
				if item == nil {
					t.Fatalf("equipped item %d is absent from the embedded database", equipped.Id)
				}
				if player.Race == proto.Race_RaceNightElf && item.FactionRestriction == proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY {
					t.Errorf("Night Elf build equips Horde-only item %d %s", item.Id, item.Name)
				}
				if player.Race == proto.Race_RaceTauren && item.FactionRestriction == proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY {
					t.Errorf("Tauren build equips Alliance-only item %d %s", item.Id, item.Name)
				}
			}
		})
	}
}

func starterRequest(t testing.TB) *proto.RaidSimRequest {
	return buildRequest(t, "starter")
}

func buildRequest(t testing.TB, name string) *proto.RaidSimRequest {
	t.Helper()
	data, err := os.ReadFile("../../../ui/specs/druid/feralcat/builds/" + name + ".build.json")
	if err != nil {
		t.Fatal(err)
	}
	settings := &proto.IndividualSimSettings{}
	if err := protojson.Unmarshal(data, settings); err != nil {
		t.Fatal(err)
	}
	raid := core.SinglePlayerRaidProto(settings.Player, settings.PartyBuffs, settings.RaidBuffs, settings.Debuffs)
	raid.Tanks = settings.Tanks
	if settings.Encounter == nil {
		settings.Encounter = &proto.Encounter{Duration: core.LongDuration, Targets: []*proto.Target{core.FreshDefaultTargetConfig()}}
	}
	return &proto.RaidSimRequest{Raid: raid, Encounter: settings.Encounter, SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1, IsTest: true}}
}

func TestFeralCatStarterCasts(t *testing.T) {
	request := starterRequest(t)
	for _, race := range []proto.Race{proto.Race_RaceNightElf, proto.Race_RaceTauren} {
		t.Run(race.String(), func(t *testing.T) {
			request.Raid.Parties[0].Players[0].Race = race
			stats := core.ComputeStats(&proto.ComputeStatsRequest{Raid: request.Raid})
			for _, action := range stats.RaidStats.Parties[0].Players[0].RotationStats.PriorityList {
				for _, validation := range action.Validations {
					if validation.LogLevel != proto.LogLevel_Information {
						t.Fatalf("rotation validation: %v", validation)
					}
				}
			}
			result := core.RunRaidSim(request)
			if result.Error != nil {
				t.Fatal(result.Error.Message)
			}
			player := result.RaidMetrics.Parties[0].Players[0]
			casts := make(map[int32]int32)
			for _, action := range player.Actions {
				for _, target := range action.Targets {
					casts[action.Id.GetSpellId()] += target.Casts
				}
			}
			for _, id := range []int32{9830, 9896, 9904, 1322605} {
				if casts[id] == 0 {
					t.Errorf("starter rotation never casts spell %d", id)
				}
			}
			t.Logf("casts=%v DPS=%g DTPS=%g", casts, player.Dps.Avg, player.Dtps.Avg)
		})
	}
}
