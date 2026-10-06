package core

import (
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

// A wand-like autorepeat channel: one 100 damage hit per 2 s tick, for as long as the fight lasts. Like
// RegisterShoot's Shoot it is hasted by cast speed and refused while moving.
func fakeShoot(char *Character, ticks *[]time.Duration) *Spell {
	spell := char.RegisterSpell(SpellConfig{
		ActionID:         ActionID{SpellID: 5019},
		Flags:            SpellFlagChanneled | SpellFlagAPL,
		SpellSchool:      SpellSchoolShadow,
		DefenseType:      DefenseTypeMagic,
		ProcMask:         ProcMaskRangedAuto,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ExtraCastCondition: func(_ *Simulation, _ *Unit) bool {
			return !char.Moving
		},
		Dot: DotConfig{
			Aura:                Aura{Label: "Shoot"},
			NumberOfTicks:       1 << 20,
			TickLength:          2 * time.Second,
			AffectedByCastSpeed: true,
			Autorepeat:          true,
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

// Shoot and one more spell, which gets its own cooldown timer when the config names a cooldown.
func newSpellShootTest(config SpellConfig) (*channelTest, *Spell) {
	var spell *Spell
	ct := newChannelTest(func(ct *channelTest, char *Character) *Spell {
		if config.Cast.CD.Duration > 0 {
			config.Cast.CD.Timer = char.NewTimer()
		}
		spell = char.RegisterSpell(config)
		return fakeShoot(char, &ct.ticks)
	})
	spell.finalize()
	return ct, spell
}

func (ct *channelTest) wanding() bool {
	return ct.spell.autorepeatChannel() != nil
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

func (ct *channelTest) autocastAction() APLActionImpl {
	return ct.rot.newActionAutocastOtherCooldowns(&proto.APLActionAutocastOtherCooldowns{})
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

// A spell that only records when it is cast.
func recordedSpellConfig(id int32, castAt *[]time.Duration) SpellConfig {
	return SpellConfig{
		ActionID: ActionID{SpellID: id},
		Flags:    SpellFlagAPL,
		ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
			*castAt = append(*castAt, sim.CurrentTime)
		},
	}
}

// A one-minute major cooldown, on the GCD when gcd is set, ready from 3 s.
type testCooldown struct {
	gcd      time.Duration
	priority int32
	spell    *Spell
	castAt   []time.Duration
}

func newCooldownShootTest(cooldowns ...*testCooldown) *channelTest {
	ct := newChannelTest(func(ct *channelTest, char *Character) *Spell {
		for i, cooldown := range cooldowns {
			config := recordedSpellConfig(910100+int32(i), &cooldown.castAt)
			config.Cast = CastConfig{DefaultCast: Cast{GCD: cooldown.gcd}, CD: Cooldown{Timer: char.NewTimer(), Duration: time.Minute}}
			cooldown.spell = char.RegisterSpell(config)
			char.AddMajorCooldown(MajorCooldown{Spell: cooldown.spell, Type: CooldownTypeDPS, Priority: cooldown.priority})
		}
		return fakeShoot(char, &ct.ticks)
	})
	ct.fw.majorCooldownManager.finalize()
	ct.fw.majorCooldownManager.reset(ct.sim)
	for _, cooldown := range cooldowns {
		cooldown.spell.finalize()
		cooldown.spell.CD.Set(3 * time.Second)
	}
	ct.fw.UpdateMajorCooldowns()
	return ct
}

// A Drain Life-like channel of five 1 s ticks, with its ticks recorded.
func drainConfig(drainTicks *[]time.Duration, cast CastConfig) SpellConfig {
	config := fakeWarlockDot(689, "Drain Life", ClassFlags{Family: 5, Mask: [4]uint32{8}}, SpellFlagChanneled|SpellFlagAPL)
	config.Cast = cast
	onTick := config.Dot.OnTick
	config.Dot.OnTick = func(sim *Simulation, target *Unit, dot *Dot) {
		onTick(sim, target, dot)
		*drainTicks = append(*drainTicks, sim.CurrentTime)
	}
	return config
}

// An instant on the GCD that records its casts.
func nukeConfig(castAt *[]time.Duration) SpellConfig {
	config := recordedSpellConfig(910110, castAt)
	config.Cast = CastConfig{DefaultCast: Cast{GCD: GCDDefault}}
	return config
}

func TestWandAutorepeatChannel(t *testing.T) {
	t.Run("a cast fires one shot", func(t *testing.T) {
		ct := newShootTest()
		ct.start(&timedStubAction{readyAt: time.Minute})
		ct.runUntil(0)
		ct.castAction().Execute(ct.sim)

		ct.runUntil(499 * time.Millisecond)
		if len(ct.ticks) != 0 || !ct.wanding() {
			t.Fatalf("a cast Shoot should still wind up at 0.499 s, shots at %v", ct.ticks)
		}
		ct.runUntil(2001 * time.Millisecond)
		if !slices.Equal(ct.ticks, seconds(0.5)) || ct.damage() != 100 {
			t.Fatalf("a cast Shoot should fire one 100 damage shot at 0.5 s, shots at %v for %.1f", ct.ticks, ct.damage())
		}
		if ct.fw.ChanneledDot != nil {
			t.Fatalf("a cast Shoot should stop after its one shot")
		}
	})

	t.Run("a channel keeps shooting", func(t *testing.T) {
		ct := newShootTest()
		ct.start(ct.channelAction())

		ct.runUntil(3500 * time.Millisecond)
		wand := ct.fw.Wand
		if !ct.wanding() || wand.TimeToNext(ct.sim) != time.Second || wand.TimeSinceLast(ct.sim) != time.Second || wand.SwingSpeed() != 2*time.Second {
			t.Fatalf("at 3.5 s the wand should be shooting with 1 s to the next shot, 1 s since the last and a 2 s speed, got %t, %s, %s, %s",
				ct.wanding(), wand.TimeToNext(ct.sim), wand.TimeSinceLast(ct.sim), wand.SwingSpeed())
		}

		ct.runUntil(6501 * time.Millisecond)
		if !slices.Equal(ct.ticks, seconds(0.5, 2.5, 4.5, 6.5)) || !ct.wanding() || len(ct.channelEnd) != 0 {
			t.Fatalf("a Shoot channel with nothing else ready should shoot at 0.5 s and then every 2 s and keep going, shots at %v, ended at %v", ct.ticks, ct.channelEnd)
		}

		ct.sim.Cleanup()
		ct.sim.Reset()
		ct.sim.advance(500 * time.Millisecond)
		if got := wand.TimeSinceLast(ct.sim); got != 500*time.Millisecond {
			t.Fatalf("after a reset the time since the last shot should count from the pull, got %s at 0.5 s", got)
		}
	})

	t.Run("the first shot after no spell for a wand speed comes after the wind-up", func(t *testing.T) {
		var nukeAt []time.Duration
		ct, nuke := newSpellShootTest(nukeConfig(&nukeAt))
		ct.start(&timedStubAction{readyAt: time.Minute})
		ct.runUntil(0)
		nuke.Cast(ct.sim, ct.fw.CurrentTarget)
		ct.runUntil(3 * time.Second)
		ct.castAction().Execute(ct.sim)

		ct.runUntil(5 * time.Second)
		if !slices.Equal(ct.ticks, seconds(3.5)) {
			t.Fatalf("Shoot cast at 3 s, 3 s after the last spell, should fire 0.5 s later, shots at %v", ct.ticks)
		}
	})

	t.Run("the first shot after a recent spell comes a wand speed after it", func(t *testing.T) {
		var nukeAt []time.Duration
		ct, nuke := newSpellShootTest(nukeConfig(&nukeAt))
		ct.start(&timedStubAction{readyAt: time.Minute})
		ct.runUntil(0)
		nuke.Cast(ct.sim, ct.fw.CurrentTarget)
		ct.runUntil(time.Second)
		ct.castAction().Execute(ct.sim)

		ct.runUntil(1999 * time.Millisecond)
		if len(ct.ticks) != 0 {
			t.Fatalf("Shoot cast at 1 s, 1 s after a spell, should wait until 2 s, shots at %v", ct.ticks)
		}
		ct.runUntil(3 * time.Second)
		if !slices.Equal(ct.ticks, seconds(2)) {
			t.Fatalf("Shoot cast at 1 s should fire one wand speed after the spell at 0, shots at %v", ct.ticks)
		}
	})

	t.Run("what restarts the wand timer", func(t *testing.T) {
		var castAt, drainTicks []time.Duration
		racial := recordedSpellConfig(910120, &castAt)
		racial.Flags |= SpellFlagNoOnCastComplete
		racial.Cast = CastConfig{CD: Cooldown{Duration: time.Minute}}
		proc := recordedSpellConfig(910121, &castAt)
		proc.Flags = SpellFlagNoOnCastComplete
		hardcast := recordedSpellConfig(910122, &castAt)
		hardcast.Cast = CastConfig{DefaultCast: Cast{GCD: GCDDefault, CastTime: time.Second}}

		for _, c := range []struct {
			name      string
			config    SpellConfig
			shootAt   time.Duration
			firstShot time.Duration
		}{
			{"an off-GCD racial or potion cast at 0", racial, 500 * time.Millisecond, 2 * time.Second},
			{"a proc at 0 does not", proc, 500 * time.Millisecond, time.Second},
			{"a hardcast from its end at 1 s", hardcast, time.Second, 3 * time.Second},
			{"another channel from its end at 5 s", drainConfig(&drainTicks, CastConfig{DefaultCast: Cast{GCD: GCDDefault}}), 5 * time.Second, 7 * time.Second},
		} {
			ct, spell := newSpellShootTest(c.config)
			ct.start(&timedStubAction{readyAt: time.Minute})
			ct.runUntil(0)
			if !spell.Cast(ct.sim, ct.fw.CurrentTarget) {
				t.Fatalf("%s: the spell could not be cast", c.name)
			}
			ct.runUntil(c.shootAt)
			ct.castAction().Execute(ct.sim)
			ct.runUntil(c.firstShot + time.Millisecond)
			if !slices.Equal(ct.ticks, []time.Duration{c.firstShot}) {
				t.Fatalf("%s: Shoot cast at %s should fire at %s, shots at %v", c.name, c.shootAt, c.firstShot, ct.ticks)
			}
		}
	})

	t.Run("an interrupt takes the shot's slot", func(t *testing.T) {
		ct := newShootTest()
		stub := &timedStubAction{readyAt: 3 * time.Second}
		ct.start(stub, ct.channelAction())

		ct.runUntil(4501 * time.Millisecond)
		if !slices.Equal(ct.ticks, seconds(0.5, 2.5)) || ct.damage() != 200 {
			t.Fatalf("Shoot should fire at 0.5 and 2.5 s and give its 4.5 s slot away, shots at %v for %.1f", ct.ticks, ct.damage())
		}
		if !slices.Equal(ct.channelEnd, seconds(4.5)) || !slices.Equal(stub.ranAt, seconds(4.5)) {
			t.Fatalf("an action ready at 3 s should end Shoot and run at the 4.5 s slot, channel ended at %v, action ran at %v", ct.channelEnd, stub.ranAt)
		}
	})

	t.Run("no interrupt when the GCD ends", func(t *testing.T) {
		ct := newShootTest()
		stub := &timedStubAction{readyAt: time.Second}
		ct.start(stub, ct.channelAction())
		ct.runUntil(0)
		ct.fw.SetGCDTimer(ct.sim, 1500*time.Millisecond)

		ct.runUntil(2501 * time.Millisecond)
		if !slices.Equal(ct.ticks, seconds(0.5)) {
			t.Fatalf("Shoot should fire at 0.5 s and give its 2.5 s slot to the waiting action, shots at %v", ct.ticks)
		}
		if !slices.Equal(ct.channelEnd, seconds(2.5)) || !slices.Equal(stub.ranAt, seconds(2.5)) {
			t.Fatalf("the action waiting on the GCD should end Shoot at the 2.5 s slot, not at 1.5 s, channel ended at %v, action ran at %v", ct.channelEnd, stub.ranAt)
		}
	})

	t.Run("an interrupt inside the spell queue window still takes the slot", func(t *testing.T) {
		ct := newShootTest()
		stub := &timedStubAction{readyAt: time.Second}
		ct.start(stub, ct.channelAction())
		ct.runUntil(time.Second)
		ct.fw.SetGCDTimer(ct.sim, 2800*time.Millisecond)
		ct.runUntil(1200 * time.Millisecond)
		ct.fw.ReactToEvent(ct.sim, false, false)
		ct.runUntil(1200 * time.Millisecond)

		ct.runUntil(2501 * time.Millisecond)
		if !slices.Equal(ct.ticks, seconds(0.5)) || !slices.Equal(ct.channelEnd, seconds(2.5)) {
			t.Fatalf("with the GCD 0.3 s out at the 2.5 s slot Shoot should end there with no shot, shots at %v, ended at %v", ct.ticks, ct.channelEnd)
		}
		ct.runUntil(2801 * time.Millisecond)
		if !slices.Equal(stub.ranAt, seconds(2.8)) {
			t.Fatalf("the action should run when the GCD ends at 2.8 s, ran at %v", stub.ranAt)
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

		ct.runUntil(4501 * time.Millisecond)
		if !slices.Equal(ct.ticks, seconds(0.5, 2.5, 4.5)) || len(ct.channelEnd) != 0 {
			t.Fatalf("a Shoot channel allowed to recast should shoot at 0.5, 2.5 and 4.5 s without being cut, shots at %v, ended at %v", ct.ticks, ct.channelEnd)
		}
	})

	t.Run("no cooldown is used while wanding", func(t *testing.T) {
		onGCD := &testCooldown{gcd: GCDDefault, priority: 1}
		potion := &testCooldown{}
		ct := newCooldownShootTest(onGCD, potion)
		channel := ct.rot.newActionChannelSpell(&proto.APLActionChannelSpell{SpellId: ct.spell.ActionID.ToProto(), InterruptIf: constValue("false")})
		ct.start(ct.autocastAction(), channel)

		ct.runUntil(6501 * time.Millisecond)
		if len(potion.castAt) != 0 || len(onGCD.castAt) != 0 {
			t.Fatalf("with Shoot never interrupted no cooldown should be used, the potion at %v, the one on the GCD at %v", potion.castAt, onGCD.castAt)
		}
		if !slices.Equal(ct.ticks, seconds(0.5, 2.5, 4.5, 6.5)) || len(ct.channelEnd) != 0 {
			t.Fatalf("Shoot should keep shooting, shots at %v, ended at %v", ct.ticks, ct.channelEnd)
		}
	})

	t.Run("a ready cooldown above Shoot takes the shot's slot", func(t *testing.T) {
		potion := &testCooldown{}
		ct := newCooldownShootTest(potion)
		ct.start(ct.autocastAction(), ct.channelAction())

		ct.runUntil(6501 * time.Millisecond)
		if !slices.Equal(potion.castAt, seconds(4.5)) {
			t.Fatalf("a potion ready at 3 s should be used at the 4.5 s slot, used at %v", potion.castAt)
		}
		if !slices.Equal(ct.ticks, seconds(0.5, 2.5, 6.5)) || !slices.Equal(ct.channelEnd, seconds(4.5)) || !ct.wanding() {
			t.Fatalf("Shoot should give its 4.5 s slot to the potion and start again, shooting a wand speed after it, shots at %v, ended at %v", ct.ticks, ct.channelEnd)
		}
	})

	t.Run("cooldowns take one slot each", func(t *testing.T) {
		onGCD := &testCooldown{gcd: GCDDefault, priority: 1}
		potion := &testCooldown{}
		ct := newCooldownShootTest(onGCD, potion)
		if ct.fw.majorCooldownManager.majorCooldowns[0].Spell != onGCD.spell {
			t.Fatalf("the cooldown on the GCD should sort ahead of the potion")
		}
		ct.start(ct.autocastAction(), ct.channelAction())

		ct.runUntil(8501 * time.Millisecond)
		if !slices.Equal(onGCD.castAt, seconds(4.5)) || !slices.Equal(potion.castAt, seconds(6.5)) {
			t.Fatalf("the one on the GCD should take the 4.5 s slot and the potion, held by that GCD, the 6.5 s slot: the one on the GCD at %v, the potion at %v", onGCD.castAt, potion.castAt)
		}
		if !slices.Equal(ct.ticks, seconds(0.5, 2.5, 8.5)) || !slices.Equal(ct.channelEnd, seconds(4.5, 6.5)) {
			t.Fatalf("Shoot should shoot at 0.5, 2.5 and 8.5 s around the two cooldowns, shots at %v, ended at %v", ct.ticks, ct.channelEnd)
		}
	})

	t.Run("a cast fires its shot with another action ready", func(t *testing.T) {
		ct := newShootTest()
		stub := &timedStubAction{readyAt: 200 * time.Millisecond}
		ct.start(stub, ct.castAction())

		ct.runUntil(501 * time.Millisecond)
		if !slices.Equal(ct.ticks, seconds(0.5)) || !slices.Equal(stub.ranAt, seconds(0.5)) || !slices.Equal(ct.channelEnd, seconds(0.5)) {
			t.Fatalf("a cast Shoot should fire at 0.5 s although an action is ready from 0.2 s, then end and let it run, shots at %v, action ran at %v, ended at %v", ct.ticks, stub.ranAt, ct.channelEnd)
		}
		ct.runUntil(2501 * time.Millisecond)
		if !slices.Equal(ct.ticks, seconds(0.5, 2.5)) {
			t.Fatalf("the cast Shoot after the action should fire a wand speed after the last shot, shots at %v", ct.ticks)
		}
	})

	t.Run("a cast Shoot fires on the wand speed with a clip delay set", func(t *testing.T) {
		ct := newShootTest()
		ct.fw.ChannelClipDelay = 150 * time.Millisecond
		ct.start(ct.castAction())

		ct.runUntil(4501 * time.Millisecond)
		if !slices.Equal(ct.ticks, seconds(0.5, 2.5, 4.5)) {
			t.Fatalf("Shoot cast after each shot should fire every 2 s with no clip delay, shots at %v", ct.ticks)
		}
	})

	t.Run("starting Shoot ends the rotation pass", func(t *testing.T) {
		var nukeAt []time.Duration
		ct, nuke := newSpellShootTest(nukeConfig(&nukeAt))
		ct.start(ct.channelAction(), ct.rot.newActionCastSpell(&proto.APLActionCastSpell{SpellId: nuke.ActionID.ToProto()}))

		ct.runUntil(0)
		if !ct.wanding() || len(nukeAt) != 0 {
			t.Fatalf("at 0 only Shoot should start, wanding %t, spell on the GCD cast at %v", ct.wanding(), nukeAt)
		}
		ct.runUntil(2501 * time.Millisecond)
		if !slices.Equal(ct.ticks, seconds(0.5, 2.5)) || len(nukeAt) != 0 {
			t.Fatalf("Shoot above the spell should keep shooting, shots at %v, spell cast at %v", ct.ticks, nukeAt)
		}
	})

	t.Run("a spell after an instant and Shoot waits for the shot slot", func(t *testing.T) {
		var nukeAt []time.Duration
		ct, nuke := newSpellShootTest(nukeConfig(&nukeAt))
		ct.start(ct.rot.newActionCastSpell(&proto.APLActionCastSpell{SpellId: nuke.ActionID.ToProto()}), ct.channelAction())

		ct.runUntil(0)
		if !slices.Equal(nukeAt, seconds(0)) || !ct.wanding() {
			t.Fatalf("at 0 the instant and then Shoot should start in one pass, instant at %v, wanding %t", nukeAt, ct.wanding())
		}
		ct.runUntil(2001 * time.Millisecond)
		if !slices.Equal(nukeAt, seconds(0, 2)) || len(ct.ticks) != 0 || !slices.Equal(ct.channelEnd, seconds(2)) {
			t.Fatalf("the next instant should wait past its GCD at 1.5 s for the 2 s slot and take it, instants at %v, shots at %v, ended at %v", nukeAt, ct.ticks, ct.channelEnd)
		}
	})

	t.Run("Shoot above a normal channel runs alone", func(t *testing.T) {
		var drainTicks []time.Duration
		ct, drain := newSpellShootTest(drainConfig(&drainTicks, CastConfig{DefaultCast: Cast{GCD: GCDDefault}}))
		stub := &timedStubAction{readyAt: 3 * time.Second}
		drainAction := ct.rot.newActionChannelSpell(&proto.APLActionChannelSpell{SpellId: drain.ActionID.ToProto(), InterruptIf: constValue("true")})
		ct.start(stub, ct.channelAction(), drainAction)

		ct.runUntil(0)
		if dot := ct.fw.ChanneledDot; dot == nil || dot.Spell != ct.spell || drain.Dot(ct.fw.CurrentTarget).IsActive() {
			t.Fatalf("at 0 Shoot should be the only channel")
		}
		ct.runUntil(4501 * time.Millisecond)
		if !slices.Equal(ct.ticks, seconds(0.5, 2.5)) || len(drainTicks) != 0 {
			t.Fatalf("only Shoot should run, shots at %v, Drain Life ticks at %v", ct.ticks, drainTicks)
		}
		if !slices.Equal(ct.channelEnd, seconds(4.5)) || !slices.Equal(stub.ranAt, seconds(4.5)) {
			t.Fatalf("Shoot should end once, at 4.5 s for the action ready at 3 s, ended at %v, action ran at %v", ct.channelEnd, stub.ranAt)
		}
	})

	t.Run("a failed channel leaves no interrupt condition to a cast Shoot", func(t *testing.T) {
		var drainTicks []time.Duration
		ct, drain := newSpellShootTest(drainConfig(&drainTicks, CastConfig{DefaultCast: Cast{GCD: GCDDefault}, CD: Cooldown{Duration: time.Minute}}))
		ct.start(&timedStubAction{readyAt: time.Minute})
		ct.runUntil(0)
		drain.CD.Set(time.Minute)

		ct.rot.newActionChannelSpell(&proto.APLActionChannelSpell{SpellId: drain.ActionID.ToProto(), InterruptIf: constValue("true")}).Execute(ct.sim)
		if ct.fw.ChanneledDot != nil {
			t.Fatalf("Drain Life on cooldown should not start")
		}
		ct.castAction().Execute(ct.sim)

		ct.runUntil(2501 * time.Millisecond)
		if !slices.Equal(ct.ticks, seconds(0.5)) || !slices.Equal(ct.channelEnd, seconds(0.5)) {
			t.Fatalf("the cast Shoot should fire its one shot at 0.5 s and end, shots at %v, ended at %v", ct.ticks, ct.channelEnd)
		}
	})

	t.Run("the rotation wakes at the GCD after Shoot ends inside it", func(t *testing.T) {
		ct := newShootTest()
		ct.start(ct.castAction())
		ct.runUntil(0)
		ct.fw.SetGCDTimer(ct.sim, 3*time.Second)
		ct.runUntil(200 * time.Millisecond)
		ct.fw.ReactToEvent(ct.sim, false, false)
		ct.runUntil(200 * time.Millisecond)

		ct.runUntil(3501 * time.Millisecond)
		if !slices.Equal(ct.ticks, seconds(0.5, 3.5)) {
			t.Fatalf("after the 0.5 s shot ends Shoot inside the GCD, the rotation should cast it again when the GCD ends at 3 s, shots at %v", ct.ticks)
		}
	})

	t.Run("moving ends Shoot with no shot", func(t *testing.T) {
		for _, moveAt := range seconds(0.25, 1) {
			ct := newShootTest()
			ct.start(ct.channelAction())
			ct.runUntil(moveAt)
			before := len(ct.ticks)
			ct.fw.MoveDuration(1500*time.Millisecond, ct.sim)
			if ct.wanding() || !slices.Equal(ct.channelEnd, []time.Duration{moveAt}) {
				t.Fatalf("moving at %s should end Shoot at once, ended at %v", moveAt, ct.channelEnd)
			}

			moveEnd := moveAt + 1500*time.Millisecond
			ct.runUntil(moveEnd - time.Millisecond)
			if ct.wanding() || len(ct.ticks) != before {
				t.Fatalf("Shoot should not start or shoot while moving from %s, shots at %v", moveAt, ct.ticks)
			}
			ct.runUntil(moveEnd + time.Second)
			if !ct.wanding() || len(ct.ticks) != before+1 || ct.ticks[before] < moveEnd+500*time.Millisecond {
				t.Fatalf("after the move ending at %s the rotation should start Shoot again, shots at %v", moveEnd, ct.ticks)
			}
		}
	})

	t.Run("a unit without a wand reads zero", func(t *testing.T) {
		sim := &Simulation{CurrentTime: 3 * time.Second}
		var wand *Wand
		if wand.TimeToNext(sim) != 0 || wand.TimeSinceLast(sim) != 0 || wand.SwingSpeed() != 0 {
			t.Fatalf("a nil wand should read zero")
		}
	})
}

func TestValueWand(t *testing.T) {
	type reads struct {
		swing, toNext, sinceLast time.Duration
	}
	read := func(ct *channelTest) reads {
		rot, sim := ct.rot, ct.sim
		return reads{
			swing:     rot.newValueAutoSwingTime(&proto.APLValueAutoSwingTime{AutoType: proto.APLValueAutoSwingTime_Wand}, nil).GetDuration(sim),
			toNext:    rot.newValueAutoTimeToNext(&proto.APLValueAutoTimeToNext{AutoType: proto.APLValueAutoAttackType_WandAuto}, nil).GetDuration(sim),
			sinceLast: rot.newValueAutoTimeSinceLast(&proto.APLValueAutoTimeSinceLast{AutoType: proto.APLValueAutoAttackType_WandAuto}, nil).GetDuration(sim),
		}
	}

	t.Run("mid channel", func(t *testing.T) {
		ct := newShootTest()
		ct.start(ct.channelAction())
		ct.runUntil(3 * time.Second)
		if got, want := read(ct), (reads{2 * time.Second, 1500 * time.Millisecond, 500 * time.Millisecond}); got != want {
			t.Fatalf("at 3 s into a Shoot channel the wand values read %+v, want %+v", got, want)
		}
	})

	t.Run("Any includes the wand", func(t *testing.T) {
		ct := newShootTest()
		ct.start(ct.channelAction())
		ct.runUntil(3 * time.Second)
		toNext := ct.rot.newValueAutoTimeToNext(&proto.APLValueAutoTimeToNext{}, nil).GetDuration(ct.sim)
		sinceLast := ct.rot.newValueAutoTimeSinceLast(&proto.APLValueAutoTimeSinceLast{}, nil).GetDuration(ct.sim)
		if toNext != 1500*time.Millisecond || sinceLast != 500*time.Millisecond {
			t.Fatalf("with the melee swings off, Any should read the wand at 3 s into a Shoot channel: %s to the next, %s since the last, want 1.5 s and 0.5 s", toNext, sinceLast)
		}
	})

	t.Run("swing time at the current cast speed", func(t *testing.T) {
		ct := newShootTest()
		ct.start()
		ct.fw.MultiplyCastSpeed(ct.sim, 1.1)
		if got := read(ct).swing; got != 1818*time.Millisecond {
			t.Fatalf("with 10%% cast speed the wand swing time reads %s, want 1.818 s", got)
		}
	})

	t.Run("swing time while wanding stays the running speed", func(t *testing.T) {
		ct := newShootTest()
		ct.start(ct.channelAction())
		ct.runUntil(3 * time.Second)
		ct.fw.MultiplyCastSpeed(ct.sim, 1.1)
		if got := read(ct).swing; got != 2*time.Second {
			t.Fatalf("with cast speed gained mid-channel the wand swing time reads %s while wanding, want the running 2 s", got)
		}
		ct.fw.MoveDuration(time.Second, ct.sim)
		if ct.wanding() {
			t.Fatalf("moving should end Shoot")
		}
		if got := read(ct).swing; got != 1818*time.Millisecond {
			t.Fatalf("after the channel ended the wand swing time reads %s, want the hasted 1.818 s", got)
		}
	})

	t.Run("before any cast", func(t *testing.T) {
		ct := newShootTest()
		ct.start()
		if got := read(ct); got != (reads{swing: 2 * time.Second}) {
			t.Fatalf("before any cast the wand reads %+v, want only the 2 s swing time", got)
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
