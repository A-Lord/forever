package core

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func statWeightMetrics(dps float64) *proto.UnitMetrics {
	distribution := func(v float64) *proto.DistributionMetrics {
		return &proto.DistributionMetrics{Avg: v, AllValues: []float64{v, v}}
	}
	return &proto.UnitMetrics{Dps: distribution(dps), Hps: distribution(0), Threat: distribution(0), Dtps: distribution(0), Tmi: distribution(0)}
}

func statWeightResult(dps float64) *proto.RaidSimResult {
	return &proto.RaidSimResult{RaidMetrics: &proto.RaidMetrics{Parties: []*proto.PartyMetrics{{Players: []*proto.UnitMetrics{statWeightMetrics(dps)}}}}}
}

// Builds the stat weight requests and answers each sub-sim with the DPS a point of its stat is worth
// in dpsPerPoint, then computes the weights.
func runSyntheticStatWeights(t *testing.T, statsToWeigh []proto.Stat, pseudoStatsToWeigh []proto.PseudoStat, dpsPerPoint map[stats.UnitStat]float64) (*proto.StatWeightRequestsData, *proto.StatWeightsResult) {
	t.Helper()
	const baseline = 1000.0
	requests := buildStatWeightRequests(&proto.StatWeightsRequest{
		Player:             &proto.Player{},
		Encounter:          &proto.Encounter{},
		SimOptions:         &proto.SimOptions{Iterations: 10, RandomSeed: 1},
		StatsToWeigh:       statsToWeigh,
		PseudoStatsToWeigh: pseudoStatsToWeigh,
		EpReferenceStat:    proto.Stat_StatAgility,
	})

	var results []*proto.StatWeightsStatResultData
	for _, request := range requests.StatSimRequests {
		stat := stats.UnitStatFromIdx(int(request.StatData.UnitStat))
		result := &proto.StatWeightsStatResultData{
			StatData:   request.StatData,
			ResultHigh: statWeightResult(baseline + dpsPerPoint[stat]*request.StatData.ModHigh),
		}
		if request.RequestLow != nil {
			result.ResultLow = statWeightResult(baseline + dpsPerPoint[stat]*request.StatData.ModLow)
		}
		results = append(results, result)
	}

	return requests, computeStatWeights(&proto.StatWeightsCalcRequest{
		BaseResult:      statWeightResult(baseline),
		EpReferenceStat: proto.Stat_StatAgility,
		StatSimResults:  results,
	})
}

func protoUnitStat(unitStats *proto.UnitStats, stat stats.UnitStat) float64 {
	if stat.IsStat() {
		return unitStats.Stats[stat.StatIdx()]
	}
	return unitStats.PseudoStats[stat.PseudoStatIdx()]
}

func simulatesStat(requests *proto.StatWeightRequestsData, stat stats.UnitStat) bool {
	for _, request := range requests.StatSimRequests {
		if request.StatData.UnitStat == int32(stat) {
			return true
		}
	}
	return false
}

var (
	agilityUnitStat   = stats.UnitStatFromStat(stats.Agility)
	hitRatingUnitStat = stats.UnitStatFromStat(stats.HitRating)
	meleeHitUnitStat  = stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatMeleeHitPercent)
	rangedHitUnitStat = stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatRangedHitPercent)
)

// A spec weighing melee and ranged hit percent still simulates hit rating, and its weight is that sim's
// answer rather than the sum of the percents, which would count the ranged attacks twice.
func TestStatWeightsSimulateHitRatingBesideItsPercents(t *testing.T) {
	requests, result := runSyntheticStatWeights(t,
		[]proto.Stat{proto.Stat_StatAgility, proto.Stat_StatHitRating},
		[]proto.PseudoStat{proto.PseudoStat_PseudoStatMeleeHitPercent, proto.PseudoStat_PseudoStatRangedHitPercent},
		map[stats.UnitStat]float64{agilityUnitStat: 1, hitRatingUnitStat: 3, meleeHitUnitStat: 25, rangedHitUnitStat: 10})

	if !simulatesStat(requests, hitRatingUnitStat) {
		t.Fatal("no sub-sim for hit rating")
	}
	for _, c := range []struct {
		name string
		got  float64
		want float64
	}{
		{"hit rating weight", protoUnitStat(result.Dps.Weights, hitRatingUnitStat), 3},
		{"hit rating EP", protoUnitStat(result.Dps.EpValues, hitRatingUnitStat), 3},
		{"melee hit percent weight", protoUnitStat(result.Dps.Weights, meleeHitUnitStat), 25},
		{"ranged hit percent weight", protoUnitStat(result.Dps.Weights, rangedHitUnitStat), 10},
	} {
		if math.Abs(c.got-c.want) > 1e-9 {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

// A hunter weighs ranged hit percent alone and still gets a hit rating weight.
func TestStatWeightsSimulateHitRatingBesideRangedHitPercent(t *testing.T) {
	requests, result := runSyntheticStatWeights(t,
		[]proto.Stat{proto.Stat_StatAgility, proto.Stat_StatHitRating},
		[]proto.PseudoStat{proto.PseudoStat_PseudoStatRangedHitPercent},
		map[stats.UnitStat]float64{agilityUnitStat: 1, hitRatingUnitStat: 2, rangedHitUnitStat: 18})

	if !simulatesStat(requests, hitRatingUnitStat) {
		t.Fatal("no sub-sim for hit rating")
	}
	if got := protoUnitStat(result.Dps.Weights, hitRatingUnitStat); math.Abs(got-2) > 1e-9 {
		t.Errorf("hit rating weight = %v, want 2", got)
	}
}

// A percent the spec weighs without its rating leaves the rating unweighed.
func TestStatWeightsLeaveAnUnrequestedRatingAtZero(t *testing.T) {
	requests, result := runSyntheticStatWeights(t,
		[]proto.Stat{proto.Stat_StatAgility},
		[]proto.PseudoStat{proto.PseudoStat_PseudoStatMeleeHitPercent},
		map[stats.UnitStat]float64{agilityUnitStat: 1, meleeHitUnitStat: 25})

	if simulatesStat(requests, hitRatingUnitStat) {
		t.Error("a sub-sim for hit rating, which the request does not weigh")
	}
	if got := protoUnitStat(result.Dps.Weights, hitRatingUnitStat); got != 0 {
		t.Errorf("hit rating weight = %v, want 0", got)
	}
}
