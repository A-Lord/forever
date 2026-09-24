package core

import (
	"slices"
	"testing"

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
