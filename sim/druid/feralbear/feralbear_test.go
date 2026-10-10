package feralbear

import (
	"google.golang.org/protobuf/encoding/protojson"
	"os"
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
	generators := core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassDruid,
			Race:       proto.Race_RaceNightElf,
			OtherRaces: []proto.Race{proto.Race_RaceTauren},

			GearSet: core.GetGearSet("../../../ui/specs/druid/feralbear/gear_sets", "default"),

			Talents: DefaultTalents,
			OtherTalentSets: []core.TalentsCombo{
				{Label: "Hybrid", Talents: HybridTalents},
			},

			SpecOptions: core.SpecOptionsCombo{Label: "Standard", SpecOptions: DefaultSpecOptions},
			OtherSpecOptions: []core.SpecOptionsCombo{
				{Label: "BleedAssumed", SpecOptions: &proto.Player_FeralBearDruid{FeralBearDruid: &proto.FeralBearDruid{
					Options: &proto.FeralBearDruid_Options{StartingRage: 25, AssumeBleedActive: true},
				}}},
			},

			Rotation: core.GetAplRotation("../../../ui/specs/druid/feralbear/apls", "default"),

			Consumables: DefaultConsumables,

			Profession1: proto.Profession_Engineering,
			Profession2: proto.Profession_Enchanting,

			IndividualBuffs: core.FullTankIndividualBuffs,

			IsTank:          true,
			InFrontOfTarget: true,

			ItemFilter: core.ItemFilter{
				ArmorType: proto.ArmorType_ArmorTypeLeather,
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeDagger,
					proto.WeaponType_WeaponTypeFist,
					proto.WeaponType_WeaponTypeMace,
					proto.WeaponType_WeaponTypeStaff,
				},
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeIdol,
				},
			},

			EPReferenceStat: proto.Stat_StatAgility,
			StatsToWeigh: []proto.Stat{
				proto.Stat_StatHealth,
				proto.Stat_StatStamina,
				proto.Stat_StatAgility,
				proto.Stat_StatStrength,
				proto.Stat_StatAttackPower,
				proto.Stat_StatArmor,
				proto.Stat_StatBonusArmor,
				proto.Stat_StatDodgeRating,
				proto.Stat_StatDefenseRating,
				proto.Stat_StatHitRating,
				proto.Stat_StatExpertiseRating,
			},
		},
	})
	generators = append(generators, &core.SingleDpsTestGenerator{Name: "StarterCasts", Request: starterRequest(t)})
	core.RunTestSuite(t, t.Name(), generators)
}

func BenchmarkSimulate(b *testing.B) {
	rsr := &proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Class:         proto.Class_ClassDruid,
				Race:          proto.Race_RaceNightElf,
				TalentsString: DefaultTalents,
				Equipment:     core.GetGearSet("../../../ui/specs/druid/feralbear/gear_sets", "default").GearSet,
				Consumables:   DefaultConsumables,
				Spec:          DefaultSpecOptions,
				Rotation:      core.GetAplRotation("../../../ui/specs/druid/feralbear/apls", "default").Rotation,
			},
			nil, nil, nil,
		),
		Encounter: &proto.Encounter{
			Duration: 300,
			Targets:  []*proto.Target{core.NewDefaultTarget()},
		},
		SimOptions: core.AverageDefaultSimTestOptions,
	}

	core.RaidBenchmark(b, rsr)
}

// The bear UI default. Balance 9 is for Nature's Reach (melee hit) and Nature's Majesty (melee crit).
// Feral Instinct raises only Swipe, which a single target never sees.
const DefaultTalents = "0500220000000000-55200322120132012551-0000000000000000"

// The bear UI's Hybrid preset (presets.ts HybridTalents).
const HybridTalents = "0500220000000000-35230332121132212051-0000000000000000"

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
	settings.Player.Rotation = core.GetAplRotation("../../../ui/specs/druid/feralbear/apls", "default").Rotation
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
