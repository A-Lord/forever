package feralbear

import (
	"google.golang.org/protobuf/encoding/protojson"
	"os"
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	RegisterFeralBearDruid()
	common.RegisterAllEffects()
}

func TestFeralBear(t *testing.T) {
	generators := core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{starterConfig()})
	generators = append(generators, &core.SingleDpsTestGenerator{Name: "StarterCasts", Request: starterRequest(t)})
	core.RunTestSuite(t, t.Name(), generators)
}

func BenchmarkSimulate(b *testing.B) { core.CharacterBenchmark(b, starterConfig()) }

// The bear UI default. Feral Instinct raises only Swipe, which a single target never sees, so its
// points are in Blood Frenzy (rage on crits) and Shredding Attacks.
const DefaultTalents = "0000000000000000-55200312120132012551-0550000000000000"
const DemoRoarTalents = DefaultTalents

var DefaultSpecOptions = &proto.Player_FeralBearDruid{
	FeralBearDruid: &proto.FeralBearDruid{
		Options: &proto.FeralBearDruid_Options{
			StartingRage: 25,
		},
	},
}

var DefaultConsumables = &proto.ConsumesSpec{
	PotId:      13455,                // Greater Stoneshield Potion
	ElixirIds:  []int32{13452, 9088}, // Elixir of the Mongoose, Gift of Arthas
	FoodId:     21023,                // Dirge's Kickin' Chimaerok Chops
	ConjuredId: 9421,                 // Major Healthstone
	// Goblin Sapper Charge, Scroll of Agility IV, Scroll of Strength IV, Scroll of Protection IV
	ExtraIds: []int32{10646, 10309, 10310, 10305},
}

func starterConfig() core.CharacterSuiteConfig {
	reference := proto.Stat_StatAttackPower
	weigh := []proto.Stat{proto.Stat_StatStrength, proto.Stat_StatAgility, proto.Stat_StatStamina}
	config := core.GetTestBuildFromJSON(
		proto.Class_ClassDruid, "../../../ui/specs/druid/feralbear/builds", "starter",
		core.ItemFilter{
			ArmorType: proto.ArmorType_ArmorTypeLeather,
			WeaponTypes: []proto.WeaponType{
				proto.WeaponType_WeaponTypeDagger, proto.WeaponType_WeaponTypeFist,
				proto.WeaponType_WeaponTypeMace, proto.WeaponType_WeaponTypeStaff,
			},
			RangedWeaponTypes: []proto.RangedWeaponType{proto.RangedWeaponType_RangedWeaponTypeIdol},
		}, &reference, &weigh,
	)
	config.OtherRaces = []proto.Race{proto.Race_RaceTauren}
	config.Rotation = core.GetAplRotation("../../../ui/specs/druid/feralbear/apls", "lacerate_first")
	return config
}

func starterRequest(t testing.TB) *proto.RaidSimRequest {
	t.Helper()
	data, err := os.ReadFile("../../../ui/specs/druid/feralbear/builds/starter.build.json")
	if err != nil {
		t.Fatal(err)
	}
	settings := &proto.IndividualSimSettings{}
	if err := protojson.Unmarshal(data, settings); err != nil {
		t.Fatal(err)
	}
	settings.Player.Rotation = core.GetAplRotation("../../../ui/specs/druid/feralbear/apls", "lacerate_first").Rotation
	raid := core.SinglePlayerRaidProto(settings.Player, settings.PartyBuffs, settings.RaidBuffs, settings.Debuffs)
	raid.Tanks = settings.Tanks
	if settings.Encounter == nil {
		settings.Encounter = &proto.Encounter{Duration: core.LongDuration, Targets: []*proto.Target{core.FreshDefaultTargetConfig()}}
	}
	return &proto.RaidSimRequest{Raid: raid, Encounter: settings.Encounter, SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1, IsTest: true}}
}

func TestFeralBearStarterCasts(t *testing.T) {
	request := starterRequest(t)
	for _, race := range []proto.Race{proto.Race_RaceNightElf, proto.Race_RaceTauren} {
		t.Run(race.String(), func(t *testing.T) {
			request.Raid.Parties[0].Players[0].Race = race
			if race != proto.Race_RaceNightElf {
				player := request.Raid.Parties[0].Players[0]
				player.Rotation.PriorityList = slices.DeleteFunc(player.Rotation.PriorityList, func(item *proto.APLListItem) bool {
					return item.Action.GetCastSpell().GetSpellId().GetSpellId() == 1259799
				})
			}
			stats := core.ComputeStats(&proto.ComputeStatsRequest{Raid: request.Raid})
			for _, action := range stats.RaidStats.Parties[0].Players[0].RotationStats.PriorityList {
				for _, validation := range action.Validations {
					if validation.LogLevel >= proto.LogLevel_Warning {
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
			for _, id := range []int32{1238073, 1235827, 9881, 9898} {
				if casts[id] == 0 {
					t.Errorf("starter rotation never casts spell %d", id)
				}
			}
			if player.Dtps.Avg <= 0 {
				t.Fatal("assigned bear receives no damage")
			}
			t.Logf("casts=%v DPS=%g DTPS=%g", casts, player.Dps.Avg, player.Dtps.Avg)
		})
	}
}
