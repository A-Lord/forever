package core

import (
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

// A wand-like autorepeat channel: one 100 damage hit per 2 s tick, for as long as the fight lasts.
func fakeShoot(char *Character, ticks *[]time.Duration) *Spell {
	spell := char.RegisterSpell(SpellConfig{
		ActionID:         ActionID{SpellID: 5019},
		Flags:            SpellFlagChanneled | SpellFlagAPL,
		SpellSchool:      SpellSchoolShadow,
		DefenseType:      DefenseTypeMagic,
		ProcMask:         ProcMaskRangedAuto,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		Dot: DotConfig{
			Aura:          Aura{Label: "Shoot"},
			NumberOfTicks: 1 << 20,
			TickLength:    2 * time.Second,
			Autorepeat:    true,
			OnTick: func(sim *Simulation, target *Unit, dot *Dot) {
				dot.Spell.CalcAndDealDamage(sim, target, 100, dot.Spell.OutcomeAlwaysHit)
				dot.Spell.Unit.Wand.RecordShot(sim)
				*ticks = append(*ticks, sim.CurrentTime)
			},
		},
		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			spell.Dot(target).Apply(sim)
		},
	})
	char.EnableWand(spell, 2*time.Second)
	return spell
}

// Ready once from readyAt.
type timedStubAction struct {
	defaultAPLActionImpl
	readyAt time.Duration
	ranAt   []time.Duration
}

func (action *timedStubAction) IsReady(sim *Simulation) bool {
	return sim.CurrentTime >= action.readyAt && len(action.ranAt) == 0
}
func (action *timedStubAction) Execute(sim *Simulation) {
	action.ranAt = append(action.ranAt, sim.CurrentTime)
}
func (action *timedStubAction) String() string { return "timed test action" }

type channelTest struct {
	sim        *Simulation
	fw         *FakeRageWarrior
	rot        *APLRotation
	spell      *Spell
	ticks      []time.Duration
	channelEnd []time.Duration
}

func newChannelTest(register func(ct *channelTest, char *Character) *Spell) *channelTest {
	ct := &channelTest{}
	ct.sim, ct.fw, ct.rot = setupFakeRageRotationWith(func(char *Character) {
		ct.spell = register(ct, char)
	})
	ct.spell.finalize()
	ct.fw.OnChannelEnd(func(sim *Simulation) {
		ct.channelEnd = append(ct.channelEnd, sim.CurrentTime)
	})
	return ct
}

func newShootTest() *channelTest {
	return newChannelTest(func(ct *channelTest, char *Character) *Spell {
		return fakeShoot(char, &ct.ticks)
	})
}

func (ct *channelTest) castAction() APLActionImpl {
	return ct.rot.newActionCastSpell(&proto.APLActionCastSpell{SpellId: ct.spell.ActionID.ToProto()})
}

func (ct *channelTest) channelAction() APLActionImpl {
	return ct.rot.newActionChannelSpell(&proto.APLActionChannelSpell{SpellId: ct.spell.ActionID.ToProto(), InterruptIf: constValue("true")})
}

func (ct *channelTest) recastChannelAction() APLActionImpl {
	return ct.rot.newActionChannelSpell(&proto.APLActionChannelSpell{SpellId: ct.spell.ActionID.ToProto(), InterruptIf: constValue("true"), AllowRecast: true})
}

// The pull's rotation action stays pending at 0; the melee swings are turned off.
func (ct *channelTest) start(actions ...APLActionImpl) {
	ct.rot.priorityList = nil
	for _, impl := range actions {
		action := &APLAction{impl: impl}
		action.Finalize(ct.rot)
		ct.rot.priorityList = append(ct.rot.priorityList, action)
	}
	ct.sim.PrePull()
	ct.sim.Step()
	ct.fw.AutoAttacks.CancelAutoSwing(ct.sim)
}

// Runs every pending action up to and including until, then moves the clock to until.
func (ct *channelTest) runUntil(until time.Duration) {
	sim := ct.sim
	for sim.pendingActions[len(sim.pendingActions)-1].NextActionAt <= until && !sim.Step() {
	}
	sim.advance(until)
}

func (ct *channelTest) damage() float64 {
	metrics := ct.spell.SpellMetrics[ct.fw.CurrentTarget.UnitIndex]
	return metrics.TotalDamage
}

func seconds(values ...float64) []time.Duration {
	return MapSlice(values, DurationFromSeconds)
}

func TestWandAutorepeatChannel(t *testing.T) {
	t.Run("a cast fires one shot", func(t *testing.T) {
		ct := newShootTest()
		ct.start()
		ct.runUntil(0)
		ct.castAction().Execute(ct.sim)

		ct.runUntil(1999 * time.Millisecond)
		if len(ct.ticks) != 0 || !ct.fw.Wand.IsWanding() {
			t.Fatalf("a cast Shoot should still wait for its shot at 1.999 s, shots at %v", ct.ticks)
		}
		ct.runUntil(2001 * time.Millisecond)
		if !slices.Equal(ct.ticks, seconds(2)) || ct.damage() != 100 {
			t.Fatalf("a cast Shoot should fire one 100 damage shot at 2 s, shots at %v for %.1f", ct.ticks, ct.damage())
		}
		if ct.fw.ChanneledDot != nil || ct.fw.Wand.IsWanding() {
			t.Fatalf("a cast Shoot should stop after its one shot")
		}
	})

	t.Run("a channel keeps shooting", func(t *testing.T) {
		ct := newShootTest()
		ct.start(ct.channelAction())

		ct.runUntil(3 * time.Second)
		wand := ct.fw.Wand
		if !wand.IsWanding() || wand.TimeToNext(ct.sim) != time.Second || wand.TimeSinceLast(ct.sim) != time.Second || wand.SwingSpeed() != 2*time.Second {
			t.Fatalf("at 3 s the wand should be shooting with 1 s to the next shot, 1 s since the last and a 2 s speed, got %t, %s, %s, %s",
				wand.IsWanding(), wand.TimeToNext(ct.sim), wand.TimeSinceLast(ct.sim), wand.SwingSpeed())
		}

		ct.runUntil(6001 * time.Millisecond)
		if !slices.Equal(ct.ticks, seconds(2, 4, 6)) || !wand.IsWanding() || len(ct.channelEnd) != 0 {
			t.Fatalf("a Shoot channel with nothing else ready should shoot at 2, 4 and 6 s and keep going, shots at %v, ended at %v", ct.ticks, ct.channelEnd)
		}

		ct.sim.Cleanup()
		ct.sim.Reset()
		ct.sim.advance(500 * time.Millisecond)
		if got := wand.TimeSinceLast(ct.sim); got != 500*time.Millisecond {
			t.Fatalf("after a reset the time since the last shot should count from the pull, got %s at 0.5 s", got)
		}
	})

	t.Run("an interrupt takes the shot's slot", func(t *testing.T) {
		ct := newShootTest()
		stub := &timedStubAction{readyAt: 3 * time.Second}
		ct.start(stub, ct.channelAction())

		ct.runUntil(5999 * time.Millisecond)
		if !slices.Equal(ct.ticks, seconds(2)) || ct.damage() != 100 {
			t.Fatalf("Shoot should fire at 2 s and give its 4 s slot away, shots at %v for %.1f", ct.ticks, ct.damage())
		}
		if !slices.Equal(ct.channelEnd, seconds(4)) || !slices.Equal(stub.ranAt, seconds(4)) {
			t.Fatalf("an action ready at 3 s should end Shoot and run at the 4 s boundary, channel ended at %v, action ran at %v", ct.channelEnd, stub.ranAt)
		}
	})

	t.Run("no interrupt when the GCD ends", func(t *testing.T) {
		ct := newShootTest()
		stub := &timedStubAction{readyAt: 500 * time.Millisecond}
		ct.start(stub, ct.channelAction())
		ct.runUntil(0)
		ct.fw.SetGCDTimer(ct.sim, 1500*time.Millisecond)

		ct.runUntil(3999 * time.Millisecond)
		if len(ct.ticks) != 0 {
			t.Fatalf("Shoot should give its 2 s slot to the waiting action, shots at %v", ct.ticks)
		}
		if !slices.Equal(ct.channelEnd, seconds(2)) || !slices.Equal(stub.ranAt, seconds(2)) {
			t.Fatalf("the action waiting on the GCD should end Shoot at the 2 s boundary, not at 1.5 s, channel ended at %v, action ran at %v", ct.channelEnd, stub.ranAt)
		}
	})

	t.Run("a normal channel ticks before its interrupt", func(t *testing.T) {
		ct := newChannelTest(func(ct *channelTest, char *Character) *Spell {
			config := fakeWarlockDot(689, "Drain Life", ClassFlags{Family: 5, Mask: [4]uint32{8}}, SpellFlagChanneled|SpellFlagAPL)
			config.Cast = CastConfig{DefaultCast: Cast{GCD: GCDDefault}}
			onTick := config.Dot.OnTick
			config.Dot.OnTick = func(sim *Simulation, target *Unit, dot *Dot) {
				onTick(sim, target, dot)
				ct.ticks = append(ct.ticks, sim.CurrentTime)
			}
			return char.RegisterSpell(config)
		})
		stub := &timedStubAction{readyAt: 500 * time.Millisecond}
		ct.start(stub, ct.channelAction())

		ct.runUntil(1001 * time.Millisecond)
		metrics := ct.spell.SpellMetrics[ct.fw.CurrentTarget.UnitIndex]
		if !slices.Equal(ct.ticks, seconds(1)) || metrics.TotalTickDamage == 0 {
			t.Fatalf("Drain Life should deal its 1 s tick before it ends, ticks at %v for %.1f", ct.ticks, metrics.TotalTickDamage)
		}
		if !slices.Equal(ct.channelEnd, seconds(1)) {
			t.Fatalf("Drain Life should end at its 1 s tick, ended at %v", ct.channelEnd)
		}
		ct.runUntil(1500 * time.Millisecond)
		if !slices.Equal(stub.ranAt, seconds(1.5)) {
			t.Fatalf("the action that ended Drain Life should run when the GCD ends at 1.5 s, ran at %v", stub.ranAt)
		}
	})

	t.Run("a channel allowed to recast keeps shooting", func(t *testing.T) {
		ct := newShootTest()
		ct.start(ct.recastChannelAction())

		ct.runUntil(6001 * time.Millisecond)
		if !slices.Equal(ct.ticks, seconds(2, 4, 6)) || len(ct.channelEnd) != 0 {
			t.Fatalf("a Shoot channel allowed to recast should shoot at 2, 4 and 6 s without being cut, shots at %v, ended at %v", ct.ticks, ct.channelEnd)
		}
	})

	t.Run("major cooldowns fire at a tick without ending the channel", func(t *testing.T) {
		var cooldown *Spell
		var cooldownAt []time.Duration
		ct := newChannelTest(func(ct *channelTest, char *Character) *Spell {
			cooldown = char.RegisterSpell(SpellConfig{
				ActionID: ActionID{SpellID: 910100},
				Cast:     CastConfig{CD: Cooldown{Timer: char.NewTimer(), Duration: time.Minute}},
				ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
					cooldownAt = append(cooldownAt, sim.CurrentTime)
				},
			})
			char.AddMajorCooldown(MajorCooldown{Spell: cooldown, Type: CooldownTypeDPS})
			return fakeShoot(char, &ct.ticks)
		})
		cooldown.finalize()
		ct.fw.majorCooldownManager.finalize()
		ct.fw.majorCooldownManager.reset(ct.sim)
		cooldown.CD.Set(3 * time.Second)
		ct.fw.UpdateMajorCooldowns()
		ct.start(ct.rot.newActionAutocastOtherCooldowns(&proto.APLActionAutocastOtherCooldowns{}), ct.channelAction())

		ct.runUntil(4001 * time.Millisecond)
		if !slices.Equal(cooldownAt, seconds(4)) {
			t.Fatalf("a cooldown ready at 3 s should fire at the 4 s tick, fired at %v", cooldownAt)
		}
		if !slices.Equal(ct.ticks, seconds(2, 4)) || len(ct.channelEnd) != 0 {
			t.Fatalf("Shoot should still fire at 4 s and keep going, shots at %v, ended at %v", ct.ticks, ct.channelEnd)
		}
	})

	t.Run("a unit without a wand reads zero", func(t *testing.T) {
		sim := &Simulation{CurrentTime: 3 * time.Second}
		var wand *Wand
		wand.RecordShot(sim)
		if wand.IsWanding() || wand.TimeToNext(sim) != 0 || wand.TimeSinceLast(sim) != 0 || wand.SwingSpeed() != 0 {
			t.Fatalf("a nil wand should read idle and zero")
		}
	})
}

func TestValueWand(t *testing.T) {
	type reads struct {
		swing, toNext, sinceLast time.Duration
		wanding                  bool
	}
	read := func(ct *channelTest) reads {
		rot, sim := ct.rot, ct.sim
		return reads{
			swing:     rot.newValueAutoSwingTime(&proto.APLValueAutoSwingTime{AutoType: proto.APLValueAutoSwingTime_Wand}, nil).GetDuration(sim),
			toNext:    rot.newValueAutoTimeToNext(&proto.APLValueAutoTimeToNext{AutoType: proto.APLValueAutoAttackType_WandAuto}, nil).GetDuration(sim),
			sinceLast: rot.newValueAutoTimeSinceLast(&proto.APLValueAutoTimeSinceLast{AutoType: proto.APLValueAutoAttackType_WandAuto}, nil).GetDuration(sim),
			wanding:   rot.newValueIsWanding(&proto.APLValueIsWanding{}, nil).GetBool(sim),
		}
	}

	t.Run("mid channel", func(t *testing.T) {
		ct := newShootTest()
		ct.start(ct.channelAction())
		ct.runUntil(3 * time.Second)
		if got, want := read(ct), (reads{2 * time.Second, time.Second, time.Second, true}); got != want {
			t.Fatalf("at 3 s into a Shoot channel the wand values read %+v, want %+v", got, want)
		}
	})

	t.Run("before any cast", func(t *testing.T) {
		ct := newShootTest()
		ct.start()
		if got := read(ct); got.toNext != 0 || got.wanding {
			t.Fatalf("before any cast the wand reads %+v, want no wait and not wanding", got)
		}
	})

	t.Run("no wand", func(t *testing.T) {
		ct := newShootTest()
		ct.fw.Wand = nil
		ct.start()
		ct.runUntil(3 * time.Second)
		if got := read(ct); got != (reads{}) {
			t.Fatalf("a unit without a wand reads %+v, want all zero", got)
		}
	})
}
