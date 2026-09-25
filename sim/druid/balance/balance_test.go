package balance

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	_ "github.com/wowsims/forever/sim/common" // imported to get caster sets included. (we use spellfire here)
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	RegisterBalanceDruid()
	common.RegisterAllEffects()
}

func TestBalance(t *testing.T) {
	t.Skip("class talents and abilities are stubbed pending their Forever implementations; " +
		"the expected .results numbers cannot be meaningful until then")
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassDruid,
			Race:       proto.Race_RaceNightElf,
			OtherRaces: []proto.Race{proto.Race_RaceTauren},
			SpecOptions: core.SpecOptionsCombo{Label: "Standard", SpecOptions: &proto.Player_BalanceDruid{
				BalanceDruid: &proto.BalanceDruid{
					Options: &proto.BalanceDruid_Options{
						ClassOptions: &proto.DruidOptions{},
					},
				},
			}},
			GearSet: core.GetGearSet("../../../ui/specs/druid/balance/gear_sets", "p1_a"),
			OtherGearSets: []core.GearSetCombo{
				core.GetGearSet("../../../ui/specs/druid/balance/gear_sets", "p2_a"),
				core.GetGearSet("../../../ui/specs/druid/balance/gear_sets", "p3"),
				core.GetGearSet("../../../ui/specs/druid/balance/gear_sets", "p4"),
				core.GetGearSet("../../../ui/specs/druid/balance/gear_sets", "p5"),
			},
			Talents:  DefaultTalents,
			Rotation: core.GetAplRotation("../../../ui/specs/druid/balance/apls", "default"),
			ItemFilter: core.ItemFilter{
				WeaponTypes:       DefaultWeaponTypes,
				ArmorType:         DefaultArmorType,
				RangedWeaponTypes: DefaultRangedWeaponTypes,
			},
		},
	}))
}

const DefaultTalents = "510022312503135231351--520033"

const DefaultArmorType = proto.ArmorType_ArmorTypeLeather

var DefaultWeaponTypes = []proto.WeaponType{
	proto.WeaponType_WeaponTypeDagger,
	proto.WeaponType_WeaponTypeMace,
	proto.WeaponType_WeaponTypeStaff,
	proto.WeaponType_WeaponTypeOffHand,
}

var DefaultRangedWeaponTypes = []proto.RangedWeaponType{
	proto.RangedWeaponType_RangedWeaponTypeIdol,
}
