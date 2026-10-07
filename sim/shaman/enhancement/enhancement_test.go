package enhancement

import (
	"testing"

	"github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	RegisterEnhancementShaman()
	common.RegisterAllEffects()
}

func TestEnhancement(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:       proto.Class_ClassShaman,
			Race:        proto.Race_RaceOrc,
			OtherRaces:  []proto.Race{proto.Race_RaceTroll},
			SpecOptions: core.SpecOptionsCombo{Label: "Windfury", SpecOptions: imbueOptions(proto.ShamanImbue_WindfuryWeapon)},
			OtherSpecOptions: []core.SpecOptionsCombo{
				{Label: "Flametongue", SpecOptions: imbueOptions(proto.ShamanImbue_FlametongueWeapon)},
				{Label: "Frostbrand", SpecOptions: imbueOptions(proto.ShamanImbue_FrostbrandWeapon)},
				{Label: "Rockbiter", SpecOptions: imbueOptions(proto.ShamanImbue_RockbiterWeapon)},
			},
			GearSet:     core.GetGearSet("../../../ui/specs/shaman/enhancement/gear_sets", "preraid"),
			Talents:     DefaultTalents,
			Consumables: DefaultConsumables,
			Profession1: proto.Profession_Engineering,
			Profession2: proto.Profession_Leatherworking,
			Rotation:    core.GetAplRotation("../../../ui/specs/shaman/enhancement/apls", "default"),
			ItemFilter: core.ItemFilter{
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeAxe,
					proto.WeaponType_WeaponTypeMace,
				},
				HandTypes: []proto.HandType{proto.HandType_HandTypeTwoHand},
				ArmorType: proto.ArmorType_ArmorTypeMail,
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeTotem,
				},
			},
		},
	}))
}

func imbueOptions(imbue proto.ShamanImbue) *proto.Player_EnhancementShaman {
	return &proto.Player_EnhancementShaman{
		EnhancementShaman: &proto.EnhancementShaman{
			Options: &proto.EnhancementShaman_Options{
				ClassOptions: &proto.ShamanOptions{
					ImbueMh: imbue,
				},
			},
		},
	}
}

const DefaultTalents = "0505331-055030031005112251"

// The UI's Enhancement defaults: Major Frenzy Potion, Flask of Natural Aggression, Demonic Rune, the
// Owl, Sages, Greater Mageblood, Nature Power and Greater Arcane elixirs, Winterfall Firewater,
// Ez-Thro Dynamite II, Goblin Sapper Charge and the Agility and Strength scrolls. On top of them, the
// main hand takes Elemental Sharpening Stone, the best stone for its axe or mace, which sits beside
// the shaman's imbue.
var DefaultConsumables = &proto.ConsumesSpec{
	PotId:       250943,
	FlaskId:     274274,
	ConjuredId:  12662,
	ElixirIds:   []int32{250337, 250338, 250341, 250343, 13454, 12820},
	ExplosiveId: 18588,
	ExtraIds:    []int32{10646, 10309, 10310},
	MhImbueId:   18262,
}
