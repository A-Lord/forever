//go:build with_db

package feralcat

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func TestFeralFormsUseFeralSkillAndRestoreEquippedSkill(t *testing.T) {
	sim, cat := newCatWearing(t, "", nil, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: 866})
	cat.ClearForm(sim)
	cat.PseudoStats.WeaponSkills[proto.WeaponSkill_WeaponSkillStaves] = -20
	cat.SetFeralCombat(false, 0)
	table := cat.AttackTables[cat.CurrentTarget.UnitIndex]
	if got := table.MH.BaseMissChance; got != 0.09 {
		t.Fatalf("humanoid staff miss = %v, want 0.09", got)
	}

	fillResource(sim, cat)
	if !cat.CatForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Cat Form did not cast")
	}
	if got := table.MH.BaseMissChance; got != 0.05 {
		t.Fatalf("Cat Form miss = %v, want 0.05", got)
	}

	cat.ClearForm(sim)
	sim.CurrentTime += 2 * time.Second
	fillResource(sim, cat)
	if !cat.BearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Bear Form did not cast")
	}
	if got := table.MH.BaseMissChance; got != 0.05 {
		t.Fatalf("Bear Form miss = %v, want 0.05", got)
	}

	cat.ClearForm(sim)
	if got := table.MH.BaseMissChance; got != 0.09 {
		t.Fatalf("restored humanoid staff miss = %v, want 0.09", got)
	}
}

func TestFeralAttackPowerConvertsOneForOneOnlyInFeralForms(t *testing.T) {
	// Given a cat with 77 Feral Attack Power added after initialization.
	sim, cat := newCat(t, "", nil)
	catAP := cat.GetStat(stats.AttackPower)
	cat.AddStatDynamic(sim, stats.FeralAttackPower, 77)

	// When the druid moves through Cat, humanoid, Bear, and humanoid forms.
	if got := cat.GetStat(stats.AttackPower) - catAP; got != 77 {
		t.Errorf("Cat Form gained %v attack power, want 77", got)
	}
	cat.ClearForm(sim)
	humanoidAP := cat.GetStat(stats.AttackPower)
	cat.AddStatDynamic(sim, stats.FeralAttackPower, -77)
	if got := cat.GetStat(stats.AttackPower); got != humanoidAP {
		t.Errorf("removing Feral Attack Power in humanoid form changed attack power from %v to %v", humanoidAP, got)
	}
	cat.AddStatDynamic(sim, stats.FeralAttackPower, 77)
	fillResource(sim, cat)
	if !cat.BearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Bear Form did not cast")
	}
	bearAP := cat.GetStat(stats.AttackPower)
	cat.AddStatDynamic(sim, stats.FeralAttackPower, -77)
	if got := bearAP - cat.GetStat(stats.AttackPower); got != 77 {
		t.Errorf("Bear Form gained %v attack power, want 77", got)
	}
	cat.AddStatDynamic(sim, stats.FeralAttackPower, 77)
	cat.ClearForm(sim)

	// Then the form-only dependency is removed again on exit.
	secondHumanoidAP := cat.GetStat(stats.AttackPower)
	cat.AddStatDynamic(sim, stats.FeralAttackPower, -77)
	if got := cat.GetStat(stats.AttackPower); got != secondHumanoidAP {
		t.Errorf("removing Feral Attack Power after Bear changed humanoid attack power from %v to %v", secondHumanoidAP, got)
	}
}

func TestFurorKeepsPullEnergyOnFirstEntry(t *testing.T) {
	// Given a newly reset cat with Furor 5/5.
	_, cat := newCat(t, furorTalents, nil)

	// When its starting Cat Form activates, then it keeps the energy bar's pull value.
	if got := cat.CurrentEnergy(); got != 100 {
		t.Errorf("energy on first Cat Form entry = %v, want 100", got)
	}
}

func TestFurorRegeneratesStoredEnergyWhileFormIsCancelled(t *testing.T) {
	// Given a Furor cat that cancels form at 40 energy.
	sim, cat := newCat(t, furorTalents, nil)
	cat.SpendEnergy(sim, 60, cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	cat.ClearForm(sim)
	sim.CurrentTime += 3 * time.Second
	fillResource(sim, cat)

	// When it returns after three seconds in humanoid form.
	if !cat.CatForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Cat Form did not cast after cancellation")
	}

	// Then the supported assumption restores 40 stored plus 30 regenerated energy.
	if got := cat.CurrentEnergy(); got != 70 {
		t.Errorf("energy after cancelling Cat Form for three seconds = %v, want 70", got)
	}
}

func TestNaturalReactionOnlyRestoresRageInBearForm(t *testing.T) {
	// Given Natural Reaction 5/5 and a dodged incoming attack while in Cat Form.
	sim, cat := newCat(t, "0000000000000000-00000000000000000500-0000000000000000", nil)
	naturalReaction := cat.GetAura("Natural Reaction")
	if naturalReaction == nil || naturalReaction.OnSpellHitTaken == nil {
		t.Fatal("Natural Reaction listener is missing")
	}
	cat.AutoAttacks.CancelAutoSwing(sim)
	cat.CurrentTarget.AutoAttacks.CancelAutoSwing(sim)
	dodge := &core.SpellResult{Target: &cat.Unit, Outcome: core.OutcomeDodge}
	naturalReaction.OnSpellHitTaken(naturalReaction, sim, cat.AutoAttacks.MHAuto(), dodge)
	for end := sim.CurrentTime + core.SpellBatchWindow; sim.CurrentTime <= end && !sim.Step(); {
	}
	if got := cat.CurrentRage(); got != 0 {
		t.Errorf("Natural Reaction restored %v rage in Cat Form, want 0", got)
	}

	// When the same dodge occurs in Bear Form, then the triggered row restores 5 rage.
	fillResource(sim, cat)
	if !cat.BearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Bear Form did not cast")
	}
	cat.SpendRage(sim, cat.CurrentRage(), cat.NewRageMetrics(core.ActionID{SpellID: 1}))
	naturalReaction.OnSpellHitTaken(naturalReaction, sim, cat.AutoAttacks.MHAuto(), dodge)
	for end := sim.CurrentTime + core.SpellBatchWindow; sim.CurrentTime <= end && !sim.Step(); {
	}
	if got := cat.CurrentRage(); got != 5 {
		t.Errorf("Natural Reaction restored %v rage in Bear Form, want 5", got)
	}
}

func TestOmenUsesTwoPPMWithTenSecondICDOnMeleeHits(t *testing.T) {
	// Given Omen's permanent listener and a landed main-hand auto result.
	sim, cat := newCat(t, "0000000000000000-00000000000000000000-0000000000000001", nil)
	omen := cat.GetAura("Omen of Clarity")
	if omen == nil || omen.Dpm == nil || omen.Icd == nil {
		t.Fatalf("Omen listener is incomplete: aura=%v DPM=%v ICD=%v", omen, omen != nil && omen.Dpm != nil, omen != nil && omen.Icd != nil)
	}
	if omen.Icd.Duration != 10*time.Second {
		t.Fatalf("Omen ICD = %v, want 10s", omen.Icd.Duration)
	}
	cat.AutoAttacks.CancelAutoSwing(sim)
	result := &core.SpellResult{Target: cat.CurrentTarget, Outcome: core.OutcomeHit, Damage: 1}
	before := omen.Icd.ReadyAt()
	wrongMask := cat.RegisterSpell(core.SpellConfig{ActionID: core.ActionID{SpellID: 2}, ProcMask: core.ProcMaskRangedAuto})

	// When a ranged hit reaches the listener directly, then its decoded mask rejects it without using the ICD.
	for range 200 {
		omen.OnSpellHitDealt(omen, sim, wrongMask.Spell, result)
	}
	if got := omen.Icd.ReadyAt(); got != before {
		t.Fatalf("a ranged hit moved Omen's ICD from %v to %v", before, got)
	}

	// When enough eligible one-second paw hits roll the 2 PPM manager, one activates Clearcasting.
	for range 500 {
		omen.OnSpellHitDealt(omen, sim, cat.AutoAttacks.MHAuto(), result)
	}
	for sim.CurrentTime <= core.SpellBatchWindow && !sim.Step() {
	}
	if !cat.ClearcastingAura.IsActive() {
		t.Fatal("500 eligible paw hits produced no Clearcasting proc")
	}
	if got := omen.Icd.ReadyAt(); got != 10*time.Second {
		t.Errorf("Omen ICD ready at %v, want 10s", got)
	}

	// Then an eligible hit inside the ICD cannot reactivate a consumed Clearcasting aura.
	cat.ClearcastingAura.Deactivate(sim)
	for range 500 {
		omen.OnSpellHitDealt(omen, sim, cat.AutoAttacks.MHAuto(), result)
	}
	for end := sim.CurrentTime + core.SpellBatchWindow; sim.CurrentTime <= end && !sim.Step(); {
	}
	if cat.ClearcastingAura.IsActive() {
		t.Error("Omen proc bypassed its 10-second ICD")
	}
}

func TestClearcastingIsSpentByACoveredCatAbility(t *testing.T) {
	// Given Clearcasting at full energy.
	sim, cat := newCat(t, "", nil)
	prepareFeralDamageCase(sim, cat)
	cat.AutoAttacks.CancelAutoSwing(sim)
	cat.ClearcastingAura.Activate(sim)
	cat.ClearcastingAura.SetStacks(sim, cat.ClearcastingAura.MaxStacks)

	// When Rake casts, then its covered class flags make the cast free and consume the charge.
	if !cat.Rake.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Rake did not cast")
	}
	for sim.CurrentTime <= core.SpellBatchWindow && !sim.Step() {
	}
	if got := cat.CurrentEnergy(); got != 100 {
		t.Errorf("Clearcasting Rake left %v energy, want 100", got)
	}
	if cat.ClearcastingAura.IsActive() {
		t.Error("Rake did not consume Clearcasting")
	}

	// Given another charge, when the covered ability misses, then the charge is retained.
	cat.ClearcastingAura.Activate(sim)
	cat.ClearcastingAura.SetStacks(sim, cat.ClearcastingAura.MaxStacks)
	miss := &core.SpellResult{Target: cat.CurrentTarget, Outcome: core.OutcomeMiss}
	cat.ClearcastingAura.OnSpellHitDealt(cat.ClearcastingAura, sim, cat.Rake.Spell, miss)
	if !cat.ClearcastingAura.IsActive() || cat.ClearcastingAura.GetStacks() != 1 {
		t.Error("a missed Rake consumed Clearcasting")
	}

	// When an auto attack lands, then the ability-only class mask also keeps the charge.
	landed := &core.SpellResult{Target: cat.CurrentTarget, Outcome: core.OutcomeHit, Damage: 1}
	cat.ClearcastingAura.OnSpellHitDealt(cat.ClearcastingAura, sim, cat.AutoAttacks.MHAuto(), landed)
	if !cat.ClearcastingAura.IsActive() {
		t.Error("a paw auto attack consumed Clearcasting")
	}
}
