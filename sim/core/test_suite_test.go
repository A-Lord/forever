package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func TestStatWeightsKeepPercentWeightsPerPercent(t *testing.T) {
	pseudoStats := make([]float64, stats.PseudoStatsLen)
	pseudoStats[proto.PseudoStat_PseudoStatBlockPercent] = 5
	pseudoStats[proto.PseudoStat_PseudoStatMeleeHitPercent] = 2
	pseudoStats[proto.PseudoStat_PseudoStatRangedHitPercent] = 3

	got := weightsFromUnitStatsProto(&proto.UnitStats{Stats: make([]float64, stats.ProtoStatsLen), PseudoStats: pseudoStats})
	if got[stats.BlockPercent] != 5 {
		t.Errorf("Block%% weight %v, want the 5 per percent the EP values state", got[stats.BlockPercent])
	}
	if got[stats.PhysicalHitPercent] != 2 || got[stats.RangedHitPercent] != 3 {
		t.Errorf("hit weights %v melee and %v ranged, want 2 and 3", got[stats.PhysicalHitPercent], got[stats.RangedHitPercent])
	}
}
