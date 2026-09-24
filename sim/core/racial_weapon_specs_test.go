package core

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

const (
	racialTestSwordID int32 = 991511 + iota
	racialTestMaceID
	racialTestDaggerID
	racialTestTwoHandAxeID
	racialTestTwoHandMaceID
)

func addRacialTestWeapons() {
	weapon := func(id int32, weaponType proto.WeaponType, handType proto.HandType) *proto.SimItem {
		return &proto.SimItem{
			Id:             id,
			Name:           "Test Weapon",
			Type:           proto.ItemType_ItemTypeWeapon,
			WeaponType:     weaponType,
			HandType:       handType,
			WeaponSpeed:    2.6,
			ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {WeaponDamageMin: 50, WeaponDamageMax: 90}},
		}
	}
	addToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{
		weapon(racialTestSwordID, proto.WeaponType_WeaponTypeSword, proto.HandType_HandTypeOneHand),
		weapon(racialTestMaceID, proto.WeaponType_WeaponTypeMace, proto.HandType_HandTypeOneHand),
		weapon(racialTestDaggerID, proto.WeaponType_WeaponTypeDagger, proto.HandType_HandTypeOneHand),
		weapon(racialTestTwoHandAxeID, proto.WeaponType_WeaponTypeAxe, proto.HandType_HandTypeTwoHand),
		weapon(racialTestTwoHandMaceID, proto.WeaponType_WeaponTypeMace, proto.HandType_HandTypeTwoHand),
	}})
}

func racialCritWithMainHand(race proto.Race, itemID int32) (physical float64, spell float64) {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: itemID}

	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 100},
		Raid: &proto.Raid{
			Parties: []*proto.Party{{
				Players: []*proto.Player{{
					Name:      "Racial Tester",
					Class:     proto.Class_ClassWarrior,
					Race:      race,
					Buffs:     &proto.IndividualBuffs{},
					Spec:      &proto.Player_DpsWarrior{},
					Equipment: &proto.EquipmentSpec{Items: items},
					Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
				}},
				Buffs: &proto.PartyBuffs{},
			}},
		},
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 63, MobType: proto.MobType_MobTypeHumanoid}},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()

	character := sim.Raid.Parties[0].Players[0].GetCharacter()
	return character.GetStat(stats.PhysicalCritPercent), character.GetStat(stats.SpellCritPercent)
}

func TestRacialWeaponSpecializationsRaiseAllCrit(t *testing.T) {
	addRacialTestWeapons()

	for _, tc := range []struct {
		name            string
		race            proto.Race
		weapon, without int32
		crit            float64
	}{
		{"human with a sword", proto.Race_RaceHuman, racialTestSwordID, racialTestDaggerID, 2},
		{"human with a mace", proto.Race_RaceHuman, racialTestMaceID, racialTestDaggerID, 0},
		{"orc with a two-handed axe", proto.Race_RaceOrc, racialTestTwoHandAxeID, racialTestTwoHandMaceID, 1},
		{"dwarf with a mace", proto.Race_RaceDwarf, racialTestMaceID, racialTestDaggerID, 1},
		{"dwarf with a two-handed mace", proto.Race_RaceDwarf, racialTestTwoHandMaceID, racialTestTwoHandAxeID, 1},
	} {
		physical, spell := racialCritWithMainHand(tc.race, tc.weapon)
		basePhysical, baseSpell := racialCritWithMainHand(tc.race, tc.without)
		if got := physical - basePhysical; math.Abs(got-tc.crit) > 1e-9 {
			t.Errorf("%s: physical crit +%v%%, want +%v%%", tc.name, got, tc.crit)
		}
		if got := spell - baseSpell; math.Abs(got-tc.crit) > 1e-9 {
			t.Errorf("%s: spell crit +%v%%, want +%v%%", tc.name, got, tc.crit)
		}
	}
}
