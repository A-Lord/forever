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
		return defender.resistCoeff(fire, &attacker, false), defender.resistCoeff(frost, &attacker, false)
	}

	fireBare, frostBare := coeffs(0, 0)
	fireCoeff, frostCoeff := coeffs(0, 30)
	if want := 70.0 / 300; !WithinToleranceFloat64(want, fireCoeff, 1e-9) {
		t.Errorf("fire coefficient with 30 fire piercing = %v, want %v", fireCoeff, want)
	}
	if fireCoeff >= fireBare {
		t.Errorf("fire coefficient with 30 fire piercing = %v, not below %v without it", fireCoeff, fireBare)
	}
	if frostCoeff != frostBare {
		t.Errorf("frost coefficient with 30 fire piercing = %v, want %v as without it", frostCoeff, frostBare)
	}

	fireBoth, frostBoth := coeffs(20, 30)
	if want := 50.0 / 300; !WithinToleranceFloat64(want, fireBoth, 1e-9) {
		t.Errorf("fire coefficient with 20 piercing and 30 fire piercing = %v, want %v", fireBoth, want)
	}
	if want := 80.0 / 300; !WithinToleranceFloat64(want, frostBoth, 1e-9) {
		t.Errorf("frost coefficient with 20 piercing and 30 fire piercing = %v, want %v", frostBoth, want)
	}
}
