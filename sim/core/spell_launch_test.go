package core

import (
	"testing"
	"time"
)

type launchEvent struct {
	kind string
	at   time.Duration
}

type launchTest struct {
	*channelTest
	missile *Spell
	silent  *Spell
	heal    *Spell
	events  []launchEvent
}

func launchSpellConfig(id int32, flags SpellFlag) SpellConfig {
	return SpellConfig{
		ActionID:         ActionID{SpellID: id},
		Flags:            flags,
		SpellSchool:      SpellSchoolFire,
		DefenseType:      DefenseTypeMagic,
		ProcMask:         ProcMaskSpellDamage,
		MissileSpeed:     20,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
	}
}

// A missile that travels for a second, the same missile with SpellFlagNoOnDamageDealt, a heal, and
// an aura that records the missile's launches and hits. register adds what a test needs on top.
func newLaunchTest(t *testing.T, register func(lt *launchTest, char *Character)) *launchTest {
	t.Helper()
	lt := &launchTest{}
	lt.channelTest = newChannelTest(func(ct *channelTest, char *Character) *Spell {
		lt.missile = char.RegisterSpell(launchSpellConfig(910201, SpellFlagNone))
		lt.silent = char.RegisterSpell(launchSpellConfig(910202, SpellFlagNoOnDamageDealt))
		lt.heal = char.RegisterSpell(SpellConfig{
			ActionID:         ActionID{SpellID: 910203},
			Flags:            SpellFlagHelpful,
			SpellSchool:      SpellSchoolHoly,
			ProcMask:         ProcMaskSpellHealing,
			DamageMultiplier: 1,
			ThreatMultiplier: 1,
		})
		record := func(kind string) OnSpellHit {
			return func(_ *Aura, sim *Simulation, spell *Spell, _ *SpellResult) {
				if spell == lt.missile || spell == lt.silent || spell == lt.heal {
					lt.events = append(lt.events, launchEvent{kind, sim.CurrentTime})
				}
			}
		}
		char.RegisterAura(Aura{
			Label:           "Launch Recorder",
			Duration:        NeverExpires,
			OnReset:         func(aura *Aura, sim *Simulation) { aura.Activate(sim) },
			OnSpellLaunch:   record("launch"),
			OnSpellHitDealt: record("hit"),
			OnHealDealt:     record("heal"),
		})
		if register != nil {
			register(lt, char)
		}
		return lt.missile
	})
	lt.silent.finalize()
	lt.heal.finalize()
	lt.start()
	lt.fw.DistanceFromTarget = 20
	if lt.missile.TravelTime() != time.Second {
		t.Fatalf("missile travel time %v, want %v", lt.missile.TravelTime(), time.Second)
	}
	return lt
}

func (lt *launchTest) target() *Unit {
	return lt.fw.CurrentTarget
}

func (lt *launchTest) result(spell *Spell, outcome HitOutcome) *SpellResult {
	result := spell.NewResult(lt.target())
	result.Outcome = outcome
	result.Damage = 100
	return result
}

func (lt *launchTest) count(kind string) int {
	n := 0
	for _, event := range lt.events {
		if event.kind == kind {
			n++
		}
	}
	return n
}

func wantEvents(t *testing.T, got []launchEvent, want ...launchEvent) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("events %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("events %v, want %v", got, want)
		}
	}
}

// A result launched at the cast and dealt after travel launches at the cast and not again at impact.
func TestLaunchFiresOnceAtTheLaunch(t *testing.T) {
	lt := newLaunchTest(t, nil)
	lt.runUntil(time.Second)

	result := lt.result(lt.missile, OutcomeHit)
	lt.missile.Launch(lt.sim, result)
	lt.missile.Launch(lt.sim, result)
	lt.missile.WaitTravelTime(lt.sim, func(sim *Simulation) { lt.missile.DealDamage(sim, result) })
	lt.runUntil(3 * time.Second)

	wantEvents(t, lt.events, launchEvent{"launch", time.Second}, launchEvent{"hit", 2 * time.Second})
}

// The travel helper launches at the cast and deals the hit at impact.
func TestDealDamageAfterTravelLaunchesAtTheCast(t *testing.T) {
	lt := newLaunchTest(t, nil)
	lt.runUntil(time.Second)

	lt.missile.DealDamageAfterTravel(lt.sim, lt.result(lt.missile, OutcomeCrit))
	lt.runUntil(3 * time.Second)

	wantEvents(t, lt.events, launchEvent{"launch", time.Second}, launchEvent{"hit", 2 * time.Second})
}

// A result dealt without a launch of its own launches right before its hit, at the same moment.
func TestDealDamageLaunchesRightBeforeTheHit(t *testing.T) {
	lt := newLaunchTest(t, nil)
	lt.runUntil(time.Second)

	lt.missile.DealDamage(lt.sim, lt.result(lt.missile, OutcomeHit))

	wantEvents(t, lt.events, launchEvent{"launch", time.Second}, launchEvent{"hit", time.Second})
}

// Ticks and heals never launch, and SpellFlagNoOnDamageDealt keeps a direct hit silent.
func TestLaunchSkipsTicksHealsAndSilentSpells(t *testing.T) {
	lt := newLaunchTest(t, nil)

	lt.missile.DealPeriodicDamage(lt.sim, lt.result(lt.missile, OutcomeHit))
	heal := lt.heal.NewResult(&lt.fw.Unit)
	heal.Outcome = OutcomeHit
	heal.Damage = 100
	lt.heal.DealHealing(lt.sim, heal)
	lt.silent.DealDamage(lt.sim, lt.result(lt.silent, OutcomeHit))
	lt.silent.DealDamageAfterTravel(lt.sim, lt.result(lt.silent, OutcomeHit))
	lt.runUntil(2 * time.Second)

	if got := lt.count("launch"); got != 0 {
		t.Errorf("launches %d, want none", got)
	}
	if got := lt.count("heal"); got != 1 {
		t.Errorf("heals %d, want the one heal dealt", got)
	}
}

// A pooled result handed out again after it was dealt starts unlaunched.
func TestReusedResultLaunchesAgain(t *testing.T) {
	lt := newLaunchTest(t, nil)

	first := lt.result(lt.missile, OutcomeHit)
	lt.missile.Launch(lt.sim, first)
	lt.missile.DealDamage(lt.sim, first)
	second := lt.result(lt.missile, OutcomeHit)
	if second != first {
		t.Fatalf("the result cache handed out a new result, want the disposed one back")
	}
	if second.launched {
		t.Fatalf("a reused result is still marked launched")
	}
	lt.missile.DealDamage(lt.sim, second)

	if got := lt.count("launch"); got != 2 {
		t.Errorf("launches %d, want one per dealt result", got)
	}
}

// A second registration under the same label replaces the launch callback.
func TestGetOrRegisterAuraCarriesOnSpellLaunch(t *testing.T) {
	var first, second int
	lt := newLaunchTest(t, func(_ *launchTest, char *Character) {
		char.GetOrRegisterAura(Aura{
			Label:         "Relaunch",
			Duration:      NeverExpires,
			OnReset:       func(aura *Aura, sim *Simulation) { aura.Activate(sim) },
			OnSpellLaunch: func(_ *Aura, _ *Simulation, _ *Spell, _ *SpellResult) { first++ },
		})
		char.GetOrRegisterAura(Aura{
			Label:         "Relaunch",
			OnSpellLaunch: func(_ *Aura, _ *Simulation, _ *Spell, _ *SpellResult) { second++ },
		})
	})
	lt.missile.DealDamage(lt.sim, lt.result(lt.missile, OutcomeHit))

	if first != 0 || second != 1 {
		t.Errorf("first callback ran %d times and second %d, want only the second, once", first, second)
	}
}

func launchTrigger(name string, fired *[]launchEvent, config ProcTrigger) func(*launchTest, *Character) {
	return func(lt *launchTest, char *Character) {
		config.Name = name
		config.Callback = CallbackOnSpellLaunch
		config.ProcMask = ProcMaskSpellDamage
		config.Handler = func(sim *Simulation, _ *Spell, result *SpellResult) {
			kind := "fired"
			if result == nil || !result.Landed() {
				kind = "fired without a landed result"
			}
			*fired = append(*fired, launchEvent{kind, sim.CurrentTime})
		}
		char.MakeProcTriggerAura(config)
	}
}

// A launch trigger takes its outcome filter from the launched result.
func TestLaunchTriggerHonoursOutcome(t *testing.T) {
	var fired []launchEvent
	lt := newLaunchTest(t, launchTrigger("Launch Outcome", &fired, ProcTrigger{Outcome: OutcomeLanded, TriggerImmediately: true}))

	lt.missile.DealDamageAfterTravel(lt.sim, lt.result(lt.missile, OutcomeMiss))
	lt.missile.DealDamageAfterTravel(lt.sim, lt.result(lt.missile, OutcomeHit))
	lt.runUntil(2 * time.Second)

	wantEvents(t, fired, launchEvent{"fired", 0})
}

// A launch trigger's cooldown starts at the launch it fires on.
func TestLaunchTriggerHonoursICD(t *testing.T) {
	var fired []launchEvent
	lt := newLaunchTest(t, launchTrigger("Launch ICD", &fired, ProcTrigger{ICD: 5 * time.Second, TriggerImmediately: true}))

	for _, at := range seconds(1, 3, 6) {
		lt.runUntil(at)
		lt.missile.DealDamageAfterTravel(lt.sim, lt.result(lt.missile, OutcomeHit))
	}
	lt.runUntil(8 * time.Second)

	wantEvents(t, fired, launchEvent{"fired", time.Second}, launchEvent{"fired", 6 * time.Second})
}

// A launch trigger rolls its chance on each launch.
func TestLaunchTriggerHonoursChance(t *testing.T) {
	var never, always []launchEvent
	lt := newLaunchTest(t, func(lt *launchTest, char *Character) {
		launchTrigger("Launch Never", &never, ProcTrigger{ProcChance: 1e-6, TriggerImmediately: true})(lt, char)
		launchTrigger("Launch Always", &always, ProcTrigger{TriggerImmediately: true})(lt, char)
	})

	const launches = 200
	for range launches {
		lt.missile.Launch(lt.sim, lt.result(lt.missile, OutcomeHit))
	}

	if len(never) != 0 {
		t.Errorf("a near-zero chance fired %d times, want none", len(never))
	}
	if len(always) != launches {
		t.Errorf("a certain chance fired %d times, want every launch", len(always))
	}
}

// An immediate launch trigger fires at the launch, a deferred one a spell batch later with a copy of
// the result, and both before the missile lands.
func TestLaunchTriggerImmediateAndDeferred(t *testing.T) {
	var immediate, deferred []launchEvent
	lt := newLaunchTest(t, func(lt *launchTest, char *Character) {
		launchTrigger("Launch Immediate", &immediate, ProcTrigger{TriggerImmediately: true})(lt, char)
		launchTrigger("Launch Deferred", &deferred, ProcTrigger{})(lt, char)
	})
	lt.runUntil(time.Second)

	lt.missile.DealDamageAfterTravel(lt.sim, lt.result(lt.missile, OutcomeHit))
	lt.runUntil(3 * time.Second)

	wantEvents(t, immediate, launchEvent{"fired", time.Second})
	wantEvents(t, deferred, launchEvent{"fired", time.Second + SpellBatchWindow})
	wantEvents(t, lt.events, launchEvent{"launch", time.Second}, launchEvent{"hit", 2 * time.Second})
}
