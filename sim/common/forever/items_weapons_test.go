//go:build with_db

package forever_test

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

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
