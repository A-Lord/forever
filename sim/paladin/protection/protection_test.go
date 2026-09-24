package protection

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	_ "github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	RegisterProtectionPaladin()
	common.RegisterAllEffects()
}

func setValueVariable(apl *proto.APLRotation, name string, val string) {
	for i, v := range apl.ValueVariables {
		if v.Name == name {
			apl.ValueVariables[i].Value = &proto.APLValue{
				Value: &proto.APLValue_Const{
					Const: &proto.APLValueConst{
						Val: val,
					},
				},
			}
			return
		}
	}

	panic("value variable " + name + " not found, APL probably changed, fix tests!")
}

func TestProtection(t *testing.T) {
	t.Skip("class talents and abilities are stubbed pending their Forever implementations; " +
		"the golden numbers cannot be meaningful until then")
	// Set all boolean options to true to test everything
	apl := core.GetAplRotation("../../../ui/specs/paladin/protection/apls", "default")
	setValueVariable(apl.Rotation, "Prioritize Holy Shield", "true")
	setValueVariable(apl.Rotation, "Use Exorcism", "true")
	setValueVariable(apl.Rotation, "Use Avenger's Shield", "true")
	setValueVariable(apl.Rotation, "Use Hammer of Wrath", "true")

	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:            proto.Class_ClassPaladin,
			Race:             proto.Race_RaceUndead,
			OtherRaces:       []proto.Race{proto.Race_RaceHuman},
			GearSet:          core.GetGearSet("../../../ui/specs/paladin/protection/gear_sets", "p2"),
			Talents:          DefaultProtectionTalents,
			Consumables:      DefaultConsumables,
			SpecOptions:      core.SpecOptionsCombo{Label: "Protection", SpecOptions: DefaultOptions},
			StartingDistance: 5,
			Profession1:      proto.Profession_Engineering,
			Profession2:      proto.Profession_Enchanting,

			Rotation: apl,

			IndividualBuffs: core.FullTankIndividualBuffs,

			IsTank:          true,
			InFrontOfTarget: true,

			ItemFilter: core.ItemFilter{
				ArmorType: proto.ArmorType_ArmorTypePlate,
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeAxe,
					proto.WeaponType_WeaponTypeSword,
					proto.WeaponType_WeaponTypeMace,
					proto.WeaponType_WeaponTypeShield,
				},
				HandTypes: []proto.HandType{
					proto.HandType_HandTypeMainHand,
					proto.HandType_HandTypeOffHand,
					proto.HandType_HandTypeOneHand,
				},
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeLibram,
				},
			},
		},
	}))
}

var DefaultOptions = &proto.Player_ProtectionPaladin{
	ProtectionPaladin: &proto.ProtectionPaladin{
		Options: &proto.ProtectionPaladin_Options{
			ClassOptions: &proto.PaladinOptions{},
		},
	},
}

var DefaultProtectionTalents = "-0530513050000142521051-052050003003"

var DefaultConsumables = &proto.ConsumesSpec{
	ConjuredId: 12662, // Dark Rune
}
