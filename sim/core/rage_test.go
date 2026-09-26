package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

func init() {
	RegisterAgentFactory(
		proto.Player_DpsWarrior{},
		proto.Spec_SpecDpsWarrior,
		NewFakeRageWarrior,
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_DpsWarrior)
			if !ok {
				panic("Invalid spec value for Dps Warrior!")
			}
			player.Spec = playerSpec
		},
	)
}

const (
	fakeMHSwingSpeed = 2.6
	fakeOHSwingSpeed = 1.8
)

type FakeRageWarrior struct {
	Character
}

func (fw *FakeRageWarrior) GetCharacter() *Character { return &fw.Character }

func (fw *FakeRageWarrior) Initialize() {
	fw.registerFakeThreatSpell()
	fw.registerFakeGCDSpell()
	fw.registerFakeCategorySpells()
	fw.registerFakeRend()
}
func (fw *FakeRageWarrior) ApplyTalents()                  {}
func (fw *FakeRageWarrior) Reset(_ *Simulation)            {}
func (fw *FakeRageWarrior) OnGCDReady(_ *Simulation)       {}
func (fw *FakeRageWarrior) OnEncounterStart(_ *Simulation) {}

func NewFakeRageWarrior(char *Character, _ *proto.Player, _ *proto.Raid) Agent {
	fw := &FakeRageWarrior{
		Character: *char,
	}

	fw.EnableRageBar(RageBarOptions{
		MaxRage:            100,
		BaseRageMultiplier: 1,
		StartingRage:       0,
	})

	fw.EnableAutoAttacks(fw, AutoAttackOptions{
		MainHand:       Weapon{SwingSpeed: fakeMHSwingSpeed},
		OffHand:        Weapon{SwingSpeed: fakeOHSwingSpeed},
		AutoSwingMelee: true,
	})

	return fw
}

func SetupFakeRageSim() *Simulation {
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
							Buffs:     &proto.IndividualBuffs{},
							Spec:      &proto.Player_DpsWarrior{},
							Equipment: &proto.EquipmentSpec{},
							// AddRage() pokes the rotation, so it needs to be non-nil.
							Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
						},
					},
					Buffs: &proto.PartyBuffs{},
				},
			},
		},
		Encounter: &proto.Encounter{
			Targets: []*proto.Target{
				{Name: "target", Level: 63, MobType: proto.MobType_MobTypeHumanoid},
			},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()

	return sim
}

// Feeds a hand-crafted auto attack result to the Rage bar and returns the Rage gained from it.
// PostArmorAndResistanceMultiplier is always populated, because the sim fills it in before the
// outcome is applied, even for swings which end up doing no damage.
func rageFromAutoAttack(sim *Simulation, fw *FakeRageWarrior, spell *Spell, outcome HitOutcome, preOutcomeDamage float64) float64 {
	result := &SpellResult{
		Target:                           sim.Encounter.ActiveTargetUnits[0],
		Outcome:                          outcome,
		PostArmorAndResistanceMultiplier: preOutcomeDamage,
	}
	if outcome.Matches(OutcomeLanded) {
		result.Damage = preOutcomeDamage
	}

	rageBefore := fw.CurrentRage()
	fw.Unit.OnSpellHitDealt(sim, spell, result)

	return fw.CurrentRage() - rageBefore
}

func TestAutoAttackRageGeneration(t *testing.T) {
	// A one-hand MH swing at 2.6 speed: 2.6 * 3.46 = 8.996, whatever it dealt.
	const swingDamage = 500.0

	tests := []struct {
		name     string
		outcome  HitOutcome
		wantRage float64
	}{
		{
			name:     "hit",
			outcome:  OutcomeHit,
			wantRage: 8.996,
		},
		{
			name:     "crit",
			outcome:  OutcomeCrit,
			wantRage: 8.996,
		},
		{
			name:     "glance",
			outcome:  OutcomeGlance,
			wantRage: 8.996,
		},
		{
			name:     "dodge",
			outcome:  OutcomeDodge,
			wantRage: 0,
		},
		{
			name:     "parry",
			outcome:  OutcomeParry,
			wantRage: 0,
		},
		{
			name:     "miss",
			outcome:  OutcomeMiss,
			wantRage: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sim := SetupFakeRageSim()
			fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)

			gained := rageFromAutoAttack(sim, fw, fw.AutoAttacks.MHAuto(), test.outcome, swingDamage)

			if !WithinToleranceFloat64(test.wantRage, gained, 0.01) {
				t.Fatalf("Incorrect Rage generated on %s: Expected: %0.3f, Actual: %0.3f", test.name, test.wantRage, gained)
			}
		})
	}
}

func TestTwoHandAutoAttackRageGeneration(t *testing.T) {
	// A two-hand MH swing at 2.6 speed: 2.6 * 4.5 = 11.7.
	const swingDamage = 500.0

	sim := SetupFakeRageSim()
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	fw.AutoAttacks.MH().NormalizedSwingSpeed = TwoHandNormalizedSwingSpeed

	hitRage := rageFromAutoAttack(sim, fw, fw.AutoAttacks.MHAuto(), OutcomeHit, swingDamage)
	if !WithinToleranceFloat64(11.7, hitRage, 0.001) {
		t.Fatalf("Incorrect Rage generated on 2H hit: Expected: %0.3f, Actual: %0.3f", 11.695, hitRage)
	}
}

func TestOffHandAutoAttackRageGeneration(t *testing.T) {
	// An OH swing at 1.8 speed: 1.8 * 3.46 * 0.5 = 3.114.
	const swingDamage = 500.0

	sim := SetupFakeRageSim()
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	ohAuto := fw.AutoAttacks.OHAuto()

	hitRage := rageFromAutoAttack(sim, fw, ohAuto, OutcomeHit, swingDamage)
	if !WithinToleranceFloat64(3.114, hitRage, 0.01) {
		t.Fatalf("Incorrect Rage generated on OH hit: Expected: %0.3f, Actual: %0.3f", 3.114, hitRage)
	}

	dodgeRage := rageFromAutoAttack(sim, fw, ohAuto, OutcomeDodge, swingDamage)
	if dodgeRage != 0 {
		t.Fatalf("Dodged OH swing generated %0.3f Rage, expected none", dodgeRage)
	}
}

func TestOffHandRageMultiplier(t *testing.T) {
	const swingDamage = 500.0

	sim := SetupFakeRageSim()
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)

	unmodified := rageFromAutoAttack(sim, fw, fw.AutoAttacks.OHAuto(), OutcomeHit, swingDamage)
	if !WithinToleranceFloat64(3.114, unmodified, 0.01) {
		t.Fatalf("Incorrect Rage generated on OH hit: Expected: %0.3f, Actual: %0.3f", 3.114, unmodified)
	}

	fw.SetOffHandRageMultiplier(1.5)

	modified := rageFromAutoAttack(sim, fw, fw.AutoAttacks.OHAuto(), OutcomeHit, swingDamage)
	if !WithinToleranceFloat64(4.671, modified, 0.01) {
		t.Fatalf("Incorrect Rage generated on multiplied OH hit: Expected: %0.3f, Actual: %0.3f", 4.671, modified)
	}

	mhRage := rageFromAutoAttack(sim, fw, fw.AutoAttacks.MHAuto(), OutcomeHit, swingDamage)
	if !WithinToleranceFloat64(8.996, mhRage, 0.01) {
		t.Fatalf("Incorrect Rage generated on MH hit: Expected: %0.3f, Actual: %0.3f", 8.996, mhRage)
	}
}

func newPowerBarCharacter() *Character {
	character := NewCharacter(&Party{}, 0, &proto.Player{
		Name:      "Power Bar Tester",
		Class:     proto.Class_ClassWarrior,
		Race:      proto.Race_RaceHuman,
		Spec:      &proto.Player_ProtectionWarrior{ProtectionWarrior: &proto.ProtectionWarrior{}},
		Equipment: &proto.EquipmentSpec{},
	})
	return &character
}

func TestSetMaximumRage(t *testing.T) {
	sim := SetupFakeRageSim()
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	metrics := fw.NewRageMetrics(ActionID{SpellID: 1310236})

	fw.SetMaximumRage(sim, 130, metrics)
	if fw.MaximumRage() != 130 || fw.CurrentRage() != 0 {
		t.Fatalf("raised to 130: max %v with %v rage, want 130 with none granted", fw.MaximumRage(), fw.CurrentRage())
	}

	fw.AddRage(sim, 120, metrics)
	if fw.CurrentRage() != 120 {
		t.Fatalf("rage under the raised maximum: %v, want 120", fw.CurrentRage())
	}

	fw.SetMaximumRage(sim, 100, metrics)
	if fw.MaximumRage() != 100 || fw.CurrentRage() != 100 {
		t.Fatalf("lowered to 100: max %v with %v rage, want 100 with 100", fw.MaximumRage(), fw.CurrentRage())
	}
	if metrics.ActualGain != 100 || metrics.Events != 2 {
		t.Errorf("metrics after gaining 120 and losing 20: actual gain %v over %d events, want 100 over 2",
			metrics.ActualGain, metrics.Events)
	}
}

func TestSetMaximumRageClampsTheStartingRage(t *testing.T) {
	character := newPowerBarCharacter()
	character.EnableRageBar(RageBarOptions{MaxRage: 100, StartingRage: 120, BaseRageMultiplier: 1})

	steps := []struct {
		maxRage float64
		want    float64
	}{
		{100, 100},
		{130, 120},
		{110, 110},
	}
	for _, step := range steps {
		character.SetMaximumRage(nil, step.maxRage, nil)
		character.rageBar.reset(nil)
		if got := character.CurrentRage(); got != step.want {
			t.Errorf("starting rage 120 under a maximum of %v: %v, want %v", step.maxRage, got, step.want)
		}
	}
}

func TestSetMaximumEnergy(t *testing.T) {
	character := newPowerBarCharacter()
	character.EnableEnergyBar(EnergyBarOptions{MaxEnergy: 100, MaxComboPoints: 5, HasNoRegen: true})
	metrics := character.NewEnergyMetrics(ActionID{SpellID: 14983})
	sim := &Simulation{}

	character.energyBar.currentEnergy = 100
	character.SetMaximumEnergy(sim, 110, metrics)
	if character.MaximumEnergy() != 110 || character.CurrentEnergy() != 100 {
		t.Fatalf("raised to 110: max %v with %v energy, want 110 with 100", character.MaximumEnergy(), character.CurrentEnergy())
	}

	character.energyBar.currentEnergy = 110
	character.SetMaximumEnergy(sim, 100, metrics)
	if character.MaximumEnergy() != 100 || character.CurrentEnergy() != 100 {
		t.Fatalf("lowered to 100: max %v with %v energy, want 100 with 100", character.MaximumEnergy(), character.CurrentEnergy())
	}
	if metrics.ActualGain != -10 || metrics.Events != 1 {
		t.Errorf("metrics after losing 10: actual gain %v over %d events, want -10 over 1", metrics.ActualGain, metrics.Events)
	}
}
