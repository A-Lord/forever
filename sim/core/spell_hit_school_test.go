package core

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/stats"
)

// A spell of several schools takes the school hit of each: Frostfire Bolt's Fire and Frost add up. A
// spell with no class mask takes none.
func TestSpellHitChanceAddsEachSchoolsHit(t *testing.T) {
	unit := &Unit{}
	unit.PseudoStats.SchoolBonusHitChance[stats.SchoolIndexFire] = 1
	unit.PseudoStats.SchoolBonusHitChance[stats.SchoolIndexFrost] = 2

	for _, tc := range []struct {
		name      string
		school    SpellSchool
		classMask int64
		want      float64
	}{
		{"fire", SpellSchoolFire, 1, 0.01},
		{"frostfire", SpellSchoolFire | SpellSchoolFrost, 1, 0.03},
		{"frostfire with no class mask", SpellSchoolFire | SpellSchoolFrost, 0, 0},
	} {
		spell := &Spell{Unit: unit, SpellSchool: tc.school, ClassSpellMask: tc.classMask}
		if got := spell.SpellHitChance(nil); math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("%s: %v, want %v", tc.name, got, tc.want)
		}
	}
}
