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
	"github.com/wowsims/forever/sim/core/testagent"
)

// The caster stand-in's spells: a 2 s cast and an instant, both on the global cooldown.
const (
	castTimeSpellID = 403
	instantSpellID  = 8042
)

func init() {
	testagent.Register(proto.Player_ElementalShaman{}, proto.Spec_SpecElementalShaman, func(player *proto.Player, spec any) {
		player.Spec = spec.(*proto.Player_ElementalShaman)
	}, func(agent *testagent.Agent) {
		agent.EnableMelee()
		for _, spell := range []struct {
			id       int32
			castTime time.Duration
		}{{castTimeSpellID, 2 * time.Second}, {instantSpellID, 0}} {
			agent.RegisterSpell(core.SpellConfig{
				ActionID:    core.ActionID{SpellID: spell.id},
				SpellSchool: core.SpellSchoolNature,
				DefenseType: core.DefenseTypeMagic,
				ProcMask:    core.ProcMaskSpellDamage,
				Cast: core.CastConfig{
					DefaultCast: core.Cast{GCD: core.GCDDefault, CastTime: spell.castTime},
				},
				DamageMultiplier: 1,
				ThreatMultiplier: 1,
				ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
					spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeMagicHitAndCrit)
				},
			})
		}
	})
	testagent.Register(proto.Player_Rogue{}, proto.Spec_SpecRogue, func(player *proto.Player, spec any) {
		player.Spec = spec.(*proto.Player_Rogue)
	}, func(agent *testagent.Agent) {
		agent.EnableAutoAttacks(agent, core.AutoAttackOptions{Ranged: agent.WeaponFromRanged(), AutoSwingRanged: true})
	})
}

// A character swinging one weapon in the main hand against level 60 targets with no armor: the
// warrior stand-in, or with a shaman the caster stand-in.
func weaponWearer(t *testing.T, class proto.Class, itemID int32, targets int) (*core.Simulation, *core.Character) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = itemID
	player := &proto.Player{
		Name: "Wearer", Class: class, Race: proto.Race_RaceTauren,
		Equipment:   &proto.EquipmentSpec{Items: items},
		Consumables: &proto.ConsumesSpec{},
		Buffs:       &proto.IndividualBuffs{},
		Rotation:    &proto.APLRotation{},
	}
	if class == proto.Class_ClassShaman {
		core.WithSpec(player, &proto.Player_ElementalShaman{ElementalShaman: &proto.ElementalShaman{}})
	} else {
		core.WithSpec(player, &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{}})
	}
	encounter := &proto.Encounter{Duration: 180}
	for range targets {
		encounter.Targets = append(encounter.Targets, &proto.Target{Name: "target", Level: 60})
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}},
		Encounter:  encounter,
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

// Graverobber's Shovel 273817's equip spell 1319047 grants 12 attack power and 12 ranged attack
// power, which the item's stats carry, and 12 of each against Undead (creature type mask 32).
func TestGraverobbersShovelGrantsAttackPowerAgainstUndead(t *testing.T) {
	const shovel = 273817
	_, againstUndead := warriorWithConsumes(t, &proto.ConsumesSpec{}, shovel, 0, proto.MobType_MobTypeUndead)
	if got, want := bonusAgainstTarget(againstUndead), (stats.Stats{stats.AttackPower: 12, stats.RangedAttackPower: 12}); got != want {
		t.Errorf("the bonus against Undead is %v, want %v", got, want)
	}
	_, againstDemons := warriorWithConsumes(t, &proto.ConsumesSpec{}, shovel, 0, proto.MobType_MobTypeDemon)
	if got := bonusAgainstTarget(againstDemons); got != (stats.Stats{}) {
		t.Errorf("the bonus against Demons is %v, want none", got)
	}
}

// Sulfuras, Hand of Ragnaros 17182 carries both its effects. Immolation 21142 deals 5 Fire to the
// attacker of each melee hit taken, and nothing for a spell. Fireball 21162 procs at the 0.7 procs
// per minute of the item's chance-on-hit entry and deals 303 Fire (273 to 333) and 15 every 2 s for
// 10 s.
func TestSulfurasCarriesItsDamageShieldAndFireball(t *testing.T) {
	sim, character := itemWearer(t, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotMainHand, 17182)
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100})
	enemy := sim.Encounter.ActiveTargetUnits[0]

	shield := character.GetSpell(core.ActionID{SpellID: 21142, Tag: 2})
	if shield == nil {
		t.Fatal("17182 registers no damage shield")
	}
	takeHit(t, sim, character, meleeSwing(enemy))
	if got := shield.SpellMetrics[enemy.UnitIndex]; got.Casts != 1 || got.TotalDamage != 5 {
		t.Errorf("a melee hit taken cast the shield %d times for %v, want once for 5", got.Casts, got.TotalDamage)
	}
	takeHit(t, sim, character, fireSpell(enemy))
	if got := shield.SpellMetrics[enemy.UnitIndex].Casts; got != 1 {
		t.Errorf("after a spell hit taken the shield cast %d times, want once still", got)
	}

	proc := character.GetAura("Sulfuras, Hand of Ragnaros")
	if proc == nil || proc.Dpm == nil {
		t.Fatal("17182 registers no Fireball proc on a procs per minute rate")
	}
	want := 0.7 * core.GetItemByID(17182).SwingSpeed / 60
	if got := proc.Dpm.Chance(core.ProcMaskMeleeMHAuto, sim); math.Abs(got-want) > 1e-9 {
		t.Errorf("the Fireball's chance on a main hand swing is %v, want %v", got, want)
	}

	fireball := character.GetSpell(core.ActionID{SpellID: 21162})
	fireball.Cast(sim, enemy)
	dot := fireball.Dot(enemy)
	if dot.BaseTickCount != 5 || dot.BaseTickLength != 2*time.Second {
		t.Errorf("the Fireball ticks %d times every %v, want 5 every 2s", dot.BaseTickCount, dot.BaseTickLength)
	}
	stepTo(t, sim, sim.CurrentTime+10*time.Second+time.Millisecond)
	metrics := fireball.SpellMetrics[enemy.UnitIndex]
	metrics.TotalDamage -= 75
	if metrics.Casts != 1 || metrics.Ticks != 5 || !damageInRange(metrics, 273, 333) {
		t.Errorf("the Fireball cast %d times and ticked %d times for %v besides 75, want once, 5 ticks and 273 to 333",
			metrics.Casts, metrics.Ticks, metrics.TotalDamage)
	}
}

// Ravager 7717 procs Bladestorm 9632 at the 1 procs per minute of the item's chance-on-hit entry.
// Bladestorm is a 9 s channel on the wearer that cuts off a cast in progress and holds back the auto
// attacks, and every 3 s strikes each enemy with 9633: main-hand weapon damage (104 to 157 and the
// attack power share) plus 5. The strikes do not proc Bladestorm.
func TestRavagerChannelsBladestormOverEveryEnemy(t *testing.T) {
	sim, character := weaponWearer(t, proto.Class_ClassWarrior, 7717, 3)

	proc := character.GetAura("Ravager")
	if proc == nil || proc.Dpm == nil {
		t.Fatal("7717 registers no Bladestorm proc on a procs per minute rate")
	}
	want := 1 * core.GetItemByID(7717).SwingSpeed / 60
	if got := proc.Dpm.Chance(core.ProcMaskMeleeMHAuto, sim); math.Abs(got-want) > 1e-9 {
		t.Errorf("Bladestorm's chance on a main hand swing is %v, want %v", got, want)
	}
	// Every melee hit that may proc Bladestorm now does.
	*proc.Dpm = *character.NewFixedProcChanceManager(1, core.ProcMaskMelee)

	targets := sim.Encounter.ActiveTargetUnits
	for _, target := range targets {
		target.PseudoStats.InFrontOfTarget = false
	}
	bladestorm := character.GetSpell(core.ActionID{SpellID: 9632})
	strike := character.GetSpell(core.ActionID{SpellID: 9633})
	if bladestorm == nil || strike == nil {
		t.Fatal("7717 registers no Bladestorm or no strike")
	}
	extra := 5 + strike.MeleeAttackPower(targets[0])*3.5/14
	casts := func(spell *core.Spell) (total int32) {
		for _, target := range targets {
			total += spell.SpellMetrics[target.UnitIndex].Casts
		}
		return total
	}
	character.AutoAttacks.EnableAutoSwing(sim)
	swings := casts(character.AutoAttacks.MHAuto())

	start := sim.CurrentTime
	bladestorm.Cast(sim, targets[0])
	dot := bladestorm.AOEDot()
	if dot.Unit != &character.Unit || dot.BaseTickCount != 3 || dot.BaseTickLength != 3*time.Second {
		t.Errorf("Bladestorm ticks %d times every %v on %s, want 3 every 3s on the wearer",
			dot.BaseTickCount, dot.BaseTickLength, dot.Unit.Label)
	}
	stepTo(t, sim, start+6*time.Second+time.Millisecond)
	if !character.IsChanneling() {
		t.Error("the wearer is not channeling Bladestorm after 6 s")
	}
	if got := casts(bladestorm); got != 1 {
		t.Errorf("two rounds of strikes cast Bladestorm %d times in all, want once", got)
	}
	if got := casts(character.AutoAttacks.MHAuto()) - swings; got != 0 {
		t.Errorf("the wearer swung %d times during Bladestorm, want none", got)
	}
	// The first swing after the channel would cast it again.
	proc.Deactivate(sim)

	stepTo(t, sim, start+9*time.Second+time.Millisecond)
	if dot.IsActive() || character.IsChanneling() {
		t.Error("Bladestorm is still up after 9 s")
	}

	for _, target := range targets {
		metrics := strike.SpellMetrics[target.UnitIndex]
		if outcomes := metrics.Hits + metrics.Crits + metrics.Misses + metrics.Dodges; outcomes != 3 {
			t.Errorf("%s took %d strikes, want 3", target.Label, outcomes)
		}
		hitDamage := metrics.TotalDamage - metrics.TotalCritDamage
		if low, high := float64(metrics.Hits)*(104+extra), float64(metrics.Hits)*(157+extra); hitDamage < low || hitDamage > high {
			t.Errorf("%s took %v from %d strikes that hit, want %v to %v", target.Label, hitDamage, metrics.Hits, low, high)
		}
	}

	sim, character = weaponWearer(t, proto.Class_ClassShaman, 7717, 1)
	character.AutoAttacks.CancelAutoSwing(sim)
	enemy := sim.Encounter.ActiveTargetUnits[0]
	castTime := character.GetSpell(core.ActionID{SpellID: castTimeSpellID})
	castTime.Cast(sim, enemy)
	character.GetSpell(core.ActionID{SpellID: 9632}).Cast(sim, enemy)
	stepTo(t, sim, sim.CurrentTime+2*time.Second+time.Millisecond)
	if got := castTime.SpellMetrics[enemy.UnitIndex].Casts; got != 0 {
		t.Errorf("the 2 s cast Bladestorm started on completed %d times, want none", got)
	}
}

// Heartseeking Crossbow 13040's 29626 procs at the 3 procs per minute of the item's entry, on the
// crossbow's 3.1 s speed, and strikes the target with Shadowbolt 29640 for 16 Shadow (13 to 19).
func TestHeartseekingCrossbowStrikesWithAShadowbolt(t *testing.T) {
	sim, character := itemWearer(t, proto.Class_ClassRogue, proto.ItemSlot_ItemSlotRanged, 13040)
	enemy := sim.Encounter.ActiveTargetUnits[0]
	proc := character.GetAura("Heartseeking Crossbow")
	shadowbolt := character.GetSpell(core.ActionID{SpellID: 29640})
	if proc == nil || proc.Dpm == nil || shadowbolt == nil {
		t.Fatal("13040 registers no Shadowbolt or no proc on a procs per minute rate")
	}
	if got, want := proc.Dpm.Chance(core.ProcMaskRangedAuto, sim), 3*3.1/60; math.Abs(got-want) > 1e-9 {
		t.Errorf("the Shadowbolt's chance on a ranged shot is %v, want %v", got, want)
	}
	// Every ranged hit now procs it.
	*proc.Dpm = *character.NewFixedProcChanceManager(1, core.ProcMaskRanged)

	shot := &core.Spell{ProcMask: core.ProcMaskRangedAuto, SpellSchool: core.SpellSchoolPhysical, Unit: &character.Unit}
	dealHit(t, sim, character, shot)
	metrics := shadowbolt.SpellMetrics[enemy.UnitIndex]
	hitDamage := metrics.TotalDamage - metrics.TotalCritDamage
	if metrics.Casts != 1 || hitDamage < 13*float64(metrics.Hits) || hitDamage > 19*float64(metrics.Hits) {
		t.Errorf("a ranged hit cast the Shadowbolt %d times for %v from %d hits, want once for 13 to 19 a hit",
			metrics.Casts, hitDamage, metrics.Hits)
	}
}

// Hand of Edward the Odd 2243 procs Focus 18803 at the 1 procs per minute of the item's chance-on-hit
// entry. Focus makes the next spell with a cast time instant, on its whole global cooldown, and that
// cast spends it; an instant leaves it up, and unspent it fades after 4 s.
func TestHandOfEdwardTheOddMakesTheNextCastInstant(t *testing.T) {
	sim, character := weaponWearer(t, proto.Class_ClassShaman, 2243, 1)
	character.AutoAttacks.CancelAutoSwing(sim)
	enemy := sim.Encounter.ActiveTargetUnits[0]

	proc := character.GetAura("Hand of Edward the Odd")
	if proc == nil || proc.Dpm == nil {
		t.Fatal("2243 registers no Focus proc on a procs per minute rate")
	}
	want := 1 * core.GetItemByID(2243).SwingSpeed / 60
	if got := proc.Dpm.Chance(core.ProcMaskMeleeMHAuto, sim); math.Abs(got-want) > 1e-9 {
		t.Errorf("Focus's chance on a main hand swing is %v, want %v", got, want)
	}

	focus := character.GetAura("Hand of Edward the Odd Focus")
	if focus == nil || focus.Duration != 4*time.Second {
		t.Fatalf("2243 registers Focus as %v, want a 4s aura", focus)
	}
	castTime := character.GetSpell(core.ActionID{SpellID: castTimeSpellID})
	instant := character.GetSpell(core.ActionID{SpellID: instantSpellID})

	focus.Activate(sim)
	instant.Cast(sim, enemy)
	if !focus.IsActive() {
		t.Error("an instant spent Focus, want it kept")
	}

	stepTo(t, sim, character.GCD.ReadyAt())
	start := sim.CurrentTime
	castTime.Cast(sim, enemy)
	if got := castTime.SpellMetrics[enemy.UnitIndex].Casts; got != 1 {
		t.Errorf("the 2 s cast under Focus completed %d times at once, want once", got)
	}
	if focus.IsActive() {
		t.Error("the 2 s cast left Focus up, want it spent")
	}
	if got := character.GCD.ReadyAt() - start; got != core.GCDDefault {
		t.Errorf("the instant cast spent a %v global cooldown, want %v", got, core.GCDDefault)
	}

	stepTo(t, sim, character.GCD.ReadyAt())
	castTime.Cast(sim, enemy)
	if got := character.Hardcast.Expires - sim.CurrentTime; got != 2*time.Second {
		t.Errorf("the next cast after Focus is spent takes %v, want 2s", got)
	}

	stepTo(t, sim, character.Hardcast.Expires+time.Millisecond)
	focus.Activate(sim)
	stepTo(t, sim, sim.CurrentTime+4*time.Second+time.Millisecond)
	if focus.IsActive() {
		t.Error("Focus is still up after 4 s unspent")
	}
}
