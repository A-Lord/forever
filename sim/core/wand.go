package core

import "time"

// Shoot's scheduler. Shoot applies a channel that carries the channeling state (Is Channeling, the
// rotation's hold, the timeline bar) but never ticks; the wand fires the shots at its own slots.
type Wand struct {
	unit       *Unit
	shoot      *Spell
	swingSpeed func() time.Duration // The equipped wand's speed; zero with no wand.
	shot       func(sim *Simulation, target *Unit)

	target      *Unit
	interruptIf APLValue      // The channel action's condition; nil for a cast, which fires one shot.
	interval    time.Duration // Fixed when Shoot starts.
	slot        *PendingAction
	lastShotAt  time.Duration
	timerAt     time.Duration // When the wand's timer last restarted: a rotation cast, another channel's end, or a shot.
}

const wandWindup = 500 * time.Millisecond

// Shoot's channel. Wand.Start sets its length to the rest of the fight; ChanneledTicks(Shoot) reads 0,
// since the shots are not its ticks.
func WandChannel() DotConfig {
	return DotConfig{
		Aura:          Aura{Label: "Shoot"},
		NumberOfTicks: 1,
		TickLength:    wandWindup,
	}
}

// Makes shoot the unit's wand Shoot, which fires shot at each of its slots. Shoot's ApplyEffects calls
// Start; Shoot cannot be cast while it runs.
func (unit *Unit) EnableWand(shoot *Spell, swingSpeed func() time.Duration, shot func(sim *Simulation, target *Unit)) *Wand {
	wand := &Wand{unit: unit, shoot: shoot, swingSpeed: swingSpeed, shot: shot}
	unit.Wand = wand
	unit.RegisterResetEffect(func(_ *Simulation) {
		wand.lastShotAt = 0
		wand.timerAt = -NeverExpires
		wand.interruptIf = nil
	})

	castCondition := shoot.ExtraCastCondition
	shoot.ExtraCastCondition = func(sim *Simulation, target *Unit) bool {
		return !wand.wanding() && (castCondition == nil || castCondition(sim, target))
	}

	unit.OnChannelEnd(func(sim *Simulation, dot *Dot) {
		if dot.Spell != shoot {
			wand.timerAt = sim.CurrentTime
			return
		}
		wand.slot.Cancel(sim)
		wand.interruptIf = nil
		// The rotation acts at the slot with no clip delay, or when the GCD ends: a wake-up that came
		// while it was held is spent.
		if unit.GCD.IsReady(sim) {
			unit.WaitUntil(sim, sim.CurrentTime)
		} else if unit.rotationAction.IsConsumed() || unit.NextRotationActionAt() > unit.GCD.ReadyAt() {
			unit.WaitUntil(sim, unit.GCD.ReadyAt())
		}
	})

	// Starting to move ends wanding, with no shot.
	unit.RegisterMovementCallback(func(sim *Simulation, _ float64, kind MovementUpdateType) {
		if kind == MovementStart && wand.wanding() {
			unit.ChanneledDot.Deactivate(sim)
		}
	})
	return wand
}

// Starts wanding at target. The first shot comes after the wind-up, and one wand speed after the timer
// last restarted. A channel action's interruptIf is taken before the channel starts, so the rotation
// never holds it and never cuts Shoot between slots.
func (wand *Wand) Start(sim *Simulation, target *Unit) {
	rot := wand.unit.Rotation
	wand.interruptIf = nil
	if action := rot.channelAction; action != nil && action.spell == wand.shoot {
		wand.interruptIf = action.interruptIf
	}
	rot.channelAction = nil

	wand.target = target
	wand.interval = wand.SwingSpeed()
	firstShotAt := max(sim.CurrentTime+wandWindup, wand.timerAt+wand.interval)

	dot := wand.shoot.Dot(target)
	dot.BaseTickLength = max(sim.GetRemainingDuration(), firstShotAt-sim.CurrentTime)
	dot.Apply(sim)
	dot.tickAction.Cancel(sim)
	dot.tickAction.NextActionAt = NeverExpires

	wand.slot = &PendingAction{NextActionAt: firstShotAt, OnAction: wand.fire}
	sim.AddPendingAction(wand.slot)
}

// A slot that an interrupt takes fires no shot. A cast rather than a channel action fires one shot and stops.
func (wand *Wand) fire(sim *Simulation) {
	dot := wand.unit.ChanneledDot
	if wand.interruptIf != nil && wand.slotTaken(sim) {
		dot.Deactivate(sim)
		return
	}

	wand.shot(sim, wand.target)
	wand.lastShotAt = sim.CurrentTime
	wand.timerAt = sim.CurrentTime
	if !wand.wanding() {
		return
	}
	if wand.interruptIf == nil {
		dot.Deactivate(sim)
		return
	}

	wand.slot.NextActionAt = sim.CurrentTime + wand.interval
	// A fight that runs past its expected length keeps the channel going.
	if dot.ExpiresAt() <= wand.slot.NextActionAt {
		dot.UpdateExpires(wand.slot.NextActionAt + wand.interval)
	}
	sim.AddPendingAction(wand.slot)
}

// Whether the rotation takes the slot: interruptIf holds and the rotation, with Shoot ended, would do
// something other than Shoot. With nothing else ready Shoot keeps shooting. Inside a controlling action
// (a strict sequence Shoot started) the slot always goes to it, since asking it for its next action
// moves it on.
func (wand *Wand) slotTaken(sim *Simulation) bool {
	rot := wand.unit.Rotation
	rot.evalGeneration++
	if !wand.interruptIf.GetBool(sim) {
		return false
	}
	if len(rot.controllingActions) != 0 {
		return true
	}

	dot := wand.unit.ChanneledDot
	wand.unit.ChanneledDot = nil
	next := rot.getNextAction(sim)
	wand.unit.ChanneledDot = dot
	return next != nil && !wand.isShootAction(next)
}

func (wand *Wand) isShootAction(action *APLAction) bool {
	switch impl := action.impl.(type) {
	case *APLActionCastSpell:
		return impl.spell == wand.shoot
	case *APLActionChannelSpell:
		return impl.spell == wand.shoot
	}
	return false
}

// A completed cast by the rotation or a major cooldown of any spell but Shoot.
func (wand *Wand) castCompleted(sim *Simulation, spell *Spell) {
	if wand != nil && spell != wand.shoot {
		wand.timerAt = sim.CurrentTime
	}
}

func (wand *Wand) wanding() bool {
	return wand != nil && wand.unit.ChanneledDot != nil && wand.unit.ChanneledDot.Spell == wand.shoot
}

func (wand *Wand) IsShoot(spell *Spell) bool {
	return wand != nil && spell == wand.shoot
}

func (wand *Wand) nextShotAt() (time.Duration, bool) {
	if wand.wanding() {
		return wand.slot.NextActionAt, true
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
	if wand.wanding() {
		return wand.interval
	}
	speed := wand.swingSpeed()
	if speed == 0 {
		return 0
	}
	return wand.shoot.Unit.ApplyCastSpeedForSpell(speed, wand.shoot).Round(time.Millisecond)
}
