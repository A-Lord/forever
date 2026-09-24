package retribution

import (
	"testing"

	"github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	RegisterRetributionPaladin()
	common.RegisterAllEffects()
}

func TestRetribution(t *testing.T) {
	t.Skip("class talents and abilities are stubbed pending their Forever implementations; " +
		"the golden numbers cannot be meaningful until then")
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassPaladin,
			Race:       proto.Race_RaceUndead,
			OtherRaces: []proto.Race{proto.Race_RaceHuman},
			SpecOptions: core.SpecOptionsCombo{Label: "Default", SpecOptions: &proto.Player_RetributionPaladin{
				RetributionPaladin: &proto.RetributionPaladin{
					Options: &proto.RetributionPaladin_Options{
						ClassOptions: &proto.PaladinOptions{},
					},
				},
			}},
			Consumables: DefaultConsumables,
			Profession1: proto.Profession_Engineering,
			Profession2: proto.Profession_Blacksmithing,
			GearSet:     core.GetGearSet("../../../ui/specs/paladin/retribution/gear_sets", "p1"),
			Talents:     "5-053201-0523005120033125331051",
			Rotation:    core.GetAplRotation("../../../ui/specs/paladin/retribution/apls", "default"),
			ItemFilter: core.ItemFilter{
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypePolearm,
					proto.WeaponType_WeaponTypeSword,
					proto.WeaponType_WeaponTypeAxe,
					proto.WeaponType_WeaponTypeMace,
				},
				ArmorType: proto.ArmorType_ArmorTypePlate,
				HandTypes: []proto.HandType{proto.HandType_HandTypeTwoHand},
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeLibram,
				},
			},
		},
	}))
}

var DefaultConsumables = &proto.ConsumesSpec{
	ConjuredId:   12662,
	GoblinSapper: true,
}
