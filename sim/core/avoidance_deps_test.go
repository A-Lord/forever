package core

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/stats"
)

func TestBlockRatingAndDefenseReachTheAvoidancePercents(t *testing.T) {
	unit := &Unit{StatDependencyManager: stats.NewStatDependencyManager()}
	unit.addUniversalStatDependencies()
	unit.FinalizeStatDeps()
	unit.stats = unit.ApplyStatDependencies(stats.Stats{
		stats.BlockRating:   2 * BlockRatingPerBlockPercent,
		stats.DodgeRating:   DodgeRatingPerDodgePercent,
		stats.ParryRating:   ParryRatingPerParryPercent,
		stats.DefenseRating: 25 * DefenseRatingPerDefenseLevel,
	})

	for _, want := range []struct {
		name  string
		got   float64
		value float64
	}{
		{"block", unit.GetBlockFromRating(), 0.02 + 0.01},
		{"dodge", unit.GetDodgeFromRating(), 0.01 + 0.01},
		{"parry", unit.GetParryFromRating(), 0.01 + 0.01},
		{"crit taken", unit.GetStat(stats.ReducedCritTakenPercent) / 100, 0.01},
	} {
		if math.Abs(want.got-want.value) > 1e-12 {
			t.Errorf("%s chance from rating and 25 defense: got %v, want %v", want.name, want.got, want.value)
		}
	}
}

func TestCritTakenCountsWholeDefenseAndResilience(t *testing.T) {
	unit := &Unit{StatDependencyManager: stats.NewStatDependencyManager()}
	unit.addUniversalStatDependencies()
	unit.FinalizeStatDeps()
	unit.stats = unit.ApplyStatDependencies(stats.Stats{
		stats.DefenseRating:    25.9 * DefenseRatingPerDefenseLevel,
		stats.ResilienceRating: 2 * ResilienceRatingPerCritReductionChance,
	})

	if got := unit.GetStat(stats.ReducedCritTakenPercent); math.Abs(got-3) > 1e-12 {
		t.Errorf("25.9 defense and the resilience for 2%% reduce crit taken by %v%%, want 3", got)
	}
}
