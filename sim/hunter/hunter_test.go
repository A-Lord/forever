package hunter

import (
	"maps"
	"slices"

	"github.com/wowsims/forever/sim/common"
	_ "github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"

	"testing"
)

func init() {
	RegisterHunter()
	common.RegisterAllEffects()
}

func TestHunter(t *testing.T) {
	weaveRotation := core.GetAplRotation("../../ui/specs/hunter/dps/apls", "default")
	weaveRotation.Label = "weave"

	turretRotation := core.GetAplRotation("../../ui/specs/hunter/dps/apls", "default").Rotation
	stopWeaving(turretRotation)

	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassHunter,
			Race:       proto.Race_RaceOrc,
			OtherRaces: []proto.Race{proto.Race_RaceNightElf},
			GearSet:    core.GetGearSet("../../ui/specs/hunter/dps/gear_sets", "p1"),
			Talents:    DefaultMMTalents,
			OtherTalentSets: []core.TalentsCombo{
				{Label: "SV", Talents: DefaultSVTalents},
			},
			Consumables:      DefaultConsumables,
			SpecOptions:      core.SpecOptionsCombo{Label: "Default", SpecOptions: DefaultOptions},
			StartingDistance: 8,
			Profession1:      proto.Profession_Engineering,
			Profession2:      proto.Profession_Blacksmithing,

			Rotation: core.GetAplRotation("../../ui/specs/hunter/dps/apls", "default"),

			ItemFilter: core.ItemFilter{
				ArmorType: proto.ArmorType_ArmorTypeMail,
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeBow,
					proto.RangedWeaponType_RangedWeaponTypeCrossbow,
					proto.RangedWeaponType_RangedWeaponTypeGun,
				},
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeAxe,
					proto.WeaponType_WeaponTypeDagger,
					proto.WeaponType_WeaponTypeFist,
					proto.WeaponType_WeaponTypePolearm,
					proto.WeaponType_WeaponTypeStaff,
					proto.WeaponType_WeaponTypeSword,
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

var DefaultOptions = &proto.Player_Hunter{
	Hunter: &proto.Hunter{
		Options: &proto.Hunter_Options{
			ClassOptions: &proto.HunterOptions{
				Ammo:        proto.HunterOptions_ThoriumHeadedArrow,
				QuiverBonus: proto.HunterOptions_AncientSinewWrappedLamina,
				PetType:     proto.HunterOptions_PetNone,
			},
		},
	},
}

// Every pet family, fully trained, on one race, talent set and rotation: the pet's own suite, so
// that each family's abilities run without multiplying the hunter's. The talents are Beast Mastery,
// the pet's tree, and the default rotation sends a hawk ahead of each Arcane Shot, the two sharing a cooldown.
func TestHunterPets(t *testing.T) {
	turretRotation := core.GetAplRotation("../../ui/specs/hunter/dps/apls", "default")
	turretRotation.Label = "turret"
	stopWeaving(turretRotation.Rotation)

	families := petOptions()
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:            proto.Class_ClassHunter,
			Race:             proto.Race_RaceOrc,
			GearSet:          core.GetGearSet("../../ui/specs/hunter/dps/gear_sets", "p1"),
			Talents:          DefaultBMTalents,
			Consumables:      DefaultConsumables,
			SpecOptions:      families[0],
			OtherSpecOptions: families[1:],
			StartingDistance: 8,
			Rotation:         turretRotation,
			ItemFilter:       core.ItemFilter{ArmorType: proto.ArmorType_ArmorTypeMail},
		},
	}))
}

// One combo per pet family, fully trained.
// The default rotation with its "Melee weave" variable turned off: the hunter stays at range.
func stopWeaving(rotation *proto.APLRotation) {
	for _, variable := range rotation.ValueVariables {
		if variable.Name == "Melee weave" {
			variable.Value = &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: "false"}}}
		}
	}
}

func petOptions() []core.SpecOptionsCombo {
	var combos []core.SpecOptionsCombo
	for _, value := range slices.Sorted(maps.Values(proto.HunterOptions_PetType_value)) {
		petType := proto.HunterOptions_PetType(value)
		if petType == proto.HunterOptions_PetNone {
			continue
		}
		combos = append(combos, core.SpecOptionsCombo{
			Label: petType.String(),
			SpecOptions: &proto.Player_Hunter{
				Hunter: &proto.Hunter{
					Options: &proto.Hunter_Options{
						ClassOptions: &proto.HunterOptions{
							PetType:        petType,
							PetUptime:      1,
							CobraReflexes:  true,
							PetAggression:  5,
							PetAttackSpeed: proto.HunterOptions_FasterAttackII,
						},
					},
				},
			},
		})
	}
	return combos
}

var DefaultBMTalents = "5320001505101251-3050552"
var DefaultMMTalents = "-30535525115023051-5"
var DefaultSVTalents = "-00505505-500250030050220151"

var DefaultConsumables = &proto.ConsumesSpec{
	ElixirIds:   []int32{13452, 250341}, // Elixir of the Mongoose, Greater Mageblood Elixir
	FoodId:      250072,                 // Swiftstrike Steak
	PotId:       250943,                 // Major Frenzy Potion
	ConjuredId:  12662,                  // Demonic Rune
	ExplosiveId: 18588,                  // Ez-Thro Dynamite II
	// Goblin Sapper Charge, Scroll of Agility IV, Scroll of Strength IV
	ExtraIds: []int32{10646, 10309, 10310},
}
