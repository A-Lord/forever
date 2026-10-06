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
}

func (wand *Wand) RecordShot(sim *Simulation) {
	if wand != nil {
		wand.lastShotAt = sim.CurrentTime
	}
}

func (wand *Wand) channel() *Dot {
	if wand == nil || wand.Shoot == nil {
		return nil
	}
	if dot := wand.Shoot.Unit.ChanneledDot; dot != nil && dot.Spell == wand.Shoot {
		return dot
	}
	return nil
}

func (wand *Wand) IsWanding() bool {
	return wand.channel() != nil
}

// Zero while idle: a Shoot cast then starts a full window of its own.
func (wand *Wand) TimeToNext(sim *Simulation) time.Duration {
	if dot := wand.channel(); dot != nil {
		return max(0, dot.NextTickAt()-sim.CurrentTime)
	}
	return 0
}

func (wand *Wand) TimeSinceLast(sim *Simulation) time.Duration {
	if wand == nil {
		return 0
	}
	return sim.CurrentTime - wand.lastShotAt
}

func (wand *Wand) SwingSpeed() time.Duration {
	if wand == nil {
		return 0
	}
	return wand.swingSpeed
}
