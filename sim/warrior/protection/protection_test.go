package protection

import (
	"github.com/wowsims/forever/sim/common"
	_ "github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"

	"testing"
)

func init() {
	RegisterProtectionWarrior()
	common.RegisterAllEffects()
}

func TestProtectionWarrior(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassWarrior,
			Race:       proto.Race_RaceOrc,
			OtherRaces: []proto.Race{proto.Race_RaceHuman},
			GearSet:    core.GetGearSet("../../../ui/specs/warrior/protection/gear_sets", "preraid"),
			OtherGearSets: []core.GearSetCombo{
				core.GetGearSet("../../../ui/specs/warrior/protection/gear_sets", "p1_bis"),
			},
			Talents:          DefaultProtectionTalents,
			Consumables:      DefaultConsumables,
			SpecOptions:      core.SpecOptionsCombo{Label: "Protection", SpecOptions: DefaultOptions},
			StartingDistance: 0,
			Profession1:      proto.Profession_Engineering,
			Profession2:      proto.Profession_Blacksmithing,

			Rotation: core.GetAplRotation("../../../ui/specs/warrior/protection/apls", "default"),

			IndividualBuffs: core.FullTankIndividualBuffs,

			IsTank:          true,
			InFrontOfTarget: true,

			ItemFilter: core.ItemFilter{
				ArmorType: proto.ArmorType_ArmorTypePlate,
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeFist,
					proto.WeaponType_WeaponTypeMace,
					proto.WeaponType_WeaponTypeSword,
					proto.WeaponType_WeaponTypeAxe,
					proto.WeaponType_WeaponTypeShield,
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

var DefaultOptions = &proto.Player_ProtectionWarrior{
	ProtectionWarrior: &proto.ProtectionWarrior{
		Options: &proto.ProtectionWarrior_Options{
			ClassOptions: &proto.WarriorOptions{
				StartingRage:  100,
				DefaultShout:  proto.WarriorShout_WarriorShoutBattle,
				DefaultStance: proto.WarriorStance_WarriorStanceDefensive,
			},
		},
	},
}

var DefaultProtectionTalents = "05--552531233311210531"

var DefaultConsumables = &proto.ConsumesSpec{
	PotId:            13455, // Greater Stoneshield Potion
	FoodId:           21023, // Dirge's Kickin' Chimaerok Chops
	ConjuredId:       9421,
	ExplosiveId:      18588,
	OhImbueId:        12404,
	BattleElixirId:   13452, // Elixir of the Mongoose
	GuardianElixirId: 9088,
	// Goblin Sapper Charge, Scroll of Agility IV, Scroll of Strength IV, Scroll of Protection IV
	ExtraIds: []int32{10646, 10309, 10310, 10305},
}
