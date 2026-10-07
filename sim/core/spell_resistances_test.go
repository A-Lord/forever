package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/stats"
)

// Piercing for one school lowers the resistance a spell of that school meets, and leaves the other
// schools where SpellPiercing alone puts them.
func TestSchoolSpellPiercingLowersOnlyItsSchool(t *testing.T) {
	defender := Unit{Type: EnemyUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	defender.stats[stats.FireResistance] = 100
	defender.stats[stats.FrostResistance] = 100

	fire := &Spell{SpellSchool: SpellSchoolFire, SchoolIndex: stats.SchoolIndexFire}
	frost := &Spell{SpellSchool: SpellSchoolFrost, SchoolIndex: stats.SchoolIndexFrost}

	coeffs := func(generic, firePiercing float64) (float64, float64) {
		attacker := Unit{Type: PlayerUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
		attacker.stats[stats.SpellPiercing] = generic
		attacker.PseudoStats.SchoolSpellPiercing[stats.SchoolIndexFire] = firePiercing
		return defender.averageResist(fire, &attacker, false, false), defender.averageResist(frost, &attacker, false, false)
	}

	_, frostBare := coeffs(0, 0)
	fireResist, frostResist := coeffs(0, 30)
	if want := 70.0 / (70 + 380); !WithinToleranceFloat64(want, fireResist, 1e-9) {
		t.Errorf("fire average resist with 30 fire piercing = %v, want %v", fireResist, want)
	}
	if frostResist != frostBare {
		t.Errorf("frost average resist with 30 fire piercing = %v, want %v as without it", frostResist, frostBare)
	}

	fireBoth, frostBoth := coeffs(20, 30)
	if want := 50.0 / (50 + 380); !WithinToleranceFloat64(want, fireBoth, 1e-9) {
		t.Errorf("fire average resist with 20 piercing and 30 fire piercing = %v, want %v", fireBoth, want)
	}
	if want := 80.0 / (80 + 380); !WithinToleranceFloat64(want, frostBoth, 1e-9) {
		t.Errorf("frost average resist with 20 piercing and 30 fire piercing = %v, want %v", frostBoth, want)
	}
}

// Against 150 resistance a plain spell is often partially resisted, and one flagged
// SpellFlagNoPartialResist never is.
func TestNoPartialResistKeepsTheWholeHit(t *testing.T) {
	attacker := &Unit{Type: PlayerUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	defender := &Unit{Type: EnemyUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	defender.stats[stats.ShadowResistance] = 150
	attackTable := NewAttackTable(attacker, defender)

	partials := func(flags SpellFlag) int {
		sim := &Simulation{rand: NewSplitMix(1)}
		spell := &Spell{Unit: attacker, SpellSchool: SpellSchoolShadow, SchoolIndex: stats.SchoolIndexShadow, Flags: flags}
		n := 0
		for range 1000 {
			if multiplier, outcome := spell.ResistanceMultiplier(sim, false, attackTable); multiplier != 1 || outcome != OutcomeEmpty {
				n++
			}
		}
		return n
	}

	if n := partials(0); n == 0 {
		t.Errorf("a plain shadow spell was never partially resisted in 1000 hits against 150 shadow resistance")
	}
	if n := partials(SpellFlagNoPartialResist); n != 0 {
		t.Errorf("a spell flagged SpellFlagNoPartialResist was partially resisted %d times in 1000 hits, want 0", n)
	}
}
