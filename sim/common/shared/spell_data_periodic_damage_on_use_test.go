package shared

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// Thunderbrew's Boot Flask 744's on-use, 12257, keeps an aura on the wearer for 5 s that casts 12278,
// 75 Fire to every enemy in a cone, every 1 s, on a 5 min cooldown. Cloak of Fire 14134's 18364 casts
// 18374, 25 Fire to every enemy around the wearer, every 5 s for 15 s, on a 2 min cooldown. Nothing
// lands when the item is used, each tick reaches all three targets, and nothing lands once the aura
// is gone.
func TestPeriodicDamageOnUseTicksOnEveryTarget(t *testing.T) {
	for _, tc := range []struct {
		name     string
		itemID   int32
		spellID  int32
		damageID int32
		cooldown time.Duration
		period   time.Duration
		ticks    int
		amount   float64
	}{
		{"Thunderbrew's Boot Flask", 991701, 12257, 12278, 5 * time.Minute, time.Second, 5, 75},
		{"Cloak of Fire", 991702, 18364, 18374, 2 * time.Minute, 5 * time.Second, 3, 25},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim, caster := newOnUseSimAgainst(t, NewSpellDataPeriodicDamageOnUse,
				map[int32]*proto.ItemEffect{tc.itemID: testOnUse(tc.spellID, int32(tc.cooldown.Milliseconds()), 0, 0)}, &proto.ItemSpec{}, 3)
			onUse := onUseSpell(t, caster, tc.itemID)
			targets := sim.Encounter.ActiveTargetUnits

			if got := caster.GetInitialMajorCooldown(onUse.ActionID); !got.Type.Matches(core.CooldownTypeDPS) {
				t.Errorf("the on-use is a major cooldown of type %v, want a DPS one", got.Type)
			}

			if !onUse.Cast(sim, targets[0]) {
				t.Fatalf("%s could not be used", tc.name)
			}
			start := sim.CurrentTime
			damage := caster.GetSpell(core.ActionID{SpellID: tc.damageID})
			if damage == nil {
				t.Fatalf("the on-use registered no spell for %d", tc.damageID)
			}
			aura := onUse.SelfHot()
			if aura == nil || !aura.IsActive() {
				t.Fatalf("using %s put no aura on the wearer", tc.name)
			}

			dealt := func(i int) float64 {
				metrics := damage.SpellMetrics[targets[i].UnitIndex]
				if metrics.Misses+metrics.ResistedHits > 0 {
					t.Fatalf("%d missed target %d", tc.damageID, i)
				}
				return metrics.TotalDamage
			}
			for i := range targets {
				if got := dealt(i); got != 0 {
					t.Errorf("target %d took %v when the item was used, want nothing until the first tick", i, got)
				}
			}

			for tick := 1; tick <= tc.ticks; tick++ {
				stepPast(t, sim, start+time.Duration(tick)*tc.period+time.Millisecond)
				for i := range targets {
					if got, want := dealt(i), float64(tick)*tc.amount; got != want {
						t.Errorf("after tick %d target %d took %v in all, want %v", tick, i, got, want)
					}
				}
			}
			if aura.IsActive() {
				t.Errorf("the aura is still up after its %d ticks", tc.ticks)
			}

			stepPast(t, sim, start+time.Duration(tc.ticks+1)*tc.period+time.Millisecond)
			for i := range targets {
				if got, want := dealt(i), float64(tc.ticks)*tc.amount; got != want {
					t.Errorf("target %d took %v once the aura was gone, want the %v of its ticks", i, got, want)
				}
			}

			if got := onUse.CD.ReadyAt(); got != start+tc.cooldown {
				t.Errorf("%s is ready again at %v, want %v after its use at %v", tc.name, got, tc.cooldown, start)
			}
		})
	}
}
