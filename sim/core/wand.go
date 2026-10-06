package core

import "time"

// The wand's Shoot channel, its speed and when its last shot fired, read by the APL's wand values.
type Wand struct {
	Shoot      *Spell
	swingSpeed time.Duration
	lastShotAt time.Duration
	timerAt    time.Duration // When the wand's timer last restarted: another spell's cast or channel end, or a shot.
}

const wandWindup = 500 * time.Millisecond

func (unit *Unit) EnableWand(shoot *Spell, swingSpeed time.Duration) {
	wand := &Wand{Shoot: shoot, swingSpeed: swingSpeed, timerAt: -NeverExpires}
	unit.Wand = wand
	unit.RegisterResetEffect(func(_ *Simulation) {
		wand.lastShotAt = 0
		wand.timerAt = -NeverExpires
	})
	// Starting to move ends wanding, with no shot.
	unit.RegisterMovementCallback(func(sim *Simulation, _ float64, kind MovementUpdateType) {
		if dot := shoot.autorepeatChannel(); dot != nil && kind == MovementStart {
			dot.tickAction.NextActionAt = NeverExpires
			dot.Deactivate(sim)
		}
	})
}

func (wand *Wand) RecordShot(sim *Simulation) {
	wand.lastShotAt = sim.CurrentTime
	wand.timerAt = sim.CurrentTime
}

// A completed cast or channel of any spell but Shoot, from the real cast paths only: procs and
// autos leave the wand's timer alone.
func (wand *Wand) spellCompleted(sim *Simulation, spell *Spell) {
	if wand != nil && spell != wand.Shoot {
		wand.timerAt = sim.CurrentTime
	}
}

// The first shot of a Shoot started now: after the wind-up, and one wand speed after the timer last
// restarted.
func (wand *Wand) firstShotAt(sim *Simulation, period time.Duration) time.Duration {
	return max(sim.CurrentTime+wandWindup, wand.timerAt+period)
}

func (wand *Wand) nextShotAt() (time.Duration, bool) {
	if wand == nil {
		return 0, false
	}
	if dot := wand.Shoot.autorepeatChannel(); dot != nil {
		return dot.NextTickAt(), true
	}
	return 0, false
}

// Zero while idle: a Shoot cast then starts a full window of its own.
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

// The wand speed at the current cast speed, the shot interval of a Shoot started now.
func (wand *Wand) SwingSpeed() time.Duration {
	if wand == nil {
		return 0
	}
	return wand.Shoot.Unit.ApplyCastSpeedForSpell(wand.swingSpeed, wand.Shoot).Round(time.Millisecond)
}
