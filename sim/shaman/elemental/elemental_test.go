package elemental

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	RegisterElementalShaman()
	common.RegisterAllEffects()
}

func TestElemental(t *testing.T) {
	t.Skip("class talents and abilities are stubbed pending their Forever implementations; " +
		"the golden numbers cannot be meaningful until then")
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassShaman,
			Race:       proto.Race_RaceTroll,
			OtherRaces: []proto.Race{proto.Race_RaceOrc, proto.Race_RaceDwarf},
			SpecOptions: core.SpecOptionsCombo{Label: "Standard", SpecOptions: &proto.Player_ElementalShaman{
				ElementalShaman: &proto.ElementalShaman{
					Options: &proto.ElementalShaman_Options{
						ClassOptions: &proto.ShamanOptions{
							ShieldProcrate: 0.0,
						},
					},
				},
			}},
			GearSet: core.GetGearSet("../../../ui/specs/shaman/elemental/gear_sets", "p1_a"),
			OtherGearSets: []core.GearSetCombo{
				core.GetGearSet("../../../ui/specs/shaman/elemental/gear_sets", "p2"),
				core.GetGearSet("../../../ui/specs/shaman/elemental/gear_sets", "p3"),
				core.GetGearSet("../../../ui/specs/shaman/elemental/gear_sets", "p4"),
				core.GetGearSet("../../../ui/specs/shaman/elemental/gear_sets", "p5"),
			},
			Talents:  DefaultTalents,
			Rotation: core.GetAplRotation("../../../ui/specs/shaman/elemental/apls", "default"),
			ItemFilter: core.ItemFilter{
				WeaponTypes:       DefaultWeaponTypes,
				ArmorType:         DefaultArmorType,
				RangedWeaponTypes: DefaultRangedWeaponTypes,
			},
		},
	}))
}

const DefaultTalents = "55003105100213351051--05105301005"

const DefaultArmorType = proto.ArmorType_ArmorTypeMail

var DefaultWeaponTypes = []proto.WeaponType{
	proto.WeaponType_WeaponTypeAxe,
	proto.WeaponType_WeaponTypeDagger,
	proto.WeaponType_WeaponTypeFist,
	proto.WeaponType_WeaponTypeMace,
	proto.WeaponType_WeaponTypeStaff,
	proto.WeaponType_WeaponTypeShield,
}

var DefaultRangedWeaponTypes = []proto.RangedWeaponType{
	proto.RangedWeaponType_RangedWeaponTypeTotem,
}
