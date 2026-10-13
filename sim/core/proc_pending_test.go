package core

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

const (
	testPendingBuffID       = 910301
	testPendingDebuffID     = 910302
	testPendingUndeclaredID = 910303
	testPendingImmediateID  = 910306
	testPendingRolledID     = 910307
	testPendingManagedID    = 910308
)

type pendingProcTest struct {
	*channelTest
	buff         *Aura
	debuffs      AuraArray
	undeclared   *Aura
	immediate    *Aura
	rolled       *Aura
	managed      *Aura
	aplSpell     *Spell
	procSpell    *Spell
	castTriggers int
}

// A spell cast from the APL and one that is not, each a landed hit. Deferred triggers on those hits put up a
// buff and a debuff on the hit target, both declared, and a second buff left undeclared; an immediate trigger
// puts up a third, declared, at the hit. Two more deferred triggers declare a buff each, one rolling a stated
// chance below 1 and one a proc manager that always fires. A cast trigger declares the debuff too, though a
// cast carries no target to find it on.
func newPendingProcTest() *pendingProcTest {
	pt := &pendingProcTest{}
	pt.channelTest = newChannelTest(func(_ *channelTest, char *Character) *Spell {
		hit := func(id int32, flags SpellFlag) *Spell {
			return char.RegisterSpell(SpellConfig{
				ActionID:         ActionID{SpellID: id},
				Flags:            flags,
				SpellSchool:      SpellSchoolFire,
				DefenseType:      DefenseTypeMagic,
				ProcMask:         ProcMaskSpellDamage,
				DamageMultiplier: 1,
				ThreatMultiplier: 1,
				ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
					spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeAlwaysHit)
				},
			})
		}
		pt.aplSpell = hit(910304, SpellFlagAPL)
		pt.procSpell = hit(910305, SpellFlagNone)

		buff := func(label string, id int32) *Aura {
			return char.RegisterAura(Aura{Label: label, ActionID: ActionID{SpellID: id}, Duration: 10 * time.Second})
		}
		pt.buff = buff("Pending Buff", testPendingBuffID)
		pt.undeclared = buff("Undeclared Buff", testPendingUndeclaredID)
		pt.immediate = buff("Immediate Buff", testPendingImmediateID)
		pt.rolled = buff("Rolled Buff", testPendingRolledID)
		pt.managed = buff("Managed Buff", testPendingManagedID)
		pt.debuffs = char.NewEnemyAuraArray(func(target *Unit) *Aura {
			return target.RegisterAura(Aura{Label: "Pending Debuff", ActionID: ActionID{SpellID: testPendingDebuffID}, Duration: 10 * time.Second})
		})

		trigger := func(name string, config ProcTrigger) {
			config.Name = name
			if config.Callback == CallbackEmpty {
				config.Callback = CallbackOnSpellHitDealt
				config.Outcome = OutcomeLanded
			}
			config.ProcMask = ProcMaskSpellDamage
			char.MakeProcTriggerAura(config)
		}
		trigger("Pending Buff Trigger", ProcTrigger{
			ActivatesAura: pt.buff,
			Handler:       func(sim *Simulation, _ *Spell, _ *SpellResult) { pt.buff.Activate(sim) },
		})
		trigger("Pending Debuff Trigger", ProcTrigger{
			ActivatesAuras: pt.debuffs,
			Handler:        func(sim *Simulation, _ *Spell, result *SpellResult) { pt.debuffs.Get(result.Target).Activate(sim) },
		})
		trigger("Immediate Buff Trigger", ProcTrigger{
			TriggerImmediately: true,
			ActivatesAura:      pt.immediate,
			Handler:            func(sim *Simulation, _ *Spell, _ *SpellResult) { pt.immediate.Activate(sim) },
		})
		trigger("Rolled Buff Trigger", ProcTrigger{
			ProcChance:    0.9999999,
			ActivatesAura: pt.rolled,
			Handler:       func(sim *Simulation, _ *Spell, _ *SpellResult) { pt.rolled.Activate(sim) },
		})
		trigger("Managed Buff Trigger", ProcTrigger{
			DPM:           char.NewFixedProcChanceManager(1, ProcMaskSpellDamage),
			ActivatesAura: pt.managed,
			Handler:       func(sim *Simulation, _ *Spell, _ *SpellResult) { pt.managed.Activate(sim) },
		})
		trigger("Undeclared Buff Trigger", ProcTrigger{
			Handler: func(sim *Simulation, _ *Spell, _ *SpellResult) { pt.undeclared.Activate(sim) },
		})
		trigger("Pending Debuff Cast Trigger", ProcTrigger{
			Callback:       CallbackOnCastComplete,
			ActivatesAuras: pt.debuffs,
			Handler:        func(*Simulation, *Spell, *SpellResult) { pt.castTriggers++ },
		})
		return pt.aplSpell
	})
	pt.procSpell.finalize()
	pt.fw.ReactionTime = 100 * time.Millisecond
	pt.start()
	return pt
}

func (pt *pendingProcTest) isActive(id int32, includeReactionTime bool) APLValue {
	return pt.rot.newValueAuraIsActive(&proto.APLValueAuraIsActive{AuraId: ActionID{SpellID: id}.ToProto(), IncludeReactionTime: includeReactionTime}, nil)
}

func (pt *pendingProcTest) isInactive(id int32) APLValue {
	return pt.rot.newValueAuraIsInactive(&proto.APLValueAuraIsInactive{AuraId: ActionID{SpellID: id}.ToProto()}, nil)
}

func (pt *pendingProcTest) debuffIsActive(includeReactionTime bool) APLValue {
	return pt.rot.newValueAuraIsActive(&proto.APLValueAuraIsActive{
		SourceUnit:          &proto.UnitReference{Type: proto.UnitReference_CurrentTarget},
		AuraId:              ActionID{SpellID: testPendingDebuffID}.ToProto(),
		IncludeReactionTime: includeReactionTime,
	}, nil)
}

// Inside the spell batch window a declared aura reads as up to the APL, and an undeclared one does not.
func TestAuraIsActiveSeesADeclaredProcBeforeItsHandlerRuns(t *testing.T) {
	pt := newPendingProcTest()
	active, inactive := pt.isActive(testPendingBuffID, false), pt.isInactive(testPendingBuffID)
	debuffActive := pt.debuffIsActive(false)
	undeclaredActive, undeclaredInactive := pt.isActive(testPendingUndeclaredID, false), pt.isInactive(testPendingUndeclaredID)

	pt.runUntil(time.Second)
	pt.procSpell.SkipCastAndApplyEffects(pt.sim, pt.fw.CurrentTarget)
	if pt.buff.IsActive() {
		t.Fatalf("buff active before its handler ran, want the handler deferred")
	}
	if !active.GetBool(pt.sim) || inactive.GetBool(pt.sim) || !debuffActive.GetBool(pt.sim) {
		t.Errorf("declared in the batch window: active = %t, inactive = %t, debuff active = %t, want true, false, true",
			active.GetBool(pt.sim), inactive.GetBool(pt.sim), debuffActive.GetBool(pt.sim))
	}
	if undeclaredActive.GetBool(pt.sim) || !undeclaredInactive.GetBool(pt.sim) {
		t.Errorf("undeclared in the batch window: active = %t, inactive = %t, want false, true", undeclaredActive.GetBool(pt.sim), undeclaredInactive.GetBool(pt.sim))
	}

	pt.runUntil(time.Second + SpellBatchWindow)
	if !pt.buff.IsActive() || pt.buff.pendingActivations != 0 || !active.GetBool(pt.sim) || inactive.GetBool(pt.sim) {
		t.Errorf("after the handler ran: buff up = %t, pending = %d, active = %t, inactive = %t, want true, 0, true, false",
			pt.buff.IsActive(), pt.buff.pendingActivations, active.GetBool(pt.sim), inactive.GetBool(pt.sim))
	}
}

// Each deferral counts until its handler runs, and the ones still waiting at the end of an iteration do not
// carry into the next.
func TestPendingProcCountsEachDeferralAndClearsAtTheIterationEnd(t *testing.T) {
	pt := newPendingProcTest()
	active := pt.isActive(testPendingBuffID, false)
	pending := func() int32 { return pt.buff.pendingActivations }

	pt.runUntil(time.Second)
	pt.procSpell.SkipCastAndApplyEffects(pt.sim, pt.fw.CurrentTarget)
	pt.runUntil(time.Second + SpellBatchWindow/2)
	pt.procSpell.SkipCastAndApplyEffects(pt.sim, pt.fw.CurrentTarget)
	if pending() != 2 {
		t.Errorf("two handlers waiting: pending = %d, want 2", pending())
	}
	pt.runUntil(time.Second + SpellBatchWindow)
	if pending() != 1 {
		t.Errorf("one handler run, one waiting: pending = %d, want 1", pending())
	}

	pt.sim.Cleanup()
	pt.sim.Reset()
	if pending() != 0 || active.GetBool(pt.sim) {
		t.Errorf("next iteration: pending = %d, active = %t, want 0, false", pending(), active.GetBool(pt.sim))
	}

	pt.start()
	pt.runUntil(time.Second)
	pt.procSpell.SkipCastAndApplyEffects(pt.sim, pt.fw.CurrentTarget)
	pt.runUntil(time.Second + SpellBatchWindow)
	if pending() != 0 || !pt.buff.IsActive() {
		t.Errorf("a proc in the next iteration: pending = %d, buff up = %t, want 0, true", pending(), pt.buff.IsActive())
	}
}

// With the reaction time, a proc off a spell the APL did not cast is a proc: unseen while it waits and for the
// reaction time after it lands.
func TestReactionTimeHidesAPendingProcOffANonAPLSpell(t *testing.T) {
	pt := newPendingProcTest()
	reacted := pt.isActive(testPendingBuffID, true)

	pt.runUntil(time.Second)
	pt.procSpell.SkipCastAndApplyEffects(pt.sim, pt.fw.CurrentTarget)
	if reacted.GetBool(pt.sim) {
		t.Errorf("in the batch window: active = true, want false")
	}
	landed := time.Second + SpellBatchWindow
	pt.runUntil(landed + 50*time.Millisecond)
	if reacted.GetBool(pt.sim) {
		t.Errorf("inside the reaction time after landing: active = true, want false")
	}
	pt.runUntil(landed + 100*time.Millisecond)
	if !reacted.GetBool(pt.sim) {
		t.Errorf("at the reaction time after landing: active = false, want true")
	}
}

// With the reaction time, a declared proc off a spell cast from the APL counts as the player's own: seen while it
// waits and as it lands, and for the reaction time after it runs out. A cast trigger declaring a per-target aura
// marks nothing, having no target to find it on.
func TestReactionTimeCountsAProcOffAnAPLSpellAsOwn(t *testing.T) {
	pt := newPendingProcTest()
	reacted, debuffReacted := pt.isActive(testPendingBuffID, true), pt.debuffIsActive(true)
	debuff := pt.debuffs.Get(pt.fw.CurrentTarget)

	pt.runUntil(time.Second)
	pt.aplSpell.Cast(pt.sim, pt.fw.CurrentTarget)
	if pt.castTriggers != 0 || debuff.pendingActivations != 1 {
		t.Errorf("after the cast: cast triggers run = %d, debuff pending = %d, want 0, 1", pt.castTriggers, debuff.pendingActivations)
	}
	if !reacted.GetBool(pt.sim) || !debuffReacted.GetBool(pt.sim) {
		t.Errorf("in the batch window: active = %t, debuff active = %t, want true, true", reacted.GetBool(pt.sim), debuffReacted.GetBool(pt.sim))
	}
	pt.runUntil(time.Second + SpellBatchWindow)
	if pt.castTriggers != 1 || !reacted.GetBool(pt.sim) || !debuffReacted.GetBool(pt.sim) {
		t.Errorf("as it lands: cast triggers run = %d, active = %t, debuff active = %t, want 1, true, true", pt.castTriggers, reacted.GetBool(pt.sim), debuffReacted.GetBool(pt.sim))
	}

	expires := pt.buff.ExpiresAt()
	pt.runUntil(expires)
	pt.runUntil(expires + 50*time.Millisecond)
	if pt.buff.IsActive() || !reacted.GetBool(pt.sim) {
		t.Errorf("inside the reaction time after running out: buff up = %t, active = %t, want false, true", pt.buff.IsActive(), reacted.GetBool(pt.sim))
	}
	pt.runUntil(expires + 100*time.Millisecond)
	if reacted.GetBool(pt.sim) {
		t.Errorf("at the reaction time after running out: active = true, want false")
	}
}

// With the reaction time, an immediate declared proc off a spell cast from the APL counts as the player's own as
// it lands, and one off a spell the APL did not cast stays a proc.
func TestReactionTimeCountsAnImmediateProcOffAnAPLSpellAsOwn(t *testing.T) {
	pt := newPendingProcTest()
	reacted := pt.isActive(testPendingImmediateID, true)

	pt.runUntil(time.Second)
	pt.aplSpell.SkipCastAndApplyEffects(pt.sim, pt.fw.CurrentTarget)
	if !pt.immediate.IsActive() || !reacted.GetBool(pt.sim) {
		t.Errorf("off the APL spell, as it lands: buff up = %t, active = %t, want true, true", pt.immediate.IsActive(), reacted.GetBool(pt.sim))
	}

	pt.runUntil(pt.immediate.ExpiresAt() + time.Second)
	pt.procSpell.SkipCastAndApplyEffects(pt.sim, pt.fw.CurrentTarget)
	if !pt.immediate.IsActive() || reacted.GetBool(pt.sim) {
		t.Errorf("off the other spell, as it lands: buff up = %t, active = %t, want true, false", pt.immediate.IsActive(), reacted.GetBool(pt.sim))
	}
}

// With the reaction time, a declared proc off a spell cast from the APL that rolls a chance or a proc manager
// stays a proc: unseen while it waits and for the reaction time after it lands.
func TestReactionTimeHidesARolledProcOffAnAPLSpell(t *testing.T) {
	pt := newPendingProcTest()
	auras := map[string]*Aura{"stated chance": pt.rolled, "proc manager": pt.managed}
	reacted := map[string]APLValue{
		"stated chance": pt.isActive(testPendingRolledID, true),
		"proc manager":  pt.isActive(testPendingManagedID, true),
	}

	pt.runUntil(time.Second)
	pt.aplSpell.SkipCastAndApplyEffects(pt.sim, pt.fw.CurrentTarget)
	for name, aura := range auras {
		if aura.pendingActivations != 1 {
			t.Fatalf("%s: pending = %d after the cast, want the proc to fire", name, aura.pendingActivations)
		}
		if reacted[name].GetBool(pt.sim) {
			t.Errorf("%s in the batch window: active = true, want false", name)
		}
	}

	landed := time.Second + SpellBatchWindow
	pt.runUntil(landed + 50*time.Millisecond)
	for name, aura := range auras {
		if !aura.IsActive() || reacted[name].GetBool(pt.sim) {
			t.Errorf("%s inside the reaction time after landing: buff up = %t, active = %t, want true, false", name, aura.IsActive(), reacted[name].GetBool(pt.sim))
		}
	}
	pt.runUntil(landed + 100*time.Millisecond)
	for name := range auras {
		if !reacted[name].GetBool(pt.sim) {
			t.Errorf("%s at the reaction time after landing: active = false, want true", name)
		}
	}
}
