package core

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/stats"
)

func TestCritChanceSubtractsCritTakenReduction(t *testing.T) {
	tests := []struct {
		name             string
		rawChance        float64
		reducedCritTaken float64
		want             float64
	}{
		{name: "defense immunity", rawChance: 0.05, reducedCritTaken: 0.05, want: 0},
		{name: "partial reduction", rawChance: 0.05, reducedCritTaken: 0.03, want: 0.02},
		{name: "reduction beyond the raw chance", rawChance: 0.05, reducedCritTaken: 0.06, want: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			target := &Unit{PseudoStats: stats.NewPseudoStats()}
			target.PseudoStats.ReducedCritTakenPercent = test.reducedCritTaken

			if got := getCritChance(test.rawChance, target); math.Abs(got-test.want) > 1e-9 {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestCritTakenReductionLowersPhysicalAndSpellCritByItsPercent(t *testing.T) {
	attacker := &Unit{Type: EnemyUnit}
	attacker.stats[stats.PhysicalCritPercent] = 20
	attacker.stats[stats.SpellCritPercent] = 20
	defender := &Unit{Type: PlayerUnit, PseudoStats: stats.NewPseudoStats()}
	defender.PseudoStats.ReducedCritTakenPercent = 0.05
	attacker.AttackTables = []*AttackTable{NewAttackTable(attacker, defender)}
	spell := &Spell{Unit: attacker, ProcMask: ProcMaskMeleeMHSpecial}

	if got := spell.PhysicalCritChance(attacker.AttackTables[0]); math.Abs(got-0.15) > 1e-9 {
		t.Errorf("physical crit chance %v, want 0.15", got)
	}
	if got := spell.SpellCritChance(defender); math.Abs(got-0.15) > 1e-9 {
		t.Errorf("spell crit chance %v, want 0.15", got)
	}
}
