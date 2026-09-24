package rogue

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	_ "github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	RegisterRogue()
	common.RegisterAllEffects()
}

func TestRogue(t *testing.T) {
	t.Skip("class talents and abilities are stubbed pending their Forever implementations; " +
		"the golden numbers cannot be meaningful until then")
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassRogue,
			Race:       proto.Race_RaceHuman,
			OtherRaces: []proto.Race{proto.Race_RaceOrc},
			GearSet:    core.GetGearSet("../../ui/specs/rogue/dps/gear_sets", "preraid"),
			OtherGearSets: []core.GearSetCombo{
				core.GetGearSet("../../ui/specs/rogue/dps/gear_sets", "p1"),
				//core.GetGearSet("../../../ui/specs/rogue/combat/gear_sets", "p4_combat"),
			},
			Talents:     DefaultTalents,
			Consumables: DefaultConsumables,
			SpecOptions: core.SpecOptionsCombo{Label: "Rogue", SpecOptions: DefaultOptions},

			Rotation:       core.GetAplRotation("../../ui/specs/rogue/dps/apls", "default"),
			OtherRotations: []core.RotationCombo{},
			ItemFilter: core.ItemFilter{
				ArmorType: proto.ArmorType_ArmorTypeLeather,
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeDagger,
					proto.WeaponType_WeaponTypeFist,
					proto.WeaponType_WeaponTypeMace,
					proto.WeaponType_WeaponTypeSword,
				},
				HandTypes: []proto.HandType{
					proto.HandType_HandTypeMainHand,
					proto.HandType_HandTypeOffHand,
					proto.HandType_HandTypeOneHand,
				},
			},
		},
	}))
}

var DefaultOptions = &proto.Player_Rogue{
	Rogue: &proto.Rogue{
		Options: &proto.Rogue_Options{
			ClassOptions: &proto.RogueOptions{},
		},
	},
}

var DefaultTalents = "00532012502-023305200005015002321151"

var DefaultConsumables = &proto.ConsumesSpec{
	ConjuredId: 7676,
}
