package feralcat

import (
	"fmt"
	"github.com/wowsims/forever/assets/database"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
	goproto "google.golang.org/protobuf/proto"
	"os"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/bulk"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	RegisterFeralCatDruid()
	common.RegisterAllEffects()
}

func TestFeralCat(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{{
		Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
		GearSet:     core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "default"),
		Talents:     DefaultTalents,
		SpecOptions: core.SpecOptionsCombo{Label: "Standard", SpecOptions: DefaultSpecOptions},
		OtherSpecOptions: []core.SpecOptionsCombo{
			{Label: "BleedAssumed", SpecOptions: &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{
				Options: &proto.FeralCatDruid_Options{AssumeBleedActive: true},
			}}},
			{Label: "PrepullClearcasting", SpecOptions: &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{
				Options: &proto.FeralCatDruid_Options{PrepullClearcasting: true},
			}}},
		},
		OtherTalentSets: []core.TalentsCombo{
			{Label: "Bearweave", Talents: BearweaveTalents},
		},
		Rotation: core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "default"),
		OtherRotations: []core.RotationCombo{
			core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "bearweave"),
			core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "shadowmeld"),
		},
		Consumables: &proto.ConsumesSpec{},
		ItemFilter: core.ItemFilter{
			ArmorType: proto.ArmorType_ArmorTypeLeather,
			WeaponTypes: []proto.WeaponType{
				proto.WeaponType_WeaponTypeDagger, proto.WeaponType_WeaponTypeFist,
				proto.WeaponType_WeaponTypeMace, proto.WeaponType_WeaponTypeStaff,
			},
			RangedWeaponTypes: []proto.RangedWeaponType{proto.RangedWeaponType_RangedWeaponTypeIdol},
		},
	}}))
}
func BenchmarkSimulate(b *testing.B) {
	rsr := &proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Class:         proto.Class_ClassDruid,
				Race:          proto.Race_RaceNightElf,
				TalentsString: DefaultTalents,
				Equipment:     core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "default").GearSet,
				Consumables:   DefaultConsumables,
				Spec:          DefaultSpecOptions,
				Rotation:      core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "default").Rotation,
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

// The cat UI's Monocat and Bearweave presets (presets.ts CatTalents and BearweaveTalents).
const DefaultTalents = "050022-55000032021021212051-055"
const BearweaveTalents = "050022-55000032021111212051-055"

var DefaultSpecOptions = &proto.Player_FeralCatDruid{
	FeralCatDruid: &proto.FeralCatDruid{
		Options: &proto.FeralCatDruid_Options{},
	},
}

var DefaultConsumables = &proto.ConsumesSpec{
	PotId:      13444,                 // Major Mana Potion
	ElixirIds:  []int32{13452, 13447}, // Elixir of the Mongoose, Elixir of the Sages
	FoodId:     250072,                // Swiftstrike Steak
	MhImbueId:  12643,                 // Dense Weightstone
	ConjuredId: 7676,                  // Thistle Tea
	// Goblin Sapper Charge, Scroll of Strength IV. Scroll of Agility IV is left out, as in the UI:
	// it shares the Agility group with Elixir of the Mongoose, which comes first.
	ExtraIds: []int32{10646, 10310},
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
	settings.Player.Rotation = core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "default").Rotation
	raid := core.SinglePlayerRaidProto(settings.Player, settings.PartyBuffs, settings.RaidBuffs, settings.Debuffs)
	raid.Tanks = settings.Tanks
	if settings.Encounter == nil {
		t.Fatal("build fixture has no encounter")
	}

	return &proto.RaidSimRequest{Raid: raid, Encounter: settings.Encounter, SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1, IsTest: true}}
}

func starterRequest(t testing.TB) *proto.RaidSimRequest {
	return buildRequest(t, "starter")
}

func TestCatBuildsUseNativeRaidTargetArmor(t *testing.T) {
	if !core.WITH_DB {
		t.Skip("requires the committed game database")
	}
	for _, name := range []string{"starter", "raid"} {
		t.Run(name, func(t *testing.T) {
			request := buildRequest(t, name)
			target := request.Encounter.Targets[0]
			defaultTarget := core.FreshDefaultTargetConfig()
			if target.Level != core.CharacterLevel+3 || target.Stats[stats.Armor] != defaultTarget.Stats[stats.Armor] || target.MinBaseDamage != defaultTarget.MinBaseDamage {
				t.Fatalf("Cat build overrides native raid target level, armor, or attack damage: %v", target)
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

func TestCatBuildGearMatchesFaction(t *testing.T) {
	if !core.WITH_DB {
		t.Skip("requires the committed game database")
	}
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

func TestRaidGearFixtureUsesEquippableDruidItems(t *testing.T) {
	if !core.WITH_DB {
		t.Skip("requires the committed game database")
	}
	request := buildRequest(t, "raid")
	gear := request.Raid.Parties[0].Players[0].GetEquipment()
	if gear == nil {
		t.Fatal("raid build has no equipment")
	}
	if len(gear.Items) != int(proto.ItemSlot_ItemSlotRanged)+1 {
		t.Fatalf("gear has %d slots, want %d", len(gear.Items), int(proto.ItemSlot_ItemSlotRanged)+1)
	}
	for slot, spec := range gear.Items {
		if spec.Id == 0 {
			continue
		}
		item := core.GetItemByID(spec.Id)
		if item == nil {
			t.Errorf("slot %d item %d is absent from the generated database", slot, spec.Id)
			continue
		}
		if !item.UsableBy(proto.Class_ClassDruid) {
			t.Errorf("slot %d item %d is not usable by Druids", slot, spec.Id)
		}
	}
	for _, id := range []int32{8345, 272427} {
		found := false
		for _, item := range gear.Items {
			found = found || item.Id == id
		}
		if !found {
			t.Errorf("supported Cat item %d is absent", id)
		}
	}

	sim := core.NewSim(request, simsignals.CreateSignals())
	sim.Reset()
	cat := sim.Raid.Parties[0].Players[0].GetCharacter()
	shift := cat.GetSpell(core.ActionID{SpellID: 1322605})
	if shift == nil {
		t.Fatal("raid fixture has no Shifting Power")
	}
	cat.AddMana(sim, cat.MaxMana(), cat.NewManaMetrics(core.ActionID{SpellID: 1}))
	cat.SpendEnergy(sim, cat.CurrentEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	if !shift.Cast(sim, cat.CurrentTarget) {
		t.Fatal("raid fixture could not cast Shifting Power")
	}
	if got := cat.CurrentEnergy(); got != 45 {
		t.Errorf("raid fixture Shifting Power energy = %v, want 45", got)
	}
	// Improved Shifting Power is eight seconds, less the Howling Idol one-second reduction.
	if got := shift.ReadyAt(); got != 7*time.Second {
		t.Errorf("raid fixture Shifting Power ready at %v, want 7s", got)
	}
}

func TestFeralCatStarterCasts(t *testing.T) {
	if !core.WITH_DB {
		t.Skip("requires the committed game database")
	}
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

func TestAvailableThistleTeaRequiresSelection(t *testing.T) {
	if !core.WITH_DB {
		t.Skip("requires the committed game database")
	}
	for _, preset := range []string{"default", "bearweave"} {
		for _, selected := range []int32{0, 7676} {
			t.Run(fmt.Sprintf("%s/selected=%d", preset, selected), func(t *testing.T) {
				request := buildRequest(t, "raid")
				player := request.Raid.Parties[0].Players[0]
				player.Rotation = core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", preset).Rotation
				player.Consumables.ConjuredId = selected
				player.Consumables.ConjuredItems = []int32{7676}
				result := core.RunRaidSim(request)
				if result.Error != nil {
					t.Fatal(result.Error.Message)
				}
				casts := int32(0)
				for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
					if action.Id.GetItemId() == 7676 {
						for _, target := range action.Targets {
							casts += target.Casts
						}
					}
				}
				if selected == 0 && casts != 0 {
					t.Fatalf("unselected Tea casts = %d, want 0", casts)
				}
				if selected == 7676 && casts == 0 {
					t.Fatal("selected Tea was never used")
				}
			})
		}
	}
}

func TestCatRaidBuildsHaveBulkGearCombinations(t *testing.T) {
	if !core.WITH_DB {
		t.Skip("requires the committed game database")
	}
	for _, name := range []string{"starter", "raid"} {
		t.Run(name, func(t *testing.T) {
			result := bulk.BulkCombinationCount(&proto.BulkCombinationCountRequest{
				BaseRequest: buildRequest(t, name), BulkSettings: &proto.BulkSettings{},
			})
			if result.Error != nil {
				t.Fatal(result.Error.Message)
			}
			if result.Combinations == 0 {
				t.Fatal("raid build has no equippable bulk combination")
			}
		})
	}
}
