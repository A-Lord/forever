package shared

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

const (
	// Eye of Diminution 23001's on-use: 28862, A_MOD_THREAT -35% misc 127 on the wearer for 20 s.
	eyeOfDiminution int32 = 28862

	// Fetish of Chitinous Spikes 21488's on-use: 26168, A_DAMAGE_SHIELD 25 Nature on the wearer for 30 s.
	chitinousSpikes int32 = 26168

	// Essence of the Pure Flame 18815's equip spell: 23266, A_DAMAGE_SHIELD 13 Fire on the wearer.
	fieryAura int32 = 23266

	// Cold Basilisk Eye 5079's on-use: 1139, A_MOD_DECREASE_SPEED -40% and A_MOD_MELEE_RANGED_HASTE_2 -20%
	// on the enemy for 15 s.
	coldEye int32 = 1139

	// Weakness Analyzer 272438's on-use: 1291101, A_MOD_CRIT_PCT +5 on the wearer for 20 s, ProcCharges 1
	// spent by a non-periodic crit.
	analyzingWeaknesses int32 = 1291101

	// Adaptive Combat Assistant 272437's on-use: 1291097, A_SCHOOL_ABSORB 450 of every school for 15 s,
	// which deals 1291099's 90-110 Nature to the enemies in melee range when it breaks early.
	damageSigmoid  int32 = 1291097
	sigmoidRevenge int32 = 1291099
)

// A threat reduction is worth nothing the sim scores, so the item is never used on its own: its major
// cooldown files under no type and never asks to be used. Used by hand, it takes 35% off the wearer's
// threat for 20 s.
func TestOnUseThreatReductionIsNeverUsedOnItsOwn(t *testing.T) {
	const itemID int32 = 992701
	withAuraItem(itemID, proto.ItemType_ItemTypeTrinket, testOnUse(eyeOfDiminution, 120000, 0, 0))
	NewSpellDataAuraOnUse(itemID)

	sim := newAuraSim(shaman("Shaman", map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: itemID}))
	tester := auraTesterAt(sim, 0)
	eye := tester.GetSpell(core.ActionID{ItemID: itemID})
	mcd := tester.GetMajorCooldown(core.ActionID{ItemID: itemID})
	if eye == nil || mcd == nil {
		t.Fatalf("the eye registered no on-use major cooldown")
	}
	if mcd.Type != core.CooldownTypeUnknown || mcd.ShouldActivate(sim, &tester.Character) {
		t.Errorf("the eye's major cooldown is of type %v and asks to be used: %v; want neither", mcd.Type, mcd.ShouldActivate(sim, &tester.Character))
	}

	if !eye.Cast(sim, sim.Encounter.ActiveTargetUnits[0]) {
		t.Fatalf("the eye could not be used")
	}
	start := sim.CurrentTime
	if got := tester.PseudoStats.ThreatMultiplier; !core.WithinToleranceFloat64(0.65, got, 1e-6) {
		t.Errorf("threat multiplier while the eye is up %v, want 0.65", got)
	}
	stepPast(t, sim, start+20*time.Second+time.Millisecond)
	if got := tester.PseudoStats.ThreatMultiplier; !core.WithinToleranceFloat64(1, got, 1e-6) {
		t.Errorf("threat multiplier after its 20 s %v, want 1", got)
	}
}

// What the wearer's damage shield, the spell under its row's id with tag 2, has dealt to the attacker,
// once the handlers the hits queued have run.
func shieldDamage(t *testing.T, sim *core.Simulation, wearer *auraTester, row int32, attacker *auraTester) float64 {
	t.Helper()
	stepPast(t, sim, sim.CurrentTime+core.SpellBatchWindow+time.Millisecond)
	shield := wearer.GetSpell(core.ActionID{SpellID: row, Tag: 2})
	if shield == nil {
		t.Fatalf("the wearer has no damage shield spell for %d", row)
	}
	return shield.SpellMetrics[attacker.UnitIndex].TotalDamage
}

// A wearer certain to land its spells, and an attacker beside it.
func newShieldSim(items map[proto.ItemSlot]int32) (*core.Simulation, *auraTester, *auraTester) {
	sim := newAuraSim(shaman("Wearer", items), shaman("Attacker", nil))
	wearer, attacker := auraTesterAt(sim, 0), auraTesterAt(sim, 1)
	wearer.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100})
	return sim, wearer, attacker
}

// Worn, Essence of the Pure Flame deals 13 Fire back to each melee attacker, and nothing back for a
// spell.
func TestEquipDamageShieldDealsItsAmountBackToAMeleeAttacker(t *testing.T) {
	const itemID int32 = 992702
	withAuraItem(itemID, proto.ItemType_ItemTypeTrinket)
	NewSpellDataEquipAura(SpellDataProc{TriggerSpellID: fieryAura},
		[]ItemVariant{{ItemID: itemID, ItemName: "Test Essence of the Pure Flame"}})

	sim, wearer, attacker := newShieldSim(map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: itemID})

	dealt(sim, attacker.bolt, &wearer.Unit)
	if got := shieldDamage(t, sim, wearer, fieryAura, attacker); got != 0 {
		t.Errorf("an arcane bolt on the wearer drew %v back, want nothing", got)
	}
	dealt(sim, attacker.strike, &wearer.Unit)
	dealt(sim, attacker.strike, &wearer.Unit)
	if got := shieldDamage(t, sim, wearer, fieryAura, attacker); !core.WithinToleranceFloat64(26, got, 1e-6) {
		t.Errorf("two physical strikes on the wearer drew %v back, want 13 each", got)
	}
}

// Used, Fetish of Chitinous Spikes deals 25 Nature back to each melee attacker for its 30 s.
func TestOnUseDamageShieldLastsTheRowsDuration(t *testing.T) {
	const itemID int32 = 992703
	withAuraItem(itemID, proto.ItemType_ItemTypeTrinket, testOnUse(chitinousSpikes, 180000, 0, 0))
	NewSpellDataAuraOnUse(itemID)

	sim, wearer, attacker := newShieldSim(map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: itemID})
	fetish := wearer.GetSpell(core.ActionID{ItemID: itemID})
	if fetish == nil {
		t.Fatalf("the fetish registered no on-use")
	}

	dealt(sim, attacker.strike, &wearer.Unit)
	if got := shieldDamage(t, sim, wearer, chitinousSpikes, attacker); got != 0 {
		t.Errorf("a strike before the fetish was used drew %v back, want nothing", got)
	}

	fetish.Cast(sim, sim.Encounter.ActiveTargetUnits[0])
	start := sim.CurrentTime
	dealt(sim, attacker.strike, &wearer.Unit)
	if got := shieldDamage(t, sim, wearer, chitinousSpikes, attacker); !core.WithinToleranceFloat64(25, got, 1e-6) {
		t.Errorf("a strike while the spikes are up drew %v back, want 25", got)
	}

	stepPast(t, sim, start+30*time.Second+time.Millisecond)
	dealt(sim, attacker.strike, &wearer.Unit)
	if got := shieldDamage(t, sim, wearer, chitinousSpikes, attacker); !core.WithinToleranceFloat64(25, got, 1e-6) {
		t.Errorf("a strike after the 30 s left %v drawn back in all, want the 25 from before", got)
	}
}

// Cold Basilisk Eye slows the enemy's attacks by 20% for 15 s; its movement slow changes nothing the
// sim runs.
func TestColdBasiliskEyeSlowsTheEnemysAttacks(t *testing.T) {
	const itemID int32 = 992704
	withAuraItem(itemID, proto.ItemType_ItemTypeTrinket, testOnUse(coldEye, 420000, 0, 0))
	NewSpellDataAuraOnUse(itemID)

	sim := newAuraSim(shaman("Shaman", map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: itemID}))
	tester := auraTesterAt(sim, 0)
	target := sim.Encounter.ActiveTargetUnits[0]
	eye := tester.GetSpell(core.ActionID{ItemID: itemID})
	if eye == nil {
		t.Fatalf("the eye registered no on-use")
	}

	if !eye.Cast(sim, target) {
		t.Fatalf("the eye could not be used")
	}
	start := sim.CurrentTime
	if got := target.PseudoStats.AttackSpeedMultiplier; !core.WithinToleranceFloat64(1/1.2, got, 1e-6) {
		t.Errorf("the enemy's attack speed multiplier while slowed %v, want 1/1.2", got)
	}
	stepPast(t, sim, start+15*time.Second+time.Millisecond)
	if got := target.PseudoStats.AttackSpeedMultiplier; !core.WithinToleranceFloat64(1, got, 1e-6) {
		t.Errorf("the enemy's attack speed multiplier after the 15 s %v, want 1", got)
	}
}

// Weakness Analyzer adds 5% crit until the wearer's first crit, which spends its one charge: a hit
// that does not crit leaves it up.
func TestWeaknessAnalyzerDropsOnTheFirstCrit(t *testing.T) {
	const itemID int32 = 992705
	withAuraItem(itemID, proto.ItemType_ItemTypeTrinket, testOnUse(analyzingWeaknesses, 90000, 1141, 20000))
	NewSpellDataAuraOnUse(itemID)

	sim := newAuraSim(shaman("Shaman", map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: itemID}))
	tester := auraTesterAt(sim, 0)
	target := sim.Encounter.ActiveTargetUnits[0]
	analyzer := tester.GetSpell(core.ActionID{ItemID: itemID})
	if analyzer == nil {
		t.Fatalf("the analyzer registered no on-use")
	}
	crit := tester.GetStat(stats.PhysicalCritPercent)

	analyzer.Cast(sim, target)
	if got := tester.GetStat(stats.PhysicalCritPercent); !core.WithinToleranceFloat64(crit+5, got, 1e-6) {
		t.Fatalf("crit while the analyzer is up %v, want %v", got, crit+5)
	}

	tester.OnSpellHitDealt(sim, tester.strike, &core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: 100})
	stepPast(t, sim, sim.CurrentTime+core.SpellBatchWindow+time.Millisecond)
	if got := tester.GetStat(stats.PhysicalCritPercent); !core.WithinToleranceFloat64(crit+5, got, 1e-6) {
		t.Errorf("crit after a hit that did not crit %v, want %v", got, crit+5)
	}

	tester.OnSpellHitDealt(sim, tester.bolt, &core.SpellResult{Target: target, Outcome: core.OutcomeCrit, Damage: 150})
	stepPast(t, sim, sim.CurrentTime+core.SpellBatchWindow+time.Millisecond)
	if got := tester.GetStat(stats.PhysicalCritPercent); !core.WithinToleranceFloat64(crit, got, 1e-6) {
		t.Errorf("crit after the first crit %v, want %v", got, crit)
	}
}

// Adaptive Combat Assistant's shield deals 90-110 Nature to the enemy when a hit breaks it before its
// 15 s, and nothing when it runs out unbroken.
func TestAdaptiveCombatAssistantDealsDamageWhenItsShieldBreaks(t *testing.T) {
	const itemID int32 = 992706
	sim, caster := newOnUseSim(t, NewSpellDataAbsorbOnUse,
		map[int32]*proto.ItemEffect{itemID: testOnUse(damageSigmoid, 90000, 1141, 15000)})
	assistant := onUseSpell(t, caster, itemID)
	revenge := caster.GetSpell(core.ActionID{SpellID: sigmoidRevenge})
	if revenge == nil {
		t.Fatalf("the assistant registered no %d", sigmoidRevenge)
	}
	target := sim.Encounter.ActiveTargetUnits[0]
	revengeDealt := func() float64 { return revenge.SpellMetrics[target.UnitIndex].TotalDamage }

	assistant.Cast(sim, &caster.Unit)
	start := sim.CurrentTime
	takenAfterAbsorbs(sim, caster, core.SpellSchoolPhysical, 100)
	stepPast(t, sim, start+15*time.Second+time.Millisecond)
	if got := revengeDealt(); got != 0 {
		t.Errorf("a shield that ran out unbroken dealt %v, want nothing", got)
	}

	assistant.CD.Reset()
	assistant.SharedCD.Reset()
	assistant.Cast(sim, &caster.Unit)
	if got := takenAfterAbsorbs(sim, caster, core.SpellSchoolFire, 1000); got != 550 {
		t.Errorf("a fire hit of 1000 landed for %v, want 550 past the 450 shield", got)
	}
	stepPast(t, sim, sim.CurrentTime+time.Second)
	if got := revengeDealt(); got < 90 || got > 110 {
		t.Errorf("breaking the shield dealt %v, want 90-110", got)
	}
}
