package core

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/stats"
)

func TestResistCoefficientAgainstHigherLevelTarget(t *testing.T) {
	attacker := &Unit{Type: PlayerUnit, Level: 60}
	spell := &Spell{SpellSchool: SpellSchoolShadow, SchoolIndex: stats.SchoolIndexShadow}
	target := func(shadowResistance float64) *Unit {
		unit := &Unit{Type: EnemyUnit, Level: 63}
		unit.stats[stats.ShadowResistance] = shadowResistance
		return unit
	}

	for _, tc := range []struct {
		name       string
		resistance float64
		binary     bool
		pureDot    bool
		want       float64
	}{
		{"no resistance", 0, false, false, 0.08},
		{"no resistance, pure DoT", 0, false, true, 0.08},
		{"no resistance, binary", 0, true, false, 0},
		{"100 resistance", 100, false, false, 100.0/300 + 0.08},
		{"100 resistance, pure DoT", 100, false, true, 100.0/300/10 + 0.08},
		{"100 resistance, binary", 100, true, false, 100.0 / 300},
	} {
		if got := target(tc.resistance).resistCoeff(spell, attacker, tc.binary, tc.pureDot); math.Abs(got-tc.want) > 1e-12 {
			t.Errorf("%s at +3: coefficient %v, want %v", tc.name, got, tc.want)
		}
	}

	level72 := &Unit{Type: EnemyUnit, Level: 72}
	if got := level72.resistCoeff(spell, attacker, false, false); math.Abs(got-0.08) > 1e-12 {
		t.Errorf("no resistance at +12: coefficient %v, want the +3 value 0.08", got)
	}
}
