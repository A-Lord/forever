package warlock

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
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

// A demon's melee and spell crit are the warlock's spell crit, and follow it when it changes in combat.
func TestDemonCritIsTheWarlocksSpellCrit(t *testing.T) {
	player := &proto.Player{
		Name:      "Demon crit",
		Race:      proto.Race_RaceOrc,
		Class:     proto.Class_ClassWarlock,
		Equipment: &proto.EquipmentSpec{},
		Spec: &proto.Player_Warlock{Warlock: &proto.Warlock{Options: &proto.Warlock_Options{
			ClassOptions: &proto.WarlockOptions{Summon: proto.WarlockOptions_Succubus},
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
	sayaad := warlock.Succubus
	if !sayaad.IsEnabled() {
		t.Fatalf("Sayaad not summoned at the start")
	}

	check := func(when string) {
		spellCrit := warlock.GetStat(stats.SpellCritPercent)
		if spellCrit <= 0 {
			t.Fatalf("%s: warlock has %.2f%% spell crit, want some to inherit", when, spellCrit)
		}
		for _, stat := range []stats.Stat{stats.PhysicalCritPercent, stats.SpellCritPercent} {
			if got := sayaad.GetStat(stat); got != spellCrit {
				t.Errorf("%s: Sayaad %s %.2f%%, want the warlock's spell crit %.2f%%", when, stat.StatName(), got, spellCrit)
			}
		}
	}
	check("at the start")
	warlock.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: 5})
	for end := sim.CurrentTime + core.PetUpdateInterval; sim.CurrentTime <= end && !sim.Step(); {
	}
	check("a pet update after +5% spell crit")
}

// In execute, Decimation's buff (440873) comes with the cast of a Shadow Bolt or Searing Pain whose hit lands,
// before a Shadow Bolt reaches the target; a miss or a full resist grants nothing, and the bolt's impact
// grants nothing either.
func TestDecimationGrantedOnLandedCast(t *testing.T) {
	player := &proto.Player{
		Name:          "Decimation",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarlock,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: "-000000000002",
		Spec:          &proto.Player_Warlock{Warlock: &proto.Warlock{Options: &proto.Warlock_Options{ClassOptions: &proto.WarlockOptions{}}}},
		Rotation:      &proto.APLRotation{},
	}
	newSim := func(resistance float64) (*core.Simulation, *Warlock, *core.Unit) {
		encounter := core.MakeSingleTargetEncounter(0)
		encounter.Duration = 3600
		encounter.ExecuteProportion_90, encounter.ExecuteProportion_45, encounter.ExecuteProportion_35 = 1, 1, 1
		encounter.Targets[0].Stats[stats.FireResistance] = resistance
		encounter.Targets[0].Stats[stats.ShadowResistance] = resistance
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  encounter,
			SimOptions: &proto.SimOptions{RandomSeed: 1},
		}, simsignals.CreateSignals())
		sim.Reset()
		for !sim.IsExecutePhase35() && !sim.Step() {
		}
		warlock := sim.Raid.Parties[0].Players[0].(*Warlock)
		warlock.DistanceFromTarget = 30
		return sim, warlock, sim.Encounter.AllTargetUnits[0]
	}
	stepFor := func(sim *core.Simulation, d time.Duration) {
		for end := sim.CurrentTime + d; sim.CurrentTime < end && !sim.Step(); {
		}
	}

	sim, warlock, target := newSim(0)
	decimation := warlock.GetAuraByID(core.ActionID{SpellID: 440873})
	if warlock.ShadowBolt.TravelTime() <= 2*core.SpellBatchWindow {
		t.Fatalf("Shadow Bolt travels %v, too short to tell its cast from its impact", warlock.ShadowBolt.TravelTime())
	}
	granted, missed := 0, 0
	for i := 0; i < 400; i++ {
		spell := warlock.SearingPain
		if i%2 == 1 {
			spell = warlock.ShadowBolt
		}
		decimation.Deactivate(sim)
		misses := spell.SpellMetrics[target.UnitIndex].Misses
		spell.SkipCastAndApplyEffects(sim, target)
		landed := spell.SpellMetrics[target.UnitIndex].Misses == misses
		if decimation.IsActive() != landed {
			t.Fatalf("%s at %v landed %v and left Decimation up %v, want the buff on a landed cast only",
				spell.ActionID, sim.CurrentTime, landed, decimation.IsActive())
		}
		if landed {
			granted++
		} else {
			missed++
		}
		decimation.Deactivate(sim)
		stepFor(sim, warlock.ShadowBolt.TravelTime()+core.SpellBatchWindow)
		if decimation.IsActive() {
			t.Fatalf("%s's impact at %v granted Decimation, want it granted at the cast only", spell.ActionID, sim.CurrentTime)
		}
	}
	if granted == 0 || missed == 0 {
		t.Fatalf("%d casts granted Decimation and %d missed, want both", granted, missed)
	}

	sim, warlock, target = newSim(1e12)
	decimation = warlock.GetAuraByID(core.ActionID{SpellID: 440873})
	for i := 0; i < 100; i++ {
		spell := warlock.SearingPain
		if i%2 == 1 {
			spell = warlock.ShadowBolt
		}
		spell.SkipCastAndApplyEffects(sim, target)
		stepFor(sim, warlock.ShadowBolt.TravelTime()+core.SpellBatchWindow)
		if decimation.IsActive() {
			t.Fatalf("%s at %v granted Decimation against a target that resists it in full", spell.ActionID, sim.CurrentTime)
		}
	}
	if misses := warlock.SearingPain.SpellMetrics[target.UnitIndex].Misses; misses != 50 {
		t.Fatalf("%d of 50 Searing Pains against the fully resisting target missed, want all resisted in full", misses)
	}
}
