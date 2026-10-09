package warlock

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

func init() {
	RegisterWarlock()
	common.RegisterAllEffects()
}

func TestAffliction(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassWarlock,
			Race:       proto.Race_RaceOrc,
			OtherRaces: []proto.Race{proto.Race_RaceHuman},
			SpecOptions: core.SpecOptionsCombo{Label: "Affliction", SpecOptions: &proto.Player_Warlock{
				Warlock: &proto.Warlock{
					Options: &proto.Warlock_Options{
						ClassOptions: &proto.WarlockOptions{
							Summon:       proto.WarlockOptions_Imp,
							Armor:        proto.WarlockOptions_DemonArmor,
							CurseOptions: proto.WarlockOptions_Elements,
							SoulShards:   32,
						},
					},
				},
			}},
			GearSet:  core.GetGearSet("../../ui/specs/warlock/dps/gear_sets", "preraid"),
			Talents:  "25550320135201351--053",
			Rotation: core.GetAplRotation("../../ui/specs/warlock/dps/apls", "affliction"),
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
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassWarlock,
			Race:       proto.Race_RaceOrc,
			OtherRaces: []proto.Race{proto.Race_RaceGnome},
			SpecOptions: core.SpecOptionsCombo{Label: "Destruction", SpecOptions: &proto.Player_Warlock{
				Warlock: &proto.Warlock{
					Options: &proto.Warlock_Options{
						ClassOptions: &proto.WarlockOptions{
							SacrificedDemon: proto.WarlockOptions_Imp,
							Armor:           proto.WarlockOptions_DemonArmor,
							CurseOptions:    proto.WarlockOptions_Elements,
							SoulShards:      32,
						},
					},
				},
			}},
			GearSet:  core.GetGearSet("../../ui/specs/warlock/dps/gear_sets", "preraid"),
			Talents:  "03-0050203001-0550355103100351",
			Rotation: core.GetAplRotation("../../ui/specs/warlock/dps/apls", "destruction"),
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

// Soul Fire takes the Soul Shard its row states, and none while Decimation's buff (440873) is up.
func TestSoulFireShardAndDecimation(t *testing.T) {
	player := &proto.Player{
		Name:          "Decimation",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarlock,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: "-000000000002",
		Spec: &proto.Player_Warlock{Warlock: &proto.Warlock{Options: &proto.Warlock_Options{
			ClassOptions: &proto.WarlockOptions{SoulShards: 1},
		}}},
		Rotation: &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	warlock := sim.Raid.Parties[0].Players[0].(*Warlock)
	target := sim.Encounter.AllTargetUnits[0]
	soulFire := warlock.Soulfire
	decimation := warlock.GetAuraByID(core.ActionID{SpellID: 440873})
	if decimation == nil {
		t.Fatalf("Decimation 2/2 registers no buff 440873")
	}

	if !soulFire.ExtraCastCondition(sim, target) {
		t.Fatalf("Soul Fire cannot be cast with 1 Soul Shard")
	}
	soulFire.SkipCastAndApplyEffects(sim, target)
	if got := warlock.SoulShardsAura.GetStacks(); got != 0 {
		t.Fatalf("%d Soul Shards left after Soul Fire, want 0", got)
	}
	if soulFire.ExtraCastCondition(sim, target) {
		t.Errorf("Soul Fire can be cast with no Soul Shard and no Decimation")
	}

	decimation.Activate(sim)
	if !soulFire.ExtraCastCondition(sim, target) {
		t.Errorf("Soul Fire cannot be cast with no Soul Shard under Decimation")
	}
	soulFire.SkipCastAndApplyEffects(sim, target)
	if got := warlock.SoulShardsAura.GetStacks(); got != 0 {
		t.Errorf("%d Soul Shards after Soul Fire under Decimation, want 0 and none spent", got)
	}
}
