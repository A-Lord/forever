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

// A warrior ability as spelldata builds one: the game files' family 4 flags and no ClassSpellMask.
func fakeWarriorAbility(id int32, bit uint, damage float64) SpellConfig {
	return SpellConfig{
		ActionID:         ActionID{SpellID: id},
		ClassFlags:       ClassFlags{Family: 4, Mask: [4]uint32{1 << bit}},
		ProcMask:         ProcMaskMeleeMHSpecial,
		SpellSchool:      SpellSchoolPhysical,
		RageCost:         RageCostOptions{Cost: 30},
		DamageMultiplier: 1,
		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeAlwaysHit)
		},
	}
}

var fakeRendActionID = ActionID{SpellID: 772}

// A DoT's aura has to be registered before the environment is finalized.
func (fw *FakeRageWarrior) registerFakeRend() {
	config := fakeWarriorAbility(fakeRendActionID.SpellID, 5, 0)
	config.Dot = DotConfig{
		Aura:          Aura{Label: "Rend"},
		NumberOfTicks: 5,
		TickLength:    time.Second * 3,
		OnTick: func(sim *Simulation, target *Unit, dot *Dot) {
			dot.Spell.CalcAndDealPeriodicDamage(sim, target, 50, dot.OutcomeTick)
		},
	}
	config.ApplyEffects = func(sim *Simulation, target *Unit, spell *Spell) {
		if spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit).Landed() {
			spell.Dot(target).Apply(sim)
		}
	}
	fw.RegisterSpell(config)
}

func TestGnomeWarrior(t *testing.T) {
	sim, fw := setupRacialSim(proto.Race_RaceGnome)
	fw.AutoAttacks.CancelAutoSwing(sim)

	if fw.MaximumRage() != 105 {
		t.Fatalf("Expansive Mind should raise maximum Rage to 105, got %0.1f", fw.MaximumRage())
	}

	castRacial(t, sim, fw, 1259813)
	eureka := fw.GetAura("Eureka!")
	if !eureka.IsActive() || eureka.GetStacks() != 3 {
		t.Fatalf("Eureka! should start at 3 charges, got %d", eureka.GetStacks())
	}

	mortalStrike := fw.RegisterSpell(fakeWarriorAbility(12294, 25, 100))
	hamstring := fw.RegisterSpell(fakeWarriorAbility(1715, 1, 0))
	sunderArmor := fw.RegisterSpell(fakeWarriorAbility(7386, 14, 0))
	rend := fw.GetSpell(fakeRendActionID)
	itemSpell := fw.RegisterSpell(SpellConfig{
		ActionID:         ActionID{SpellID: 2},
		Flags:            SpellFlagAPL,
		ProcMask:         ProcMaskMeleeMHSpecial,
		SpellSchool:      SpellSchoolPhysical,
		DamageMultiplier: 1,
		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeAlwaysHit)
		},
	})

	if !WithinToleranceFloat64(1.1, mortalStrike.DamageMultiplier, 0.0001) {
		t.Fatalf("Eureka! should raise Mortal Strike's damage by 10%%, got x%0.2f", mortalStrike.DamageMultiplier)
	}
	if !WithinToleranceFloat64(18, mortalStrike.Cost.GetCurrentCost(), 0.0001) {
		t.Fatalf("Eureka! should cut Mortal Strike's 30 Rage by 40%%, got %0.1f", mortalStrike.Cost.GetCurrentCost())
	}
	if sunderArmor.DamageMultiplier != 1 || sunderArmor.Cost.GetCurrentCost() != 30 {
		t.Fatalf("Eureka! should leave Sunder Armor alone, which its class mask does not name")
	}

	// The spend lands one spell batch after the hit.
	stacksAfter := func(spell *Spell) int32 {
		spell.SkipCastAndApplyEffects(sim, fw.CurrentTarget)
		for end := sim.CurrentTime + SpellBatchWindow; sim.CurrentTime <= end && !sim.Step(); {
		}
		return eureka.GetStacks()
	}

	if stacksAfter(itemSpell) != 3 || itemSpell.DamageMultiplier != 1 {
		t.Fatalf("Eureka! should ignore spells that are not class abilities")
	}
	if stacksAfter(sunderArmor) != 3 {
		t.Fatalf("Sunder Armor should not spend a charge")
	}
	if stacksAfter(hamstring) != 3 {
		t.Fatalf("an ability the mask names should not spend a charge when it deals no damage")
	}
	if stacksAfter(rend) != 2 {
		t.Fatalf("Rend's landed application deals no damage but should spend a charge for its DoT")
	}
	stacksAfter(mortalStrike)
	stacksAfter(mortalStrike)
	if eureka.IsActive() {
		t.Fatalf("Eureka! should end after 3 abilities, %d left at %v", eureka.GetStacks(), sim.CurrentTime)
	}
	if !WithinToleranceFloat64(1, mortalStrike.DamageMultiplier, 0.0001) {
		t.Fatalf("Eureka! should drop its damage bonus when it ends, got x%0.2f", mortalStrike.DamageMultiplier)
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

func init() {
	RegisterAgentFactory(
		proto.Player_Warlock{},
		proto.Spec_SpecWarlock,
		newFakeDotWarlock,
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_Warlock)
			if !ok {
				panic("Invalid spec value for Warlock!")
			}
			player.Spec = playerSpec
		},
	)
}

type fakeDotWarlock struct {
	Character

	Corruption    *Spell
	Curse         *Spell
	damagingTicks int
}

func newFakeDotWarlock(char *Character, _ *proto.Player, _ *proto.Raid) Agent {
	return &fakeDotWarlock{Character: *char}
}

func (fw *fakeDotWarlock) GetCharacter() *Character { return &fw.Character }

func (fw *fakeDotWarlock) Initialize() {
	fw.Corruption = fw.RegisterSpell(SpellConfig{
		ActionID:         ActionID{SpellID: 172},
		SpellSchool:      SpellSchoolShadow,
		DefenseType:      DefenseTypeMagic,
		ProcMask:         ProcMaskSpellDamage,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: DotConfig{
			Aura:          Aura{Label: "Corruption"},
			NumberOfTicks: 6,
			TickLength:    time.Second * 3,
			OnTick: func(sim *Simulation, target *Unit, dot *Dot) {
				if dot.Spell.CalcAndDealPeriodicDamage(sim, target, 100, dot.OutcomeTick).Damage > 0 {
					fw.damagingTicks++
				}
			},
		},

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
			}
			spell.DealOutcome(sim, result)
		},
	})

	fw.Curse = fw.RegisterSpell(SpellConfig{
		ActionID:         ActionID{SpellID: 702},
		SpellSchool:      SpellSchoolShadow,
		DefenseType:      DefenseTypeMagic,
		ProcMask:         ProcMaskSpellDamage,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
		},
	})
}

func (fw *fakeDotWarlock) ApplyTalents()                  {}
func (fw *fakeDotWarlock) Reset(_ *Simulation)            {}
func (fw *fakeDotWarlock) OnGCDReady(_ *Simulation)       {}
func (fw *fakeDotWarlock) OnEncounterStart(_ *Simulation) {}

func setupTouchOfTheGraveWarlock(t *testing.T) (*Simulation, *fakeDotWarlock, *Aura) {
	t.Helper()
	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 100},
		Raid: SinglePlayerRaidProto(&proto.Player{
			Name:      "Warlock",
			Class:     proto.Class_ClassWarlock,
			Race:      proto.Race_RaceUndead,
			Buffs:     &proto.IndividualBuffs{},
			Spec:      &proto.Player_Warlock{},
			Equipment: &proto.EquipmentSpec{},
			Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
		}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 63, MobType: proto.MobType_MobTypeHumanoid}},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()

	fw := sim.Raid.Parties[0].Players[0].(*fakeDotWarlock)
	aura := fw.GetAura("Touch of the Grave")
	if aura == nil || !aura.IsActive() || aura.Icd == nil {
		t.Fatalf("An undead warlock should carry an active Touch of the Grave with its cooldown")
	}
	return sim, fw, aura
}

// Whether Touch of the Grave procs within the given number of hits. Time never advances, so the
// first proc leaves its cooldown running for good.
func touchOfTheGraveProcsWithin(sim *Simulation, aura *Aura, tries int, hit func()) bool {
	for range tries {
		hit()
		if !aura.Icd.IsReady(sim) {
			return true
		}
	}
	return false
}

func TestTouchOfTheGraveProcsOnTheLandedCastOfADot(t *testing.T) {
	sim, fw, aura := setupTouchOfTheGraveWarlock(t)

	if !touchOfTheGraveProcsWithin(sim, aura, 200, func() { fw.Corruption.SkipCastAndApplyEffects(sim, fw.CurrentTarget) }) {
		t.Fatalf("A DoT's landed cast deals no damage but should still trigger Touch of the Grave")
	}
}

func TestTouchOfTheGraveIgnoresSpellsWithoutDamage(t *testing.T) {
	sim, fw, aura := setupTouchOfTheGraveWarlock(t)

	if touchOfTheGraveProcsWithin(sim, aura, 200, func() { fw.Curse.SkipCastAndApplyEffects(sim, fw.CurrentTarget) }) {
		t.Fatalf("A debuff with no damage should not trigger Touch of the Grave")
	}
}

func TestTouchOfTheGraveIgnoresPeriodicTicks(t *testing.T) {
	sim, fw, aura := setupTouchOfTheGraveWarlock(t)

	dot := fw.Corruption.Dot(fw.CurrentTarget)
	dot.Apply(sim)
	if touchOfTheGraveProcsWithin(sim, aura, 200, func() { dot.TickOnce(sim) }) {
		t.Fatalf("A DoT's ticks should not trigger Touch of the Grave")
	}
	if fw.damagingTicks == 0 {
		t.Fatalf("The DoT's ticks dealt no damage, so the test proves nothing")
	}
}
