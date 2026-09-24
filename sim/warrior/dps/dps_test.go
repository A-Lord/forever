package dps

import (
	"github.com/wowsims/forever/sim/common"
	_ "github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"

	"testing"
)

func init() {
	RegisterDpsWarrior()
	common.RegisterAllEffects()
}

func TestDpsWarrior(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassWarrior,
			Race:       proto.Race_RaceOrc,
			OtherRaces: []proto.Race{proto.Race_RaceHuman},
			GearSet:    core.GetGearSet("../../../ui/specs/warrior/dps/gear_sets", "p1_fury"),
			OtherGearSets: []core.GearSetCombo{
				core.GetGearSet("../../../ui/specs/warrior/dps/gear_sets", "p1_arms"),
				core.GetGearSet("../../../ui/specs/warrior/dps/gear_sets", "p2_fury"),
				core.GetGearSet("../../../ui/specs/warrior/dps/gear_sets", "p2_arms"),
				core.GetGearSet("../../../ui/specs/warrior/dps/gear_sets", "p3_fury"),
				core.GetGearSet("../../../ui/specs/warrior/dps/gear_sets", "p3_arms"),
				core.GetGearSet("../../../ui/specs/warrior/dps/gear_sets", "p4_fury"),
				core.GetGearSet("../../../ui/specs/warrior/dps/gear_sets", "p4_arms"),
				core.GetGearSet("../../../ui/specs/warrior/dps/gear_sets", "p5_fury"),
				core.GetGearSet("../../../ui/specs/warrior/dps/gear_sets", "p5_arms"),
			},
			Talents: DefaultFuryTalents,
			OtherTalentSets: []core.TalentsCombo{
				{Label: "Arms", Talents: DefaultArmsTalents},
			},
			Consumables:      DefaultConsumables,
			SpecOptions:      core.SpecOptionsCombo{Label: "Fury", SpecOptions: DefaultOptions},
			StartingDistance: 0,
			Profession1:      proto.Profession_Engineering,
			Profession2:      proto.Profession_Blacksmithing,

			Rotation: core.GetAplRotation("../../../ui/specs/warrior/dps/apls", "fury"),
			OtherRotations: []core.RotationCombo{
				core.GetAplRotation("../../../ui/specs/warrior/dps/apls", "arms"),
			},

			ItemFilter: core.ItemFilter{
				ArmorType: proto.ArmorType_ArmorTypeLeather,
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeFist,
					proto.WeaponType_WeaponTypeMace,
					proto.WeaponType_WeaponTypeSword,
					proto.WeaponType_WeaponTypeAxe,
				},
				HandTypes: []proto.HandType{
					proto.HandType_HandTypeMainHand,
					proto.HandType_HandTypeOffHand,
					proto.HandType_HandTypeOneHand,
					proto.HandType_HandTypeTwoHand,
				},
			},
		},
	}))
}

var DefaultOptions = &proto.Player_DpsWarrior{
	DpsWarrior: &proto.DpsWarrior{
		Options: &proto.DpsWarrior_Options{
			ClassOptions: &proto.WarriorOptions{
				UseBattleShout: true,
				DefaultStance:  proto.WarriorStance_WarriorStanceBerserker,
			},
		},
	},
}

var DefaultFuryTalents = "30305003-050530205052310051"
var DefaultArmsTalents = "30305213032115201-05052030004"

var DefaultConsumables = &proto.ConsumesSpec{}
