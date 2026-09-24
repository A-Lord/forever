package warlock

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	RegisterWarlock()
	common.RegisterAllEffects()
}

func TestAffliction(t *testing.T) {
	t.Skip("class talents and abilities are stubbed pending their Forever implementations; " +
		"the golden numbers cannot be meaningful until then")
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassWarlock,
			Race:       proto.Race_RaceOrc,
			OtherRaces: []proto.Race{proto.Race_RaceHuman},
			SpecOptions: core.SpecOptionsCombo{Label: "Affliction", SpecOptions: &proto.Player_Warlock{
				Warlock: &proto.Warlock{
					Options: &proto.Warlock_Options{
						ClassOptions: &proto.WarlockOptions{
							Summon:          proto.WarlockOptions_Imp,
							SacrificeSummon: false,
							Armor:           proto.WarlockOptions_FelArmor,
							CurseOptions:    proto.WarlockOptions_Elements,
						},
					},
				},
			}},
			GearSet:  core.GetGearSet("../../ui/specs/warlock/dps/gear_sets", "preraid"),
			Talents:  "05022221112351055003--50500051220001",
			Rotation: core.GetAplRotation("../../ui/specs/warlock/dps/apls", "default"),
			ItemFilter: core.ItemFilter{
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeDagger,
					proto.WeaponType_WeaponTypeStaff,
					proto.WeaponType_WeaponTypeSword,
				},
				ArmorType: proto.ArmorType_ArmorTypeCloth,
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeWand,
				},
				EnchantBlacklist: []int32{2673, 3225, 3273},
				IDBlacklist:      []int32{28556},
			},
		},
	}))
}

func TestDestruction(t *testing.T) {
	t.Skip("class talents and abilities are stubbed pending their Forever implementations; " +
		"the golden numbers cannot be meaningful until then")
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassWarlock,
			Race:       proto.Race_RaceOrc,
			OtherRaces: []proto.Race{proto.Race_RaceGnome},
			SpecOptions: core.SpecOptionsCombo{Label: "Destruction", SpecOptions: &proto.Player_Warlock{
				Warlock: &proto.Warlock{
					Options: &proto.Warlock_Options{
						ClassOptions: &proto.WarlockOptions{
							Summon:          proto.WarlockOptions_Succubus,
							SacrificeSummon: true,
							Armor:           proto.WarlockOptions_FelArmor,
							CurseOptions:    proto.WarlockOptions_Agony,
						},
					},
				},
			}},
			GearSet:  core.GetGearSet("../../ui/specs/warlock/dps/gear_sets", "preraid"),
			Talents:  "-20500301332101-50500051220051053105",
			Rotation: core.GetAplRotation("../../ui/specs/warlock/dps/apls", "default"),
			ItemFilter: core.ItemFilter{
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeDagger,
					proto.WeaponType_WeaponTypeStaff,
					proto.WeaponType_WeaponTypeSword,
				},
				ArmorType: proto.ArmorType_ArmorTypeCloth,
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeWand,
				},
				EnchantBlacklist: []int32{2673, 3225, 3273},
				IDBlacklist:      []int32{28556},
			},
		},
	}))
}
