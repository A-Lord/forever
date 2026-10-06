package core

import "time"

// The wand's Shoot channel, its speed and when its last shot fired, read by the APL's wand values.
type Wand struct {
	shoot      *Spell
	swingSpeed func() time.Duration // The equipped wand's speed; zero with no wand.
	lastShotAt time.Duration
	timerAt    time.Duration // When the wand's timer last restarted: another spell's cast or channel end, or a shot.
}

const wandWindup = 500 * time.Millisecond

func (unit *Unit) EnableWand(shoot *Spell, swingSpeed func() time.Duration) {
	wand := &Wand{shoot: shoot, swingSpeed: swingSpeed}
	unit.Wand = wand
	unit.RegisterResetEffect(func(_ *Simulation) {
		wand.lastShotAt = 0
		wand.timerAt = -NeverExpires
	})
	for _, dot := range append(DotArray{shoot.aoeDot}, shoot.dots...) {
		if dot != nil {
			dot.firstTickAt = wand.firstShotAt
		}
	}
	unit.OnChannelEnd(func(sim *Simulation, dot *Dot) {
		wand.spellCompleted(sim, dot.Spell)
	})
	// Starting to move ends wanding, with no shot.
	unit.RegisterMovementCallback(func(sim *Simulation, _ float64, kind MovementUpdateType) {
		if dot := shoot.autorepeatChannel(); dot != nil && kind == MovementStart {
			dot.stopChannel(sim)
		}
	})
}

func (wand *Wand) recordShot(sim *Simulation) {
	wand.lastShotAt = sim.CurrentTime
	wand.timerAt = sim.CurrentTime
}

// A completed cast or channel of any spell but Shoot, from the real cast paths only: procs and
// autos leave the wand's timer alone.
func (wand *Wand) spellCompleted(sim *Simulation, spell *Spell) {
	if wand != nil && spell != wand.shoot {
		wand.timerAt = sim.CurrentTime
	}
}

// The first shot of a Shoot started now: after the wind-up, and one wand speed after the timer last
// restarted.
func (wand *Wand) firstShotAt(sim *Simulation, dot *Dot) time.Duration {
	return max(sim.CurrentTime+wandWindup, wand.timerAt+dot.tickPeriod)
}

func (wand *Wand) nextShotAt() (time.Duration, bool) {
	if wand == nil {
		return 0, false
	}
	if dot := wand.shoot.autorepeatChannel(); dot != nil {
		return dot.NextTickAt(), true
	}
	return 0, false
}

// Zero while idle. The first shot of a Shoot cast comes after the wind-up, and no sooner than one wand speed after the wand's timer last restarted.
func (wand *Wand) TimeToNext(sim *Simulation) time.Duration {
	if at, ok := wand.nextShotAt(); ok {
		return max(0, at-sim.CurrentTime)
	}
	return 0
}

func (wand *Wand) TimeSinceLast(sim *Simulation) time.Duration {
	if wand == nil {
		return 0
	}
	return sim.CurrentTime - wand.lastShotAt
}

// The running Shoot's shot interval, fixed when it started; otherwise the wand speed at the current cast
// speed, the interval of a Shoot started now. Zero with no wand.
func (wand *Wand) SwingSpeed() time.Duration {
	if wand == nil {
		return 0
	}
	if dot := wand.shoot.autorepeatChannel(); dot != nil {
		return dot.tickPeriod
	}
	speed := wand.swingSpeed()
	if speed == 0 {
		return 0
	}
	return wand.shoot.Unit.ApplyCastSpeedForSpell(speed, wand.shoot).Round(time.Millisecond)
}
