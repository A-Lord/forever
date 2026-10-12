package core

import (
	"math"
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	googleProto "google.golang.org/protobuf/proto"
)

// A distribution over values, with the aggregate data a sim result carries.
func statWeightDistribution(values []float64) *proto.DistributionMetrics {
	var aggr aggregator
	for _, value := range values {
		aggr.add(value)
	}
	mean, stdev := aggr.meanAndStdDev()
	return &proto.DistributionMetrics{Avg: mean, Stdev: stdev, AllValues: values, AggregatorData: &proto.AggregatorData{N: int32(aggr.n), SumSq: aggr.sumSq}}
}

func statWeightMetrics(dps float64) *proto.UnitMetrics {
	distribution := func(v float64) *proto.DistributionMetrics {
		return statWeightDistribution([]float64{v, v})
	}
	return &proto.UnitMetrics{Dps: distribution(dps), Hps: distribution(0), Threat: distribution(0), Dtps: distribution(0), Tmi: distribution(0)}
}

func statWeightResult(dps float64) *proto.RaidSimResult {
	return &proto.RaidSimResult{RaidMetrics: &proto.RaidMetrics{Parties: []*proto.PartyMetrics{{Players: []*proto.UnitMetrics{statWeightMetrics(dps)}}}}}
}

const syntheticBaselineDps = 1000.0

// Builds the stat weight requests and answers each sub-sim with the DPS a point of its stat is worth
// in dpsPerPoint, then computes the weights.
func runSyntheticStatWeights(t *testing.T, statsToWeigh []proto.Stat, pseudoStatsToWeigh []proto.PseudoStat, dpsPerPoint map[stats.UnitStat]float64) (*proto.StatWeightRequestsData, *proto.StatWeightsResult) {
	t.Helper()
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
			ResultHigh: statWeightResult(syntheticBaselineDps + dpsPerPoint[stat]*request.StatData.ModHigh),
		}
		if request.RequestLow != nil {
			result.ResultLow = statWeightResult(syntheticBaselineDps + dpsPerPoint[stat]*request.StatData.ModLow)
		}
		results = append(results, result)
	}

	return requests, computeStatWeights(&proto.StatWeightsCalcRequest{
		BaseResult:      statWeightResult(syntheticBaselineDps),
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

// A spec weighing melee and ranged hit percent, or a hunter weighing ranged hit percent alone, still
// simulates hit rating, and its weight is that sim's answer. Each percent sub-sim measures its own
// channel, so the rating is simulated directly to value all of its channels at once.
func TestStatWeightsSimulateHitRatingBesideItsPercents(t *testing.T) {
	for _, c := range []struct {
		name        string
		percents    []proto.PseudoStat
		dpsPerPoint map[stats.UnitStat]float64
		want        map[string]float64
	}{
		{"melee and ranged hit", []proto.PseudoStat{proto.PseudoStat_PseudoStatMeleeHitPercent, proto.PseudoStat_PseudoStatRangedHitPercent},
			map[stats.UnitStat]float64{agilityUnitStat: 1, hitRatingUnitStat: 3, meleeHitUnitStat: 25, rangedHitUnitStat: 10},
			map[string]float64{"hit rating weight": 3, "hit rating EP": 3, "melee hit percent weight": 25, "ranged hit percent weight": 10}},
		{"ranged hit alone", []proto.PseudoStat{proto.PseudoStat_PseudoStatRangedHitPercent},
			map[stats.UnitStat]float64{agilityUnitStat: 1, hitRatingUnitStat: 2, rangedHitUnitStat: 18},
			map[string]float64{"hit rating weight": 2}},
	} {
		t.Run(c.name, func(t *testing.T) {
			requests, result := runSyntheticStatWeights(t, []proto.Stat{proto.Stat_StatAgility, proto.Stat_StatHitRating}, c.percents, c.dpsPerPoint)

			if !simulatesStat(requests, hitRatingUnitStat) {
				t.Fatal("no sub-sim for hit rating")
			}
			got := map[string]float64{
				"hit rating weight":         protoUnitStat(result.Dps.Weights, hitRatingUnitStat),
				"hit rating EP":             protoUnitStat(result.Dps.EpValues, hitRatingUnitStat),
				"melee hit percent weight":  protoUnitStat(result.Dps.Weights, meleeHitUnitStat),
				"ranged hit percent weight": protoUnitStat(result.Dps.Weights, rangedHitUnitStat),
			}
			for name, want := range c.want {
				if math.Abs(got[name]-want) > 1e-9 {
					t.Errorf("%s = %v, want %v", name, got[name], want)
				}
			}
		})
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

// A stat simulated below and above its baseline pools both sides' samples, hit rating has only the
// side above, and the baseline DPS comes back without its per-iteration values.
func TestStatWeightsReportSampleCountsAndBaselineDps(t *testing.T) {
	_, result := runSyntheticStatWeights(t,
		[]proto.Stat{proto.Stat_StatAgility, proto.Stat_StatHitRating},
		nil,
		map[stats.UnitStat]float64{agilityUnitStat: 1, hitRatingUnitStat: 3})

	perSide := float64(len(statWeightMetrics(syntheticBaselineDps).Dps.AllValues))
	if got := protoUnitStat(result.SampleCounts, agilityUnitStat); got != 2*perSide {
		t.Errorf("agility samples = %v, want %v", got, 2*perSide)
	}
	if got := protoUnitStat(result.SampleCounts, hitRatingUnitStat); got != perSide {
		t.Errorf("hit rating samples = %v, want %v", got, perSide)
	}
	if got := result.BaseDps.Avg; got != syntheticBaselineDps {
		t.Errorf("baseline DPS avg = %v, want %v", got, syntheticBaselineDps)
	}
	if n := len(result.BaseDps.AllValues); n != 0 {
		t.Errorf("baseline DPS carries %d per-iteration values, want none", n)
	}
}

func statWeightRun(dps []float64) *proto.RaidSimResult {
	return statWeightRunWithHps(dps, make([]float64, len(dps)))
}

func statWeightRunWithHps(dps, hps []float64) *proto.RaidSimResult {
	zeros := make([]float64, len(dps))
	return &proto.RaidSimResult{RaidMetrics: &proto.RaidMetrics{Parties: []*proto.PartyMetrics{{Players: []*proto.UnitMetrics{{
		Dps:    statWeightDistribution(dps),
		Hps:    statWeightDistribution(hps),
		Threat: statWeightDistribution(zeros),
		Dtps:   statWeightDistribution(zeros),
		Tmi:    statWeightDistribution(zeros),
	}}}}}}
}

var noisyDpsPerPoint = map[stats.UnitStat]float64{agilityUnitStat: 1, hitRatingUnitStat: 3}

// Agility, the reference, simulated below and above the baseline, and hit rating above it only.
func noisyStatWeightRequests() *proto.StatWeightRequestsData {
	return buildStatWeightRequests(&proto.StatWeightsRequest{
		Player:          &proto.Player{},
		Encounter:       &proto.Encounter{},
		SimOptions:      &proto.SimOptions{Iterations: 2, RandomSeed: 1},
		StatsToWeigh:    []proto.Stat{proto.Stat_StatAgility, proto.Stat_StatHitRating},
		EpReferenceStat: proto.Stat_StatAgility,
	})
}

// The baseline and every stat's runs over iterations [from, to). Each iteration has a DPS shared by
// all runs, and each run adds its stat's worth and noise of its own, so the paired differences vary.
func noisyStatWeightResults(requests *proto.StatWeightRequestsData, from, to int) (*proto.RaidSimResult, []*proto.StatWeightsStatResultData) {
	values := func(run int, shift float64) []float64 {
		var dps []float64
		for i := from; i < to; i++ {
			dps = append(dps, syntheticBaselineDps+float64((i*7)%13)+shift+float64(((i+1)*(run+3)*5)%7))
		}
		return dps
	}

	var results []*proto.StatWeightsStatResultData
	for j, request := range requests.StatSimRequests {
		worth := noisyDpsPerPoint[stats.UnitStatFromIdx(int(request.StatData.UnitStat))]
		result := &proto.StatWeightsStatResultData{
			StatData:   request.StatData,
			ResultHigh: statWeightRun(values(2*j+2, worth*request.StatData.ModHigh)),
		}
		if request.RequestLow != nil {
			result.ResultLow = statWeightRun(values(2*j+1, worth*request.StatData.ModLow))
		}
		results = append(results, result)
	}
	return statWeightRun(values(0, 0)), results
}

func noisyStatWeightCalc(from, to int) *proto.StatWeightsCalcRequest {
	requests := noisyStatWeightRequests()
	base, results := noisyStatWeightResults(requests, from, to)
	return &proto.StatWeightsCalcRequest{BaseResult: base, EpReferenceStat: requests.EpReferenceStat, StatSimResults: results}
}

func shownEpCi(result *proto.StatWeightsResult, stat stats.UnitStat) float64 {
	return statWeightZ90 * protoUnitStat(result.Dps.EpValuesStdev, stat) / math.Sqrt(protoUnitStat(result.SampleCounts, stat))
}

func plannedIterations(plan *proto.StatWeightsTopUpPlan) map[stats.UnitStat]int32 {
	planned := make(map[stats.UnitStat]int32)
	for _, stat := range plan.Stats {
		planned[stats.UnitStatFromIdx(int(stat.UnitStat))] = stat.Iterations
	}
	return planned
}

// The plan sizes each stat for the target from its EP spread and pooled sides, caps it, runs the
// baseline as far as the furthest stat, and leaves out every stat already at or under the target.
func TestStatWeightTopUpPlanSizesEachStatForTheTarget(t *testing.T) {
	const perSide = 20
	calc := noisyStatWeightCalc(0, perSide)
	result := computeStatWeights(calc)
	agilityCi, hitCi := shownEpCi(result, agilityUnitStat), shownEpCi(result, hitRatingUnitStat)
	if agilityCi == 0 || hitCi == 0 || agilityCi == hitCi {
		t.Fatalf("shown ± agility %v, hit rating %v: want two different spreads", agilityCi, hitCi)
	}
	sized := func(stat stats.UnitStat, target, sides float64) int32 {
		return int32(math.Ceil(math.Pow(statWeightZ90*protoUnitStat(result.Dps.EpValuesStdev, stat)/target, 2) / sides))
	}
	const uncapped = math.MaxInt32 / perSide
	plan := func(target float64, multiplier int32) *proto.StatWeightsTopUpPlan {
		return planStatWeightTopUp(&proto.StatWeightsTopUpRequest{Calc: calc, Target: &proto.StatWeightsTarget{EpCi: target, MaxIterationMultiplier: multiplier}})
	}

	t.Run("sized by the formula", func(t *testing.T) {
		target := min(agilityCi, hitCi) / 4
		got := plan(target, uncapped)
		want := map[stats.UnitStat]int32{agilityUnitStat: sized(agilityUnitStat, target, 2), hitRatingUnitStat: sized(hitRatingUnitStat, target, 1)}
		for stat, iterations := range want {
			if plannedIterations(got)[stat] != iterations {
				t.Errorf("stat %d planned %d iterations per side, want %d", stat, plannedIterations(got)[stat], iterations)
			}
		}
		if got.BaselineIterations != max(want[agilityUnitStat], want[hitRatingUnitStat]) {
			t.Errorf("baseline planned %d iterations, want the furthest stat's", got.BaselineIterations)
		}
	})

	t.Run("capped", func(t *testing.T) {
		const multiplier = 2
		got := plan(min(agilityCi, hitCi)/4, multiplier)
		for _, stat := range []stats.UnitStat{agilityUnitStat, hitRatingUnitStat} {
			if plannedIterations(got)[stat] != multiplier*perSide {
				t.Errorf("stat %d planned %d iterations per side, want the cap", stat, plannedIterations(got)[stat])
			}
		}
		if got.BaselineIterations != multiplier*perSide {
			t.Errorf("baseline planned %d iterations, want the cap", got.BaselineIterations)
		}
	})

	t.Run("capped at the default multiplier when unset", func(t *testing.T) {
		got := plan(min(agilityCi, hitCi)/100, 0)
		if got.BaselineIterations != defaultStatWeightMaxIterationMultiplier*perSide {
			t.Errorf("baseline planned %d iterations, want the default cap", got.BaselineIterations)
		}
	})

	t.Run("a stat at its target is left out", func(t *testing.T) {
		got := plannedIterations(plan((agilityCi+hitCi)/2, uncapped))
		wider, narrower := agilityUnitStat, hitRatingUnitStat
		if hitCi > agilityCi {
			wider, narrower = narrower, wider
		}
		if _, ok := got[narrower]; ok {
			t.Errorf("stat %d is under the target and still planned", narrower)
		}
		if _, ok := got[wider]; !ok {
			t.Errorf("stat %d is over the target and not planned", wider)
		}
	})

	t.Run("done once every stat is at its target", func(t *testing.T) {
		got := plan(max(agilityCi, hitCi), uncapped)
		if len(got.Stats) != 0 || got.BaselineIterations != perSide {
			t.Errorf("plan = %v, want no stats and the baseline's own count", got)
		}
	})

	t.Run("done once the passes have run", func(t *testing.T) {
		target := &proto.StatWeightsTarget{EpCi: min(agilityCi, hitCi) / 4, MaxIterationMultiplier: uncapped}
		withSegments := func(passes int) *proto.StatWeightsTopUpPlan {
			segmented := googleProto.Clone(calc).(*proto.StatWeightsCalcRequest)
			for range passes {
				segmented.ExtraSegments = append(segmented.ExtraSegments, &proto.StatWeightsCalcSegment{})
			}
			return planStatWeightTopUp(&proto.StatWeightsTopUpRequest{Calc: segmented, Target: target})
		}
		if got := withSegments(statWeightMaxTopUpPasses - 1); len(got.Stats) == 0 {
			t.Errorf("plan after %d passes = %v, want stats", statWeightMaxTopUpPasses-1, got)
		}
		if got := withSegments(statWeightMaxTopUpPasses); len(got.Stats) != 0 {
			t.Errorf("plan after %d passes = %v, want no stats", statWeightMaxTopUpPasses, got)
		}
	})
}

// Runs merge with their iterations in run order, the mean and spread of all of them, and the first
// run's chance of death.
func TestStatWeightRunsConcatenateInOrder(t *testing.T) {
	runs := []*proto.RaidSimResult{
		statWeightRun([]float64{1000, 1012, 990}),
		statWeightRun([]float64{1031, 975}),
		statWeightRun([]float64{1004, 1019, 998, 1007}),
	}
	for i, run := range runs {
		setChanceOfDeath(run, 0.1*float64(i+1))
	}

	merged := concatStatWeightRuns(runs).RaidMetrics.Parties[0].Players[0]
	var want []float64
	for _, run := range runs {
		want = append(want, run.RaidMetrics.Parties[0].Players[0].Dps.AllValues...)
	}
	if !slices.Equal(merged.Dps.AllValues, want) {
		t.Errorf("DPS per iteration = %v, want %v", merged.Dps.AllValues, want)
	}
	wantDps := statWeightDistribution(want)
	if math.Abs(merged.Dps.Avg-wantDps.Avg) > 1e-9 || math.Abs(merged.Dps.Stdev-wantDps.Stdev) > 1e-9 {
		t.Errorf("DPS avg %v stdev %v, want %v and %v", merged.Dps.Avg, merged.Dps.Stdev, wantDps.Avg, wantDps.Stdev)
	}
	if merged.Dps.AggregatorData.N != int32(len(want)) {
		t.Errorf("DPS iterations = %d, want %d", merged.Dps.AggregatorData.N, len(want))
	}
	if merged.ChanceOfDeath != runs[0].RaidMetrics.Parties[0].Players[0].ChanceOfDeath {
		t.Errorf("chance of death = %v, want the first run's", merged.ChanceOfDeath)
	}
}

// A stat that stopped early pairs with the baseline's first iterations, so it weighs as if every run
// had stopped where it did; extra segments merge into the same weights one longer run gives.
func TestStatWeightsPairAStatThatStoppedEarlyWithTheBaselinePrefix(t *testing.T) {
	const perSide = 12
	requests := noisyStatWeightRequests()
	firstBase, firstResults := noisyStatWeightResults(requests, 0, perSide)
	secondBase, secondResults := noisyStatWeightResults(requests, perSide, 2*perSide)
	longBase, longResults := noisyStatWeightResults(requests, 0, 2*perSide)
	isHit := func(result *proto.StatWeightsStatResultData) bool {
		return result.StatData.UnitStat == int32(hitRatingUnitStat)
	}

	// Hit rating keeps only its first iterations; agility and the baseline run on.
	longResults[slices.IndexFunc(longResults, isHit)] = firstResults[slices.IndexFunc(firstResults, isHit)]
	early := computeStatWeights(&proto.StatWeightsCalcRequest{BaseResult: longBase, EpReferenceStat: requests.EpReferenceStat, StatSimResults: longResults})
	trimmed := computeStatWeights(&proto.StatWeightsCalcRequest{BaseResult: firstBase, EpReferenceStat: requests.EpReferenceStat, StatSimResults: firstResults})
	segmented := computeStatWeights(&proto.StatWeightsCalcRequest{
		BaseResult:      firstBase,
		EpReferenceStat: requests.EpReferenceStat,
		StatSimResults:  firstResults,
		ExtraSegments: []*proto.StatWeightsCalcSegment{{
			BaseResult:     secondBase,
			StatSimResults: slices.DeleteFunc(secondResults, isHit),
		}},
	})

	same := func(name string, got, want *proto.StatWeightsResult, stat stats.UnitStat) {
		t.Helper()
		for field, pair := range map[string][2]*proto.UnitStats{
			"weight":  {got.Dps.Weights, want.Dps.Weights},
			"stdev":   {got.Dps.WeightsStdev, want.Dps.WeightsStdev},
			"samples": {got.SampleCounts, want.SampleCounts},
		} {
			if g, w := protoUnitStat(pair[0], stat), protoUnitStat(pair[1], stat); math.Abs(g-w) > 1e-9 {
				t.Errorf("%s stat %d %s = %v, want %v", name, stat, field, g, w)
			}
		}
	}
	if protoUnitStat(early.SampleCounts, hitRatingUnitStat) != perSide {
		t.Fatalf("hit rating samples = %v, want its own iterations", protoUnitStat(early.SampleCounts, hitRatingUnitStat))
	}
	same("stopped early", early, trimmed, hitRatingUnitStat)
	for _, stat := range []stats.UnitStat{agilityUnitStat, hitRatingUnitStat} {
		same("segmented", segmented, early, stat)
	}
}

var attackPowerUnitStat = stats.UnitStatFromStat(stats.AttackPower)

// The weapon tester weighing attack power, its reference, in a test sim. It auto-shoots, so it holds
// a bow: an empty ranged slot deals NaN damage.
func weaponTesterStatWeightsRequest(iterations int32) *proto.StatWeightsRequest {
	return &proto.StatWeightsRequest{
		Player:          weaponTestRaid(proto.Race_RaceHuman, WeaponTestSlots(WeaponTestSwordID, 0, WeaponTestBowID), nil).Parties[0].Players[0],
		Encounter:       MakeSingleTargetEncounter(0),
		SimOptions:      &proto.SimOptions{Iterations: iterations, RandomSeed: 1, IsTest: true},
		StatsToWeigh:    []proto.Stat{proto.Stat_StatAttackPower},
		EpReferenceStat: proto.Stat_StatAttackPower,
	}
}

// A run continued at the seed after its last iteration, through the concurrent split and with labeled
// rands, repeats the iterations one longer run has at those seeds, so its segment gives the same weights.
func TestStatWeightTopUpRepeatsOneLongerRun(t *testing.T) {
	const probe, topUp = 7, 5
	requests := buildStatWeightRequests(weaponTesterStatWeightsRequest(2 * (probe + topUp)))
	run := func(simRequest *proto.RaidSimRequest, done, iterations int32) *proto.RaidSimResult {
		continued := googleProto.Clone(simRequest).(*proto.RaidSimRequest)
		continued.SimOptions.Iterations = iterations
		continued.SimOptions.RandomSeed += int64(done)
		return runSimConcurrent(continued, nil, simsignals.CreateSignals())
	}
	statRequest := requests.StatSimRequests[0]

	long := &proto.StatWeightsCalcRequest{
		BaseResult:      run(requests.BaseRequest, 0, probe+topUp),
		EpReferenceStat: requests.EpReferenceStat,
		StatSimResults: []*proto.StatWeightsStatResultData{{
			StatData:   statRequest.StatData,
			ResultLow:  run(statRequest.RequestLow, 0, probe+topUp),
			ResultHigh: run(statRequest.RequestHigh, 0, probe+topUp),
		}},
	}
	continued := &proto.StatWeightsCalcRequest{
		BaseResult:      run(requests.BaseRequest, 0, probe),
		EpReferenceStat: requests.EpReferenceStat,
		StatSimResults: []*proto.StatWeightsStatResultData{{
			StatData:   statRequest.StatData,
			ResultLow:  run(statRequest.RequestLow, 0, probe),
			ResultHigh: run(statRequest.RequestHigh, 0, probe),
		}},
		ExtraSegments: []*proto.StatWeightsCalcSegment{{
			BaseResult: run(requests.BaseRequest, probe, topUp),
			StatSimResults: []*proto.StatWeightsStatResultData{{
				StatData:   statRequest.StatData,
				ResultLow:  run(statRequest.RequestLow, probe, topUp),
				ResultHigh: run(statRequest.RequestHigh, probe, topUp),
			}},
		}},
	}

	merged := mergeStatWeightSegments(continued)
	dpsValues := func(result *proto.RaidSimResult) []float64 {
		return result.RaidMetrics.Parties[0].Players[0].Dps.AllValues
	}
	for name, pair := range map[string][2]*proto.RaidSimResult{
		"baseline":           {merged.BaseResult, long.BaseResult},
		"attack power below": {merged.StatSimResults[0].ResultLow, long.StatSimResults[0].ResultLow},
		"attack power above": {merged.StatSimResults[0].ResultHigh, long.StatSimResults[0].ResultHigh},
	} {
		if !slices.Equal(dpsValues(pair[0]), dpsValues(pair[1])) {
			t.Errorf("%s DPS per iteration = %v, want %v", name, dpsValues(pair[0]), dpsValues(pair[1]))
		}
	}

	got, want := computeStatWeights(continued), computeStatWeights(long)
	if protoUnitStat(want.Dps.Weights, attackPowerUnitStat) == 0 {
		t.Fatal("attack power has no weight, so the runs cannot be told apart")
	}
	for field, pair := range map[string][2]*proto.UnitStats{
		"weight":  {got.Dps.Weights, want.Dps.Weights},
		"stdev":   {got.Dps.WeightsStdev, want.Dps.WeightsStdev},
		"samples": {got.SampleCounts, want.SampleCounts},
	} {
		if g, w := protoUnitStat(pair[0], attackPowerUnitStat), protoUnitStat(pair[1], attackPowerUnitStat); math.Abs(g-w) > 1e-9*math.Abs(w) {
			t.Errorf("attack power %s = %v, want %v", field, g, w)
		}
	}
}

// With a target ±, the runs go on until every DPS EP shows a ± at or under it or its runs reach the
// cap, and a run that needs more iterations gets them.
func TestStatWeightsTargetModeRunsToTheTargetOrTheCap(t *testing.T) {
	const probePerSide, multiplier = 10, 4
	weigh := func(target float64) *proto.StatWeightsResult {
		request := weaponTesterStatWeightsRequest(2 * probePerSide)
		request.StatsToWeigh = []proto.Stat{proto.Stat_StatAttackPower, proto.Stat_StatHitRating}
		if target > 0 {
			request.Target = &proto.StatWeightsTarget{EpCi: target, MaxIterationMultiplier: multiplier}
		}
		return runStatWeights(request, nil, simsignals.CreateSignals())
	}
	weighed := []stats.UnitStat{attackPowerUnitStat, hitRatingUnitStat}

	probe := weigh(0)
	target := math.Inf(1)
	for _, stat := range weighed {
		target = min(target, shownEpCi(probe, stat)/2)
	}
	if target == 0 {
		t.Fatal("a stat has no spread, so there is nothing to top up")
	}

	result := weigh(target)
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}
	for _, stat := range weighed {
		sides := map[stats.UnitStat]float64{attackPowerUnitStat: 2, hitRatingUnitStat: 1}[stat]
		samples := protoUnitStat(result.SampleCounts, stat)
		if samples <= protoUnitStat(probe.SampleCounts, stat) {
			t.Errorf("stat %d kept its probe's %v samples", stat, samples)
		}
		if ci := shownEpCi(result, stat); ci > target && samples != sides*probePerSide*multiplier {
			t.Errorf("stat %d shows ± %v over the target %v with %v samples, under the cap", stat, ci, target, samples)
		}
	}
}

var (
	intellectUnitStat  = stats.UnitStatFromStat(stats.Intellect)
	healingDpsPerPoint = map[stats.UnitStat]float64{agilityUnitStat: 1, hitRatingUnitStat: 3}
	healingHpsPerPoint = map[stats.UnitStat]float64{agilityUnitStat: 0.5, hitRatingUnitStat: 0.2, intellectUnitStat: 2}
)

// Agility and hit rating weighed with agility as the DPS reference and hpsReferenceStat as the HPS one.
func healingStatWeightsRequest(hpsReferenceStat *proto.Stat) *proto.StatWeightsRequest {
	return &proto.StatWeightsRequest{
		Player:           &proto.Player{},
		Encounter:        &proto.Encounter{},
		SimOptions:       &proto.SimOptions{Iterations: 2, RandomSeed: 1},
		StatsToWeigh:     []proto.Stat{proto.Stat_StatAgility, proto.Stat_StatHitRating},
		EpReferenceStat:  proto.Stat_StatAgility,
		HpsReferenceStat: hpsReferenceStat,
	}
}

// The runs over perSide iterations with intellect as the HPS reference. Every run's DPS differs from
// the baseline's by its stat's worth alone, so each DPS ± is 0, and each run's HPS carries noise of
// its own.
func healingStatWeightCalc(perSide int) *proto.StatWeightsCalcRequest {
	requests := buildStatWeightRequests(healingStatWeightsRequest(proto.Stat_StatIntellect.Enum()))
	run := func(index int, stat stats.UnitStat, mod float64) *proto.RaidSimResult {
		var dps, hps []float64
		for i := range perSide {
			dps = append(dps, syntheticBaselineDps+float64((i*7)%13)+healingDpsPerPoint[stat]*mod)
			hps = append(hps, syntheticBaselineDps+float64((i*5)%11)+healingHpsPerPoint[stat]*mod+float64(((i+1)*(index+3)*5)%7))
		}
		return statWeightRunWithHps(dps, hps)
	}

	calc := &proto.StatWeightsCalcRequest{BaseResult: run(0, 0, 0), EpReferenceStat: requests.EpReferenceStat, HpsReferenceStat: requests.HpsReferenceStat}
	for j, request := range requests.StatSimRequests {
		stat := stats.UnitStatFromIdx(int(request.StatData.UnitStat))
		result := &proto.StatWeightsStatResultData{StatData: request.StatData, ResultHigh: run(2*j+2, stat, request.StatData.ModHigh)}
		if request.RequestLow != nil {
			result.ResultLow = run(2*j+1, stat, request.StatData.ModLow)
		}
		calc.StatSimResults = append(calc.StatSimResults, result)
	}
	return calc
}

// A set HPS reference stat is simulated, below and above its baseline like the DPS one, and an unset
// one is the DPS reference stat.
func TestStatWeightsSimulateTheHpsReferenceStatWhenSet(t *testing.T) {
	set := buildStatWeightRequests(healingStatWeightsRequest(proto.Stat_StatIntellect.Enum()))
	index := slices.IndexFunc(set.StatSimRequests, func(request *proto.StatWeightsStatRequestData) bool {
		return request.StatData.UnitStat == int32(intellectUnitStat)
	})
	if index < 0 {
		t.Fatal("no sub-sim for intellect, the HPS reference stat")
	}
	if statData := set.StatSimRequests[index].StatData; statData.ModLow >= 0 || statData.ModHigh <= 0 {
		t.Errorf("intellect modifiers %v and %v, want one below and one above", statData.ModLow, statData.ModHigh)
	}

	unset := buildStatWeightRequests(healingStatWeightsRequest(nil))
	if len(unset.StatSimRequests) != len(set.StatSimRequests)-1 || simulatesStat(unset, intellectUnitStat) {
		t.Errorf("unset HPS reference simulates %d stats, want the weighed ones alone", len(unset.StatSimRequests))
	}
	if unset.HpsReferenceStat != proto.Stat_StatAgility {
		t.Errorf("unset HPS reference = %v, want the DPS reference", unset.HpsReferenceStat)
	}
}

// HPS EP divides by the HPS reference stat's HPS weight.
func TestStatWeightsDivideHpsEpByTheHpsReference(t *testing.T) {
	calc := healingStatWeightCalc(20)
	for _, ref := range []stats.UnitStat{intellectUnitStat, agilityUnitStat} {
		calc.HpsReferenceStat = proto.Stat(ref)
		result := computeStatWeights(calc)
		want := math.Abs(protoUnitStat(result.Hps.Weights, hitRatingUnitStat) / protoUnitStat(result.Hps.Weights, ref))
		if got := protoUnitStat(result.Hps.EpValues, hitRatingUnitStat); want == 0 || math.Abs(got-want) > 1e-12 {
			t.Errorf("hit rating HPS EP against stat %d = %v, want %v", ref, got, want)
		}
	}
}

// A stat whose DPS ± is under the target and whose HPS ± is over it is planned only when HPS is
// targeted, and its HPS ± divides by the HPS reference stat's weight.
func TestStatWeightTopUpPlanSizesTargetedHpsByTheHpsReference(t *testing.T) {
	const perSide = 20
	calc := healingStatWeightCalc(perSide)
	result := computeStatWeights(calc)
	samples := protoUnitStat(result.SampleCounts, agilityUnitStat)
	hpsEpStdev := func(stat stats.UnitStat, ref stats.UnitStat) float64 {
		return protoUnitStat(result.Hps.WeightsStdev, stat) / math.Abs(protoUnitStat(result.Hps.Weights, ref))
	}
	agilityHpsCi := statWeightZ90 * hpsEpStdev(agilityUnitStat, intellectUnitStat) / math.Sqrt(samples)
	if agilityHpsCi == 0 || !(shownEpCi(result, agilityUnitStat) < agilityHpsCi/4) {
		t.Fatalf("agility shows DPS ± %v and HPS ± %v: want the DPS one well under the HPS one", shownEpCi(result, agilityUnitStat), agilityHpsCi)
	}
	target := agilityHpsCi / 4
	plan := func(metrics []proto.StatWeightsTargetMetric) map[stats.UnitStat]int32 {
		return plannedIterations(planStatWeightTopUp(&proto.StatWeightsTopUpRequest{
			Calc:   calc,
			Target: &proto.StatWeightsTarget{EpCi: target, MaxIterationMultiplier: math.MaxInt32 / perSide, Metrics: metrics},
		}))
	}

	if got := plan(nil); len(got) != 0 {
		t.Errorf("DPS alone planned %v, want nothing: every DPS ± is 0", got)
	}

	got := plan([]proto.StatWeightsTargetMetric{proto.StatWeightsTargetMetric_StatWeightsTargetMetricDps, proto.StatWeightsTargetMetric_StatWeightsTargetMetricHps})
	sized := func(ref stats.UnitStat) int32 {
		return int32(math.Ceil(math.Pow(statWeightZ90*hpsEpStdev(agilityUnitStat, ref)/target, 2) / 2))
	}
	if sized(intellectUnitStat) == sized(agilityUnitStat) {
		t.Fatal("intellect and agility have the same HPS weight, so the references cannot be told apart")
	}
	if got[agilityUnitStat] != sized(intellectUnitStat) {
		t.Errorf("agility planned %d iterations per side, want %d from the HPS reference's weight", got[agilityUnitStat], sized(intellectUnitStat))
	}
}

func setChanceOfDeath(run *proto.RaidSimResult, chance float64) {
	if run != nil {
		run.RaidMetrics.Parties[0].Players[0].ChanceOfDeath = chance
	}
}

// The chance of death weights compare the first runs alone, so runs topped up past them, and a
// baseline topped up further than a stat, leave them as they were.
func TestStatWeightsReadTheChanceOfDeathFromTheFirstRuns(t *testing.T) {
	const perSide = 12
	requests := noisyStatWeightRequests()
	isHit := func(result *proto.StatWeightsStatResultData) bool {
		return result.StatData.UnitStat == int32(hitRatingUnitStat)
	}
	probe := func() *proto.StatWeightsCalcRequest {
		base, results := noisyStatWeightResults(requests, 0, perSide)
		setChanceOfDeath(base, 0.1)
		for _, result := range results {
			setChanceOfDeath(result.ResultHigh, 0.3)
			setChanceOfDeath(result.ResultLow, 0.05)
		}
		return &proto.StatWeightsCalcRequest{BaseResult: base, EpReferenceStat: requests.EpReferenceStat, StatSimResults: results}
	}

	topped := probe()
	segmentBase, segmentResults := noisyStatWeightResults(requests, perSide, 2*perSide)
	setChanceOfDeath(segmentBase, 0.9)
	segmentResults = slices.DeleteFunc(segmentResults, isHit)
	for _, result := range segmentResults {
		setChanceOfDeath(result.ResultHigh, 0.9)
		setChanceOfDeath(result.ResultLow, 0.9)
	}
	topped.ExtraSegments = []*proto.StatWeightsCalcSegment{{BaseResult: segmentBase, StatSimResults: segmentResults}}

	probeOnly := computeStatWeights(probe())
	toppedUp := computeStatWeights(topped)
	hitModHigh := topped.StatSimResults[slices.IndexFunc(topped.StatSimResults, isHit)].StatData.ModHigh
	if got, want := protoUnitStat(probeOnly.PDeath.Weights, hitRatingUnitStat), (0.3-0.1)/hitModHigh; math.Abs(got-want) > 1e-12 {
		t.Fatalf("hit rating chance of death weight = %v, want %v", got, want)
	}
	for _, stat := range []stats.UnitStat{agilityUnitStat, hitRatingUnitStat} {
		if got, want := protoUnitStat(toppedUp.PDeath.Weights, stat), protoUnitStat(probeOnly.PDeath.Weights, stat); got != want {
			t.Errorf("stat %d chance of death weight after top-ups = %v, want the first runs' %v", stat, got, want)
		}
	}
}

func TestStatWeightRequestsRunAtLeastOneIterationPerSide(t *testing.T) {
	requests := buildStatWeightRequests(&proto.StatWeightsRequest{
		Player:          &proto.Player{},
		Encounter:       &proto.Encounter{},
		SimOptions:      &proto.SimOptions{Iterations: 1, RandomSeed: 1},
		StatsToWeigh:    []proto.Stat{proto.Stat_StatAgility, proto.Stat_StatHitRating},
		EpReferenceStat: proto.Stat_StatAgility,
	})
	if got := requests.BaseRequest.SimOptions.Iterations; got != 1 {
		t.Errorf("baseline runs %d iterations, want 1", got)
	}
	for _, request := range requests.StatSimRequests {
		for _, simRequest := range []*proto.RaidSimRequest{request.RequestLow, request.RequestHigh} {
			if simRequest != nil && simRequest.SimOptions.Iterations != 1 {
				t.Errorf("stat %d runs %d iterations, want 1", request.StatData.UnitStat, simRequest.SimOptions.Iterations)
			}
		}
	}
}

// A stat run longer than the baseline has iterations with nothing to pair with, so the weights are
// refused and the plan asks for nothing more.
func TestStatWeightsRefuseAStatRunLongerThanTheBaseline(t *testing.T) {
	const perSide = 12
	requests := noisyStatWeightRequests()
	base, _ := noisyStatWeightResults(requests, 0, perSide)
	_, longResults := noisyStatWeightResults(requests, 0, 2*perSide)
	calc := &proto.StatWeightsCalcRequest{BaseResult: base, EpReferenceStat: requests.EpReferenceStat, StatSimResults: longResults}

	if result := computeStatWeights(calc); result.Error == nil {
		t.Error("weights computed for stat runs longer than the baseline, want an error")
	}
	if plan := planStatWeightTopUp(&proto.StatWeightsTopUpRequest{Calc: calc, Target: &proto.StatWeightsTarget{EpCi: 1e-9}}); len(plan.Stats) != 0 {
		t.Errorf("plan = %v, want no stats", plan)
	}
}
