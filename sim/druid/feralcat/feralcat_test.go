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
		Rotation:    core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "default"),
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

const DefaultTalents = "05002-55000032021032212051-055"
const MonocatTalents = "-55000032021032212051-05503"

var DefaultSpecOptions = &proto.Player_FeralCatDruid{
	FeralCatDruid: &proto.FeralCatDruid{
		Rotation: &proto.FeralCatDruid_Rotation{
			FinishingMove:      proto.FeralCatDruid_Rotation_Rip,
			RipMinComboPoints:  5,
			BiteMinComboPoints: 5,
			RakeTrick:          true,
		},
		Options: &proto.FeralCatDruid_Options{},
	},
}

var DefaultConsumables = &proto.ConsumesSpec{
	PotId:      250943,                // Major Frenzy Potion
	ElixirIds:  []int32{13452, 13447}, // Elixir of the Mongoose, Elixir of the Sages
	FoodId:     250072,                // Swiftstrike Steak
	MhImbueId:  12643,                 // Dense Weightstone
	ConjuredId: 12662,                 // Demonic Rune
	// Goblin Sapper Charge, Scroll of Agility IV, Scroll of Strength IV
	ExtraIds: []int32{10646, 10309, 10310},
}
