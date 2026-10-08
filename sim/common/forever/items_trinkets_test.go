//go:build with_db

package forever_test

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

func trinketWarrior(t *testing.T, trinketID int32) (*core.Simulation, *core.Character) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotTrinket2+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotTrinket1].Id = trinketID
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Wearer", Class: proto.Class_ClassWarrior, Race: proto.Race_RaceOrc,
				Spec: &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{Options: &proto.DpsWarrior_Options{
					ClassOptions: &proto.WarriorOptions{DefaultStance: proto.WarriorStance_WarriorStanceBerserker}}}},
				Equipment:   &proto.EquipmentSpec{Items: items},
				Consumables: &proto.ConsumesSpec{},
				Buffs:       &proto.IndividualBuffs{},
				Rotation:    &proto.APLRotation{},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

func stepTo(t *testing.T, sim *core.Simulation, until time.Duration) {
	t.Helper()
	reached := false
	sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{DoAt: until, OnAction: func(*core.Simulation) { reached = true }}))
	for steps := 0; !reached; steps++ {
		if steps > 10000 {
			t.Fatalf("the sim did not reach %v", until)
		}
		sim.Step()
	}
}

// Uses the item and returns its spell, after checking the major cooldown it is filed under and the
// cooldowns its on-use effect states.
func useTrinket(t *testing.T, sim *core.Simulation, character *core.Character, itemID int32, cdType core.CooldownType,
	cooldown time.Duration, category int32, categoryCooldown time.Duration) *core.Spell {
	t.Helper()
	spell := character.GetSpell(core.ActionID{ItemID: itemID})
	if spell == nil {
		t.Fatalf("%d registers no on-use spell", itemID)
	}
	mcd := character.GetMajorCooldown(spell.ActionID)
	if mcd == nil || mcd.Type != cdType {
		t.Errorf("%d's major cooldown is %v, want type %v", itemID, mcd, cdType)
	}
	if spell.CD.Duration != cooldown {
		t.Errorf("%d's cooldown is %v, want %v", itemID, spell.CD.Duration, cooldown)
	}
	if category == 0 && spell.SharedCD.Timer != nil {
		t.Errorf("%d shares a cooldown, want none", itemID)
	}
	if category != 0 && (spell.SharedCD.Timer != character.GetOrInitSpellCategoryTimer(category) || spell.SharedCD.Duration != categoryCooldown) {
		t.Errorf("%d shares %v on another timer than category %d's, want %v on it", itemID, spell.SharedCD.Duration, category, categoryCooldown)
	}
	if !spell.Cast(sim, character.CurrentTarget) {
		t.Fatalf("%d could not be used", itemID)
	}
	return spell
}

func gained(character *core.Character, before stats.Stats, stat stats.Stat) float64 {
	return character.GetStat(stat) - before[stat]
}

func meleeSwing(unit *core.Unit) *core.Spell {
	return &core.Spell{ProcMask: core.ProcMaskMeleeMHAuto, SpellSchool: core.SpellSchoolPhysical, Unit: unit}
}

func fireSpell(unit *core.Unit) *core.Spell {
	return &core.Spell{ProcMask: core.ProcMaskSpellDamage, SpellSchool: core.SpellSchoolFire, DefenseType: core.DefenseTypeMagic, Unit: unit}
}

// A landed hit taken, and the spell batch window the listener resolves after.
func takeHit(t *testing.T, sim *core.Simulation, character *core.Character, spell *core.Spell) {
	t.Helper()
	character.OnSpellHitTaken(sim, spell, &core.SpellResult{Target: &character.Unit, Outcome: core.OutcomeHit, Damage: 100})
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
}

func dealHit(t *testing.T, sim *core.Simulation, character *core.Character, spell *core.Spell) {
	t.Helper()
	character.OnSpellHitDealt(sim, spell, &core.SpellResult{Target: character.CurrentTarget, Outcome: core.OutcomeHit, Damage: 100})
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
}

// Lifegiving Gem 19341 raises maximum health by 15% for 20 s and heals for the same amount, on its
// 5 min cooldown in no category. When the raise fades, health below the lower maximum stays, health
// above it drops to it, and neither counts as damage taken.
func TestLifegivingGemRaisesMaximumHealthBy15Percent(t *testing.T) {
	sim, character := trinketWarrior(t, 19341)
	maxHealth := character.MaxHealth()
	character.RemoveHealth(sim, maxHealth/2)
	current := character.CurrentHealth()
	damageTaken := character.DamageTakenHealthMetrics.ActualGain

	useTrinket(t, sim, character, 19341, core.CooldownTypeSurvival, 5*time.Minute, 0, 0)
	bonus := maxHealth * 0.15
	if got := character.MaxHealth(); math.Abs(got-(maxHealth+bonus)) > 1e-6 {
		t.Errorf("maximum health is %v, want %v", got, maxHealth+bonus)
	}
	if got := character.CurrentHealth(); math.Abs(got-(current+bonus)) > 1e-6 {
		t.Errorf("health is %v, want %v healed by %v", got, current, bonus)
	}

	stepTo(t, sim, sim.CurrentTime+20*time.Second+time.Millisecond)
	if got := character.MaxHealth(); math.Abs(got-maxHealth) > 1e-6 {
		t.Errorf("after 20 s maximum health is %v, want %v back", got, maxHealth)
	}
	if got := character.CurrentHealth(); math.Abs(got-(current+bonus)) > 1e-6 {
		t.Errorf("after 20 s below the maximum health is %v, want %v kept", got, current+bonus)
	}

	character.GainHealth(sim, maxHealth, character.NewHealthMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionHealingModel}))
	gift := character.GetAura("Gift of Life")
	gift.Activate(sim)
	if got := character.CurrentHealth(); math.Abs(got-(maxHealth+bonus)) > 1e-6 {
		t.Errorf("at full health the gem heals to %v, want %v", got, maxHealth+bonus)
	}
	gift.Deactivate(sim)
	if got := character.CurrentHealth(); math.Abs(got-maxHealth) > 1e-6 {
		t.Errorf("at full health the raise fades to %v health, want %v", got, maxHealth)
	}
	if got := character.DamageTakenHealthMetrics.ActualGain; got != damageTaken {
		t.Errorf("damage taken is %v after the raise fades, want %v", got, damageTaken)
	}
}

// Zandalarian Hero Badge 19948 grants 2000 armor and 30 defense, and each melee or ranged hit taken
// takes 200 armor and 3 defense off it; a spell hit taken does not.
func TestZandalarianHeroBadgeLosesArmorPerHitTaken(t *testing.T) {
	sim, character := trinketWarrior(t, 19948)
	before := character.GetStats()

	useTrinket(t, sim, character, 19948, core.CooldownTypeSurvival, 2*time.Minute, 0, 0)
	if got := gained(character, before, stats.Armor); got != 2000 {
		t.Errorf("the badge grants %v armor, want 2000", got)
	}
	if got := gained(character, before, stats.DefenseRating); got != 30*core.DefenseRatingPerDefenseLevel {
		t.Errorf("the badge grants %v defense rating, want 30 defense", got)
	}

	enemy := sim.Encounter.ActiveTargetUnits[0]
	takeHit(t, sim, character, meleeSwing(enemy))
	if got := gained(character, before, stats.Armor); got != 1800 {
		t.Errorf("after a melee hit taken the badge grants %v armor, want 1800", got)
	}
	if got := gained(character, before, stats.DefenseRating); got != 27*core.DefenseRatingPerDefenseLevel {
		t.Errorf("after a melee hit taken the badge grants %v defense rating, want 27 defense", got)
	}
	takeHit(t, sim, character, fireSpell(enemy))
	if got := gained(character, before, stats.Armor); got != 1800 {
		t.Errorf("after a spell hit taken the badge grants %v armor, want 1800 still", got)
	}
}

// Zandalarian Hero Medallion 19949 adds 40 to melee and ranged damage, 2 less for each hit dealt, on
// category 1141's 20 s shared cooldown.
func TestZandalarianHeroMedallionLosesDamagePerHit(t *testing.T) {
	sim, character := trinketWarrior(t, 19949)
	before := character.GetStats()

	useTrinket(t, sim, character, 19949, core.CooldownTypeDPS, 2*time.Minute, 1141, 20*time.Second)
	if got := gained(character, before, stats.PhysicalDamage); got != 40 {
		t.Errorf("the medallion adds %v damage, want 40", got)
	}
	dealHit(t, sim, character, meleeSwing(&character.Unit))
	if got := gained(character, before, stats.PhysicalDamage); got != 38 {
		t.Errorf("after a hit the medallion adds %v damage, want 38", got)
	}
	for range 19 {
		dealHit(t, sim, character, meleeSwing(&character.Unit))
	}
	if got := gained(character, before, stats.PhysicalDamage); got != 0 {
		t.Errorf("after 20 hits the medallion adds %v damage, want none", got)
	}
	if character.GetAuraByID(core.ActionID{SpellID: 24661}).IsActive() {
		t.Errorf("the buff is still up with no stacks left")
	}
}

// Zandalarian Hero Charm 19950 grants 204 spell damage and 408 healing, 17 and 34 less for each spell
// cast; a melee cast does not count.
func TestZandalarianHeroCharmLosesPowerPerSpellCast(t *testing.T) {
	sim, character := trinketWarrior(t, 19950)
	before := character.GetStats()

	useTrinket(t, sim, character, 19950, core.CooldownTypeDPS, 2*time.Minute, 1141, 20*time.Second)
	if got, want := gained(character, before, stats.SpellDamage), 204.0; got != want {
		t.Errorf("the charm grants %v spell damage, want %v", got, want)
	}
	if got, want := gained(character, before, stats.HealingPower), 408.0; got != want {
		t.Errorf("the charm grants %v healing, want %v", got, want)
	}

	character.OnCastComplete(sim, fireSpell(&character.Unit))
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
	if got := gained(character, before, stats.SpellDamage); got != 187 {
		t.Errorf("after a spell cast the charm grants %v spell damage, want 187", got)
	}
	if got := gained(character, before, stats.HealingPower); got != 374 {
		t.Errorf("after a spell cast the charm grants %v healing, want 374", got)
	}

	character.OnCastComplete(sim, &core.Spell{ProcMask: core.ProcMaskMeleeMHSpecial, SpellSchool: core.SpellSchoolPhysical, Unit: &character.Unit})
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
	if got := gained(character, before, stats.SpellDamage); got != 187 {
		t.Errorf("after a melee cast the charm grants %v spell damage, want 187 still", got)
	}
}

// Petrified Scarab 21685 adds 100 to every resistance for 60 s, 10 less for each hostile spell that
// lands; a melee hit taken does not count.
func TestPetrifiedScarabLosesResistancePerSpellTaken(t *testing.T) {
	sim, character := trinketWarrior(t, 21685)
	before := character.GetStats()
	resistances := []stats.Stat{stats.ArcaneResistance, stats.FireResistance, stats.FrostResistance, stats.NatureResistance, stats.ShadowResistance}
	expect := func(when string, want float64) {
		t.Helper()
		for _, stat := range resistances {
			if got := gained(character, before, stat); got != want {
				t.Errorf("%s the scarab adds %v %s, want %v", when, got, stat.StatName(), want)
			}
		}
	}

	useTrinket(t, sim, character, 21685, core.CooldownTypeSurvival, 3*time.Minute, 0, 0)
	if got := character.GetAuraByID(core.ActionID{SpellID: 26463}).Duration; got != time.Minute {
		t.Errorf("the buff lasts %v, want 60 s", got)
	}
	expect("on use", 100)

	enemy := sim.Encounter.ActiveTargetUnits[0]
	takeHit(t, sim, character, fireSpell(enemy))
	expect("after a spell taken", 90)
	takeHit(t, sim, character, meleeSwing(enemy))
	expect("after a melee hit taken", 90)
}

// Everlook Delivery Bot 274760 marks the target; once 6 to 8 spells or attacks have landed on it,
// 1295633 deals 467 Fire to it and the mark is gone.
func TestEverlookDeliveryBotDeliversAfterItsCount(t *testing.T) {
	sim, character := trinketWarrior(t, 274760)
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100})
	target := character.CurrentTarget

	useTrinket(t, sim, character, 274760, core.CooldownTypeDPS, 5*time.Minute, 1141, 20*time.Second)
	mark := target.GetAuraByID(core.ActionID{SpellID: 1295271})
	if mark == nil || !mark.IsActive() {
		t.Fatalf("using the bot put no mark on the target")
	}
	delivery := character.GetSpell(core.ActionID{SpellID: 1295633})

	hits := 0
	for mark.IsActive() && hits < 20 {
		target.OnSpellHitTaken(sim, meleeSwing(&character.Unit), &core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: 100})
		stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
		hits++
	}
	if hits < 6 || hits > 8 {
		t.Errorf("the delivery came after %d hits, want 6 to 8", hits)
	}

	stepTo(t, sim, sim.CurrentTime+10*time.Second)
	metrics := delivery.SpellMetrics[target.UnitIndex]
	if metrics.Casts != 1 || metrics.TotalDamage != 467 {
		t.Errorf("the delivery cast %d times for %v damage, want once for 467", metrics.Casts, metrics.TotalDamage)
	}
}
