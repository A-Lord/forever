package core

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/stats"
)

func TestResistCoefficientAgainstHigherLevelTarget(t *testing.T) {
	attacker := &Unit{Type: PlayerUnit, Level: 60}
	target := func(shadowResistance float64) *Unit {
		unit := &Unit{Type: EnemyUnit, Level: 63}
		unit.stats[stats.ShadowResistance] = shadowResistance
		return unit
	}

	for _, tc := range []struct {
		name       string
		resistance float64
		binary     bool
		tick       bool
		flags      SpellFlag
		want       float64
	}{
		{"no resistance", 0, false, false, 0, 0.08},
		{"no resistance, tick", 0, false, true, 0, 0.08},
		{"no resistance, binary", 0, true, false, 0, 0},
		{"100 resistance", 100, false, false, 0, 100.0/300 + 0.08},
		{"100 resistance, tick", 100, false, true, 0, 100.0/300/10 + 0.08},
		{"100 resistance, channel tick", 100, false, true, SpellFlagChanneled, 100.0/300 + 0.08},
		{"100 resistance, tick with full resistance", 100, false, true, SpellFlagFullResistanceTicks, 100.0/300 + 0.08},
		{"100 resistance, binary", 100, true, false, 0, 100.0 / 300},
	} {
		spell := &Spell{SpellSchool: SpellSchoolShadow, SchoolIndex: stats.SchoolIndexShadow, Flags: tc.flags}
		if got := target(tc.resistance).resistCoeff(spell, attacker, tc.binary, tc.tick); math.Abs(got-tc.want) > 1e-12 {
			t.Errorf("%s at +3: coefficient %v, want %v", tc.name, got, tc.want)
		}
	}
}
