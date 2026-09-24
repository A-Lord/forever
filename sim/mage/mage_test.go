package mage

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	_ "github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	RegisterMage()
	common.RegisterAllEffects()
}

func TestArcane(t *testing.T) {
	t.Skip("class talents and abilities are stubbed pending their Forever implementations; " +
		"the golden numbers cannot be meaningful until then")
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassMage,
			Race:       proto.Race_RaceTroll,
			OtherRaces: []proto.Race{proto.Race_RaceOrc},
			SpecOptions: core.SpecOptionsCombo{Label: "Arcane", SpecOptions: &proto.Player_Mage{
				Mage: &proto.Mage{
					Options: &proto.Mage_Options{
						ClassOptions: &proto.MageOptions{
							DefaultMageArmor: proto.MageArmor_MageArmorMageArmor,
						},
					},
				},
			}},
			GearSet:  core.GetGearSet("../../ui/specs/mage/dps/gear_sets", "p1Arcane"),
			Talents:  "2500052300030150330125--053500031003001",
			Rotation: core.GetAplRotation("../../ui/specs/mage/dps/apls", "default"),
			ItemFilter: core.ItemFilter{
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeDagger,
					proto.WeaponType_WeaponTypeSword,
					proto.WeaponType_WeaponTypeOffHand,
					proto.WeaponType_WeaponTypeStaff,
				},
				ArmorType: proto.ArmorType_ArmorTypeCloth,
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeWand,
				},
				EnchantBlacklist: []int32{2673, 3225, 3273},
			},
		},
	}))
}
