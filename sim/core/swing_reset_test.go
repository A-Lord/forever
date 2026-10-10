package core

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

func init() {
	RegisterAgentFactory(
		proto.Player_RetributionPaladin{},
		proto.Spec_SpecRetributionPaladin,
		NewFakeSwingCaster,
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_RetributionPaladin)
			if !ok {
				panic("Invalid spec value for Retribution Paladin!")
			}
			player.Spec = playerSpec
		},
	)
}

const swingResetCastTime = time.Second * 2

type FakeSwingCaster struct {
	Character

	Resetting *Spell // A hardcast that restarts the melee swing.
	Keeping   *Spell // The same hardcast, flagged to leave the swing timer running.
}

func (fc *FakeSwingCaster) GetCharacter() *Character { return &fc.Character }

func (fc *FakeSwingCaster) ApplyTalents()                  {}
func (fc *FakeSwingCaster) Reset(_ *Simulation)            {}
func (fc *FakeSwingCaster) OnGCDReady(_ *Simulation)       {}
func (fc *FakeSwingCaster) OnEncounterStart(_ *Simulation) {}

func (fc *FakeSwingCaster) Initialize() {
	register := func(id int32, flags SpellFlag) *Spell {
		return fc.RegisterSpell(SpellConfig{
			ActionID:     ActionID{SpellID: id},
			SpellSchool:  SpellSchoolHoly,
			ProcMask:     ProcMaskEmpty,
			Flags:        flags | SpellFlagNoMetrics,
			Cast:         CastConfig{DefaultCast: Cast{CastTime: swingResetCastTime}},
			ApplyEffects: func(_ *Simulation, _ *Unit, _ *Spell) {},
		})
	}
	fc.Resetting = register(1, 0)
	fc.Keeping = register(2, SpellFlagKeepsSwingTimer)
}

func NewFakeSwingCaster(char *Character, _ *proto.Player, _ *proto.Raid) Agent {
	fc := &FakeSwingCaster{
		Character: *char,
	}

	fc.EnableAutoAttacks(fc, AutoAttackOptions{
		MainHand:       Weapon{SwingSpeed: fakeMHSwingSpeed},
		OffHand:        Weapon{SwingSpeed: fakeOHSwingSpeed},
		AutoSwingMelee: true,
	})

	return fc
}

// A dual-wielding caster one second into the fight: the MH swung at 0 and is next due at 2.6s, the
// OH swung at 0.9s and is next due at 2.7s.
func setupSwingResetSim(t *testing.T) (*Simulation, *FakeSwingCaster) {
	t.Helper()
	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{
			RandomSeed: 100,
		},
		Raid: &proto.Raid{
			Parties: []*proto.Party{
				{
					Players: []*proto.Player{
						{
							Name:      "Paladin",
							Class:     proto.Class_ClassPaladin,
							Buffs:     &proto.IndividualBuffs{},
							Spec:      &proto.Player_RetributionPaladin{},
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
				{Name: "target", Level: 63, MobType: proto.MobType_MobTypeHumanoid},
			},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()

	fc := sim.Raid.Parties[0].Players[0].(*FakeSwingCaster)
	stepSwingSim(t, sim, time.Second)
	return sim, fc
}

func stepSwingSim(t *testing.T, sim *Simulation, until time.Duration) {
	t.Helper()
	sim.AddPendingAction(NewDelayedAction(DelayedActionOptions{DoAt: until, OnAction: func(*Simulation) {}}))
	for steps := 0; sim.CurrentTime < until; steps++ {
		if steps > 10000 {
			t.Fatalf("the sim did not reach %v", until)
		}
		sim.Step()
	}
}

// Completing a cast starts a full swing on each hand, at the attack speed the cast completes with.
// No swing lands while the cast is underway.
func TestCompletedCastResetsSwingTimer(t *testing.T) {
	t.Run("the next swings land one full swing after the cast", func(t *testing.T) {
		sim, fc := setupSwingResetSim(t)
		aa := &fc.AutoAttacks
		mhPrevious, ohPrevious := aa.MainhandPreviousSwing(), aa.OffhandPreviousSwing()

		fc.Resetting.Cast(sim, fc.CurrentTarget)
		castEnd := sim.CurrentTime + swingResetCastTime
		stepSwingSim(t, sim, castEnd)

		if aa.MainhandPreviousSwing() != mhPrevious || aa.OffhandPreviousSwing() != ohPrevious {
			t.Fatalf("expected no swing during the cast, got MH at %v and OH at %v", aa.MainhandPreviousSwing(), aa.OffhandPreviousSwing())
		}
		if expected := castEnd + aa.MainhandSwingSpeed(); aa.MainhandSwingAt() != expected {
			t.Fatalf("MH swing: expected %v, got %v", expected, aa.MainhandSwingAt())
		}
		if expected := castEnd + aa.OffhandSwingSpeed(); aa.OffhandSwingAt() != expected {
			t.Fatalf("OH swing: expected %v, got %v", expected, aa.OffhandSwingAt())
		}
	})

	t.Run("haste gained during the cast sets the swing that follows it", func(t *testing.T) {
		sim, fc := setupSwingResetSim(t)
		aa := &fc.AutoAttacks

		fc.Resetting.Cast(sim, fc.CurrentTarget)
		castEnd := sim.CurrentTime + swingResetCastTime
		stepSwingSim(t, sim, sim.CurrentTime+swingResetCastTime/2)
		fc.MultiplyMeleeSpeed(sim, 1.3)
		stepSwingSim(t, sim, castEnd)

		if expected := castEnd + DurationFromSeconds(fakeMHSwingSpeed/1.3); aa.MainhandSwingAt() != expected {
			t.Fatalf("MH swing: expected one hasted swing after the cast (%v), got %v", expected, aa.MainhandSwingAt())
		}
	})

	t.Run("a pushed back cast holds the swing to its new end", func(t *testing.T) {
		sim, fc := setupSwingResetSim(t)
		aa := &fc.AutoAttacks
		mhPrevious := aa.MainhandPreviousSwing()

		fc.Resetting.Cast(sim, fc.CurrentTarget)
		stepSwingSim(t, sim, sim.CurrentTime+swingResetCastTime/2)
		pushback := fc.Hardcast.pushBack(sim.CurrentTime)
		fc.newHardcastAction(sim)
		castEnd := time.Second + swingResetCastTime + pushback
		stepSwingSim(t, sim, castEnd)

		if aa.MainhandPreviousSwing() != mhPrevious {
			t.Fatalf("expected no swing during the pushed back cast, got one at %v", aa.MainhandPreviousSwing())
		}
		if expected := castEnd + aa.MainhandSwingSpeed(); aa.MainhandSwingAt() != expected {
			t.Fatalf("MH swing: expected %v, got %v", expected, aa.MainhandSwingAt())
		}
	})

	t.Run("a cast that keeps the swing timer leaves the swings alone", func(t *testing.T) {
		sim, fc := setupSwingResetSim(t)
		aa := &fc.AutoAttacks

		fc.Keeping.Cast(sim, fc.CurrentTarget)
		stepSwingSim(t, sim, sim.CurrentTime+swingResetCastTime)

		mhSwingSpeed := DurationFromSeconds(fakeMHSwingSpeed)
		if aa.MainhandPreviousSwing() != mhSwingSpeed {
			t.Fatalf("MH swing: expected it to land during the cast at %v, got %v", mhSwingSpeed, aa.MainhandPreviousSwing())
		}
		if aa.MainhandSwingAt() != 2*mhSwingSpeed {
			t.Fatalf("MH swing: expected the next one at %v, got %v", 2*mhSwingSpeed, aa.MainhandSwingAt())
		}
	})
}
