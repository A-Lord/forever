package core

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/stats"
)

func TestDefenseRaisesChanceToBeMissed(t *testing.T) {
	table := &AttackTable{BaseMissChance: 0.056}
	missedWith := func(defenseRating float64) float64 {
		unit := &Unit{PseudoStats: stats.NewPseudoStats()}
		unit.stats[stats.DefenseRating] = defenseRating
		return unit.GetTotalChanceToBeMissedAsDefender(table)
	}

	if got := missedWith(0); math.Abs(got-0.056) > 1e-12 {
		t.Errorf("no defense: missed %v, want the table's 0.056", got)
	}
	if got := missedWith(25.9 * DefenseRatingPerDefenseLevel); math.Abs(got-(0.056+25*0.0004)) > 1e-12 {
		t.Errorf("25.9 defense: missed %v, want 0.04%% for each of the 25 whole points on top", got)
	}
}

func TestBlockValueIsNeverNegative(t *testing.T) {
	unit := &Unit{PseudoStats: stats.NewPseudoStats()}
	unit.stats[stats.BlockValue] = -1
	if got := unit.BlockDamageReduction(); got != 0 {
		t.Errorf("block value %v below zero, want 0", got)
	}

	unit.stats[stats.BlockValue] = 120*0.05 - 1
	if got := unit.BlockDamageReduction(); math.Abs(got-5) > 1e-12 {
		t.Errorf("block value %v from 120 strength, want 5", got)
	}
}
