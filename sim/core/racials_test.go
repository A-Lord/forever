package core

import (
	"math"
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

func setupRacialSim(race proto.Race) (*Simulation, *FakeRageWarrior) {
	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{
			RandomSeed: 100,
		},
		Raid: &proto.Raid{
			Parties: []*proto.Party{
				{
					Players: []*proto.Player{
						{
							Name:      "Warrior",
							Class:     proto.Class_ClassWarrior,
							Race:      race,
							Buffs:     &proto.IndividualBuffs{},
							Spec:      &proto.Player_DpsWarrior{},
							Equipment: &proto.EquipmentSpec{},
							Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
						},
					},
					Buffs: &proto.PartyBuffs{},
				},
			},
		},
		Encounter: &proto.Encounter{
			Targets: []*proto.Target{
				{Name: "target", Level: 63, MobType: proto.MobType_MobTypeElemental},
			},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()

	return sim, sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
}

func castRacial(t *testing.T, sim *Simulation, fw *FakeRageWarrior, spellID int32) {
	t.Helper()
	spell := fw.GetSpell(ActionID{SpellID: spellID})
	if spell == nil {
		t.Fatalf("Racial %d is not registered", spellID)
	}
	if !spell.Cast(sim, fw.CurrentTarget) {
		t.Fatalf("Racial %d could not be cast", spellID)
	}
}

func TestPlayableRaces(t *testing.T) {
	for class, races := range ClassRaceCapabilities {
		for _, race := range races {
			if race == proto.Race_RaceBloodElf {
				t.Errorf("%s is not playable, but %s lists it", race, class)
			}
			if _, ok := BaseStats[BaseStatsKey{Race: race, Class: class}]; !ok {
				t.Errorf("No base stats for %s %s", race, class)
			}
		}
	}

	for class, race := range map[proto.Class]proto.Race{
		proto.Class_ClassPaladin: proto.Race_RaceUndead,
		proto.Class_ClassShaman:  proto.Race_RaceDwarf,
		proto.Class_ClassHunter:  proto.Race_RaceHuman,
		proto.Class_ClassMage:    proto.Race_RaceOrc,
		proto.Class_ClassPriest:  proto.Race_RaceGnome,
		proto.Class_ClassWarlock: proto.Race_RaceTroll,
		proto.Class_ClassDruid:   proto.Race_RaceHighOrderSkyborne,
	} {
		if !slices.Contains(ClassRaceCapabilities[class], race) {
			t.Errorf("%s should be playable as %s", class, race)
		}
	}
	if slices.Contains(ClassRaceCapabilities[proto.Class_ClassShaman], proto.Race_RaceHighOrderSkyborne) {
		t.Errorf("Only the Windshaper Skyborne can be shamans")
	}
	if slices.Contains(ClassRaceCapabilities[proto.Class_ClassMage], proto.Race_RaceWindshaperSkyborne) {
		t.Errorf("Only the High Order Skyborne can be mages")
	}
}

func TestBloodFuryIsPercentBased(t *testing.T) {
	sim, fw := setupRacialSim(proto.Race_RaceOrc)

	apBefore := fw.GetStat(stats.AttackPower)
	castRacial(t, sim, fw, 20572)

	if !WithinToleranceFloat64(apBefore*1.1, fw.GetStat(stats.AttackPower), 0.01) {
		t.Fatalf("Blood Fury should grant 10%% Attack Power: before %0.2f, after %0.2f", apBefore, fw.GetStat(stats.AttackPower))
	}
}

func TestBerserkingIsFlatTenPercent(t *testing.T) {
	sim, fw := setupRacialSim(proto.Race_RaceTroll)

	speedBefore := fw.PseudoStats.AttackSpeedMultiplier
	castRacial(t, sim, fw, 20554)

	if !WithinToleranceFloat64(speedBefore*1.1, fw.PseudoStats.AttackSpeedMultiplier, 0.0001) {
		t.Fatalf("Berserking should grant 10%% attack speed, got x%0.4f", fw.PseudoStats.AttackSpeedMultiplier/speedBefore)
	}
	if fw.CurrentRage() != 0 || fw.GetSpell(ActionID{SpellID: 20554}).Cost != nil {
		t.Fatalf("Berserking should have no cost")
	}
}

func TestElunesLight(t *testing.T) {
	sim, fw := setupRacialSim(proto.Race_RaceNightElf)

	critBefore := fw.GetStat(stats.PhysicalCritPercent)
	castRacial(t, sim, fw, 1259799)

	if !WithinToleranceFloat64(critBefore+10, fw.GetStat(stats.PhysicalCritPercent), 0.0001) {
		t.Fatalf("Elune's Light should grant 10%% crit: before %0.2f, after %0.2f", critBefore, fw.GetStat(stats.PhysicalCritPercent))
	}
}

func TestGnomeWarrior(t *testing.T) {
	sim, fw := setupRacialSim(proto.Race_RaceGnome)

	if fw.MaximumRage() != 105 {
		t.Fatalf("Expansive Mind should raise maximum Rage to 105, got %0.1f", fw.MaximumRage())
	}

	castRacial(t, sim, fw, 1259813)
	eureka := fw.GetAura("Eureka!")
	if !eureka.IsActive() || eureka.GetStacks() != 3 {
		t.Fatalf("Eureka! should start at 3 charges, got %d", eureka.GetStacks())
	}

	ability := fw.RegisterSpell(SpellConfig{
		ActionID:         ActionID{SpellID: 1},
		ClassSpellMask:   1,
		ProcMask:         ProcMaskMeleeMHSpecial,
		SpellSchool:      SpellSchoolPhysical,
		DamageMultiplier: 1,
	})
	itemSpell := fw.RegisterSpell(SpellConfig{
		ActionID:         ActionID{SpellID: 2},
		Flags:            SpellFlagAPL,
		ProcMask:         ProcMaskMeleeMHSpecial,
		SpellSchool:      SpellSchoolPhysical,
		DamageMultiplier: 1,
	})
	itemSpell.Cast(sim, fw.CurrentTarget)
	if itemSpell.DamageMultiplier != 1 || eureka.GetStacks() != 3 {
		t.Fatalf("Eureka! should ignore spells that are not class abilities")
	}
	if !WithinToleranceFloat64(1.1, ability.DamageMultiplier, 0.0001) {
		t.Fatalf("Eureka! should raise ability damage by 10%%, got x%0.2f", ability.DamageMultiplier)
	}

	for range 3 {
		ability.Cast(sim, fw.CurrentTarget)
	}
	if eureka.IsActive() {
		t.Fatalf("Eureka! should end after 3 abilities")
	}
	if !WithinToleranceFloat64(1, ability.DamageMultiplier, 0.0001) {
		t.Fatalf("Eureka! should drop its damage bonus when it ends, got x%0.2f", ability.DamageMultiplier)
	}
}

func TestTaurenEndurance(t *testing.T) {
	_, human := setupRacialSim(proto.Race_RaceHuman)
	_, tauren := setupRacialSim(proto.Race_RaceTauren)

	if !WithinToleranceFloat64(human.GetStat(stats.PhysicalHitPercent)+1, tauren.GetStat(stats.PhysicalHitPercent), 0.0001) {
		t.Fatalf("Endurance should grant 1%% hit")
	}
}

func TestTouchOfTheGraveUsesTheMeleeVariant(t *testing.T) {
	_, fw := setupRacialSim(proto.Race_RaceUndead)

	if aura := fw.GetAura("Touch of the Grave"); aura == nil || aura.ActionIDForProc.SpellID != 1260189 {
		t.Fatalf("An undead warrior should carry the 5%% Touch of the Grave")
	} else if aura.OnSpellHitDealt == nil {
		t.Fatalf("Touch of the Grave listens for no hit, so it never procs")
	}
	if fw.GetSpell(ActionID{SpellID: 1260198}) == nil {
		t.Fatalf("Touch of the Grave's drain is not registered")
	}
}

func TestSkyborneRacials(t *testing.T) {
	_, human := setupRacialSim(proto.Race_RaceHuman)

	for _, race := range []proto.Race{proto.Race_RaceHighOrderSkyborne, proto.Race_RaceWindshaperSkyborne} {
		_, fw := setupRacialSim(race)

		swingRatio := human.AutoAttacks.MainhandSwingSpeed().Seconds() / fw.AutoAttacks.MainhandSwingSpeed().Seconds()
		if !WithinToleranceFloat64(1.01, swingRatio, 0.001) {
			t.Errorf("%s: Wind Blessed should swing 1%% faster, got x%0.4f", race, swingRatio)
		}
		if at := fw.AttackTables[fw.CurrentTarget.UnitIndex]; !WithinToleranceFloat64(1.05, at.DamageDealtMultiplier, 0.0001) {
			t.Errorf("%s: Elemental Insight should grant 5%% damage against elementals, got x%0.4f", race, at.DamageDealtMultiplier)
		}

		hasReadLeyLine := fw.GetSpell(ActionID{SpellID: 1259705}) != nil
		hasSkysight := fw.GetSpell(ActionID{SpellID: 1259686}) != nil
		if hasReadLeyLine != (race == proto.Race_RaceHighOrderSkyborne) || hasSkysight != (race == proto.Race_RaceWindshaperSkyborne) {
			t.Errorf("%s: Read Ley Line is the High Order's and Skysight the Windshapers', got %t / %t", race, hasReadLeyLine, hasSkysight)
		}
	}
}

// Blood Fury's multipliers report the stats they add on gain and the stats they remove on expire.
func TestBloodFuryReportsItsTemporaryStats(t *testing.T) {
	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name:      "Orc",
				Class:     proto.Class_ClassWarrior,
				Race:      proto.Race_RaceOrc,
				Buffs:     &proto.IndividualBuffs{},
				Spec:      &proto.Player_DpsWarrior{},
				Equipment: &proto.EquipmentSpec{},
				Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 63, MobType: proto.MobType_MobTypeHumanoid}},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()

	character := sim.Raid.Parties[0].Players[0].GetCharacter()
	bloodFury := character.GetAura("Blood Fury")
	if bloodFury == nil {
		t.Fatal("no Blood Fury aura on an Orc")
	}

	var heard []stats.Stats
	character.OnTemporaryStatsChanges = append(character.OnTemporaryStatsChanges, func(_ *Simulation, aura *Aura, change stats.Stats) {
		if aura != bloodFury {
			t.Errorf("callback names aura %s, want Blood Fury", aura.Label)
		}
		heard = append(heard, change)
	})

	before := character.GetStats()
	bloodFury.Activate(sim)
	during := character.GetStats()
	bloodFury.Deactivate(sim)
	after := character.GetStats()

	if len(heard) != 2 {
		t.Fatalf("callback heard %d changes, want one on gain and one on expire", len(heard))
	}
	if heard[0] != during.Subtract(before) {
		t.Errorf("gain reported %s, want the stats the multipliers added: %s",
			heard[0].FlatString(), during.Subtract(before).FlatString())
	}
	if heard[1] != after.Subtract(during) {
		t.Errorf("expire reported %s, want the stats the multipliers removed: %s",
			heard[1].FlatString(), after.Subtract(during).FlatString())
	}
	if after != before {
		t.Errorf("stats after expire %s, want the ones before gain %s", after.FlatString(), before.FlatString())
	}

	attackPower := before[stats.AttackPower]
	if attackPower <= 0 {
		t.Fatalf("attack power is %v before Blood Fury, nothing for 10%% to show on", attackPower)
	}
	if got := heard[0][stats.AttackPower]; math.Abs(got-attackPower*0.1) > 1 {
		t.Errorf("gain reported %v attack power, want 10%% of %v", got, attackPower)
	}
	if got := heard[1][stats.AttackPower]; got != -heard[0][stats.AttackPower] {
		t.Errorf("expire reported %v attack power, want the reverse of the gain's %v", got, heard[0][stats.AttackPower])
	}
}

// A weapon specialization follows the weapons through an item swap in combat: a human's Sword
// Specialization and an orc's Axe Specialization move the crit of the sheet, of spells and of every
// auto when the main hand changes between a sword and an axe, and move it back on the swap back.
func TestWeaponSpecializationFollowsAnItemSwapInCombat(t *testing.T) {
	for _, tc := range []struct {
		race      proto.Race
		swordCrit float64
		axeCrit   float64
	}{
		{race: proto.Race_RaceHuman, swordCrit: 2},
		{race: proto.Race_RaceOrc, axeCrit: 1},
	} {
		t.Run(tc.race.String(), func(t *testing.T) {
			withSword := WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID)
			withAxe := WeaponTestSlots(WeaponTestAxeID, WeaponTestDaggerID, WeaponTestBowID)

			finalStats := func(equipped, swapped []*proto.ItemSpec) stats.Stats {
				result := ComputeStats(&proto.ComputeStatsRequest{Raid: WeaponTestRaid(tc.race, equipped, swapped), Encounter: WeaponTestEncounter()})
				return stats.FromUnitStatsProto(result.RaidStats.Parties[0].Players[0].FinalStats)
			}
			startSword, startAxe := finalStats(withSword, withAxe), finalStats(withAxe, withSword)
			for _, stat := range []stats.Stat{stats.PhysicalCritPercent, stats.SpellCritPercent} {
				if got := startSword[stat] - startAxe[stat]; !WithinToleranceFloat64(tc.swordCrit-tc.axeCrit, got, 1e-9) {
					t.Errorf("starting with the sword over the axe moves the sheet's %s by %v, want %v", stat.StatName(), got, tc.swordCrit-tc.axeCrit)
				}
			}

			sim, character := NewWeaponTestSim(t, tc.race, withSword, withAxe, nil)
			startCrit := stats.Stats{
				stats.PhysicalCritPercent: character.GetStat(stats.PhysicalCritPercent),
				stats.SpellCritPercent:    character.GetStat(stats.SpellCritPercent),
			}
			sim.PrePull()
			for sim.CurrentTime < 3*time.Second {
				sim.Step()
			}

			target := sim.Encounter.ActiveTargetUnits[0]
			attackTable := character.AttackTables[target.UnitIndex]
			spell := &Spell{Unit: &character.Unit, ProcMask: ProcMaskSpellDamage}
			// Lifts every chance clear of the level 63 target's crit suppression, which floors it at 0.
			character.AddStatsDynamic(sim, stats.Stats{stats.PhysicalCritPercent: 20, stats.SpellCritPercent: 20})

			critChances := func() map[string]float64 {
				chances := map[string]float64{
					"physical crit stat": character.GetStat(stats.PhysicalCritPercent) / 100,
					"spell crit stat":    character.GetStat(stats.SpellCritPercent) / 100,
					"spell":              spell.SpellCritChance(target),
				}
				for name, auto := range autos(character) {
					if auto.BonusCritPercent != 0 {
						t.Errorf("the %s carries %v%% crit of its own, want none", name, auto.BonusCritPercent)
					}
					chances[name] = auto.PhysicalCritChance(attackTable)
				}
				return chances
			}
			check := func(when string, weaponType proto.WeaponType, before, after map[string]float64, want float64) {
				t.Helper()
				if got := character.MainHand().WeaponType; got != weaponType {
					t.Fatalf("%s: main hand holds a %s, want a %s", when, got, weaponType)
				}
				for name := range before {
					if got := after[name] - before[name]; !WithinToleranceFloat64(want/100, got, 1e-9) {
						t.Errorf("%s at %v: the %s moved by %v, want %v", when, sim.CurrentTime, name, got, want/100)
					}
				}
			}

			sword := critChances()
			if sim.CurrentTime <= 0 {
				t.Fatalf("the swap is at %v, want it in combat", sim.CurrentTime)
			}
			character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
			axe := critChances()
			check("sword swapped for the axe", proto.WeaponType_WeaponTypeAxe, sword, axe, tc.axeCrit-tc.swordCrit)

			for sim.CurrentTime < 6*time.Second {
				sim.Step()
			}
			character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Main, false)
			check("axe swapped back for the sword", proto.WeaponType_WeaponTypeSword, axe, critChances(), tc.swordCrit-tc.axeCrit)

			// An iteration that ends on the axe starts the next one on the sword with the crit it began with.
			character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
			sim.Cleanup()
			sim.Reset()
			for _, stat := range []stats.Stat{stats.PhysicalCritPercent, stats.SpellCritPercent} {
				if got := character.GetStat(stat); !WithinToleranceFloat64(startCrit[stat], got, 1e-9) {
					t.Errorf("the next iteration starts at %v %s, want %v", got, stat.StatName(), startCrit[stat])
				}
			}
		})
	}
}
