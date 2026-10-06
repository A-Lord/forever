package core

import "time"

// The wand's Shoot channel, its speed and when its last shot fired, read by the APL's wand values.
type Wand struct {
	Shoot      *Spell
	swingSpeed time.Duration
	lastShotAt time.Duration
}

func (unit *Unit) EnableWand(shoot *Spell, swingSpeed time.Duration) {
	wand := &Wand{Shoot: shoot, swingSpeed: swingSpeed}
	unit.Wand = wand
	unit.RegisterResetEffect(func(_ *Simulation) {
		wand.lastShotAt = 0
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
