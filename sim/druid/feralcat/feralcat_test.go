package feralcat

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
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
