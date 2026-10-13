package core

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	googleProto "google.golang.org/protobuf/proto"
)

const DTPSReferenceStat = stats.Armor

type UnitStats struct {
	Stats       stats.Stats
	PseudoStats []float64
}

func NewUnitStats() UnitStats {
	return UnitStats{
		PseudoStats: make([]float64, stats.PseudoStatsLen),
	}
}
func (s *UnitStats) AddStat(stat stats.UnitStat, value float64) {
	if stat.IsStat() {
		s.Stats[stat.StatIdx()] += value
	} else {
		s.PseudoStats[stat.PseudoStatIdx()] += value
	}
}
func (s *UnitStats) Get(stat stats.UnitStat) float64 {
	if stat.IsStat() {
		return s.Stats[stat.StatIdx()]
	} else {
		return s.PseudoStats[stat.PseudoStatIdx()]
	}
}

func (s *UnitStats) ToProto() *proto.UnitStats {
	return &proto.UnitStats{
		Stats:       s.Stats.ToProtoArray(),
		PseudoStats: s.PseudoStats,
		ApiVersion:  GetCurrentProtoVersion(),
	}
}

type StatWeightValues struct {
	Weights       UnitStats
	WeightsStdev  UnitStats
	EpValues      UnitStats
	EpValuesStdev UnitStats
}

func NewStatWeightValues() StatWeightValues {
	return StatWeightValues{
		Weights:       NewUnitStats(),
		WeightsStdev:  NewUnitStats(),
		EpValues:      NewUnitStats(),
		EpValuesStdev: NewUnitStats(),
	}
}

func (swv *StatWeightValues) ToProto() *proto.StatWeightValues {
	return &proto.StatWeightValues{
		Weights:       swv.Weights.ToProto(),
		WeightsStdev:  swv.WeightsStdev.ToProto(),
		EpValues:      swv.EpValues.ToProto(),
		EpValuesStdev: swv.EpValuesStdev.ToProto(),
	}
}

type StatWeightsResult struct {
	Dps    StatWeightValues
	Hps    StatWeightValues
	Tps    StatWeightValues
	Dtps   StatWeightValues
	Tmi    StatWeightValues
	PDeath StatWeightValues

	SampleCounts UnitStats
	BaseDps      *proto.DistributionMetrics
}

func NewStatWeightsResult() *StatWeightsResult {
	return &StatWeightsResult{
		Dps:    NewStatWeightValues(),
		Hps:    NewStatWeightValues(),
		Tps:    NewStatWeightValues(),
		Dtps:   NewStatWeightValues(),
		Tmi:    NewStatWeightValues(),
		PDeath: NewStatWeightValues(),

		SampleCounts: NewUnitStats(),
	}
}

func (swr *StatWeightsResult) ToProto() *proto.StatWeightsResult {
	return &proto.StatWeightsResult{
		Dps:    swr.Dps.ToProto(),
		Hps:    swr.Hps.ToProto(),
		Tps:    swr.Tps.ToProto(),
		Dtps:   swr.Dtps.ToProto(),
		Tmi:    swr.Tmi.ToProto(),
		PDeath: swr.PDeath.ToProto(),

		SampleCounts: swr.SampleCounts.ToProto(),
		BaseDps:      swr.BaseDps,
	}
}

func buildStatWeightRequests(swr *proto.StatWeightsRequest) *proto.StatWeightRequestsData {
	if swr.Player.BonusStats == nil {
		swr.Player.BonusStats = &proto.UnitStats{}
	}
	if swr.Player.BonusStats.Stats == nil {
		swr.Player.BonusStats.Stats = make([]float64, stats.ProtoStatsLen)
	}
	if swr.Player.BonusStats.PseudoStats == nil {
		swr.Player.BonusStats.PseudoStats = make([]float64, stats.PseudoStatsLen)
	}

	raidProto := SinglePlayerRaidProto(swr.Player, swr.PartyBuffs, swr.RaidBuffs, swr.Debuffs)
	raidProto.Tanks = swr.Tanks

	swr.SimOptions.SaveAllValues = true

	// Cut in half since we're doing above and below separately.
	// This number needs to be the same for the baseline sim too, so that RNG lines up perfectly.
	// At least 1, since a sim split across workers fails on 0.
	swr.SimOptions.Iterations = max(swr.SimOptions.Iterations/2, 1)

	// Make sure an RNG seed is always set because it gives more consistent results.
	// When there is no user-supplied seed it needs to be a randomly-selected seed
	// though, so that run-run differences still exist.
	if swr.SimOptions.RandomSeed == 0 {
		swr.SimOptions.RandomSeed = time.Now().UnixNano()
	}

	// Reduce variance even more by using test-level RNG controls.
	swr.SimOptions.UseLabeledRands = true

	hpsReferenceStat := swr.EpReferenceStat
	if swr.HpsReferenceStat != nil {
		hpsReferenceStat = *swr.HpsReferenceStat
	}

	swBaseResponse := &proto.StatWeightRequestsData{
		BaseRequest: &proto.RaidSimRequest{
			Raid:       raidProto,
			Encounter:  swr.Encounter,
			SimOptions: swr.SimOptions,
		},
		EpReferenceStat:  swr.EpReferenceStat,
		HpsReferenceStat: hpsReferenceStat,
		StatSimRequests:  []*proto.StatWeightsStatRequestData{},
	}

	// Do half the iterations with a positive, and half with a negative value for better accuracy.
	const defaultStatMod = 10.0
	statModsLow := make([]float64, stats.UnitStatsLen)
	statModsHigh := make([]float64, stats.UnitStatsLen)

	// Make sure the reference stats are included.
	for _, referenceStat := range []proto.Stat{swr.EpReferenceStat, hpsReferenceStat} {
		statModsLow[referenceStat] = -defaultStatMod
		statModsHigh[referenceStat] = defaultStatMod
	}

	statsToWeigh := stats.ProtoArrayToStatsList(swr.StatsToWeigh)
	for _, s := range statsToWeigh {
		stat := stats.UnitStatFromStat(s)
		statMod := defaultStatMod
		if stat.EqualsStat(stats.Armor) || stat.EqualsStat(stats.BonusArmor) || stat.EqualsStat(stats.ArmorPenetration) {
			statMod = defaultStatMod * 10
		}
		statModsHigh[stat] = statMod
		statModsLow[stat] = -statMod

		if stat.EqualsStat(stats.HitRating) {
			statModsLow[stats.HitRating] = 0
		} else if stat.EqualsStat(stats.ExpertiseRating) {
			statModsLow[stats.ExpertiseRating] = 0
		}
	}

	for _, s := range swr.PseudoStatsToWeigh {
		stat := stats.UnitStatFromPseudoStat(s)
		statName := proto.PseudoStat_name[int32(s)]

		// Scale down the stat increment depending on the type of PseudoStat
		statMod := defaultStatMod

		if stat.EqualsPseudoStat(proto.PseudoStat_PseudoStatMeleeHitPercent) ||
			stat.EqualsPseudoStat(proto.PseudoStat_PseudoStatRangedHitPercent) ||
			stat.EqualsPseudoStat(proto.PseudoStat_PseudoStatSpellHitPercent) {
			statMod /= HitRatingPerHitPercent
		} else if stat.EqualsPseudoStat(proto.PseudoStat_PseudoStatExpertisePercent) {
			statMod /= ExpertiseRatingPerExpertisePercent
		} else if strings.Contains(statName, "SchoolHit") {
			statMod /= HitRatingPerHitPercent
		} else if strings.Contains(statName, "Crit") {
			statMod /= CritRatingPerCritPercent
		} else if strings.Contains(statName, "Dps") {
			statMod *= 0.5
		} else {
			panic(fmt.Sprintf("Unsupported PseudoStat in stat weights request: %s", statName))
		}

		statModsHigh[stat] = statMod
		statModsLow[stat] = -statMod
	}

	for i := range statModsLow {
		stat := stats.UnitStatFromIdx(i)
		if statModsLow[stat] == 0 && statModsHigh[stat] == 0 {
			continue
		}

		var lowSimRequest *proto.RaidSimRequest
		if statModsLow[stat] != 0 {
			lowSimRequest = googleProto.Clone(swBaseResponse.BaseRequest).(*proto.RaidSimRequest)
			stat.AddToStatsProto(lowSimRequest.Raid.Parties[0].Players[0].BonusStats, statModsLow[stat])
		}

		highSimRequest := googleProto.Clone(swBaseResponse.BaseRequest).(*proto.RaidSimRequest)
		stat.AddToStatsProto(highSimRequest.Raid.Parties[0].Players[0].BonusStats, statModsHigh[stat])

		swBaseResponse.StatSimRequests = append(swBaseResponse.StatSimRequests, &proto.StatWeightsStatRequestData{
			StatData: &proto.StatWeightsStatData{
				UnitStat: int32(stat),
				ModLow:   statModsLow[stat],
				ModHigh:  statModsHigh[stat],
			},
			RequestLow:  lowSimRequest,
			RequestHigh: highSimRequest,
		})
	}

	return swBaseResponse
}

const (
	statWeightZ90            = 1.645 // The two-sided 90% interval, the ± shown behind each EP.
	statWeightMaxTopUpPasses = 3     // The first top-up, which sizes from the probe alone, and two that resize from the runs so far.
)

// One result holding the first player's per-iteration metrics of every run, iterations in run order.
// Its chance of death is the first run's: one number per run cannot pair by iteration, and the first
// runs of the baseline and of every stat share their iterations.
func concatStatWeightRuns(runs []*proto.RaidSimResult) *proto.RaidSimResult {
	if len(runs) == 1 {
		return runs[0]
	}
	players := MapSlice(runs, func(run *proto.RaidSimResult) *proto.UnitMetrics { return run.RaidMetrics.Parties[0].Players[0] })
	metric := func(get func(*proto.UnitMetrics) *proto.DistributionMetrics) *proto.DistributionMetrics {
		var iterations int32
		for _, player := range players {
			iterations += get(player).AggregatorData.N
		}
		merged := &proto.DistributionMetrics{
			Min:            math.MaxFloat64,
			MinSeed:        math.MaxInt64,
			Hist:           make(map[int32]int32),
			AllValues:      make([]float64, 0, iterations),
			AggregatorData: &proto.AggregatorData{},
		}
		for i, player := range players {
			add := get(player)
			CombineDistributionMetrics(merged, add, i == len(players)-1, float64(add.AggregatorData.N)/float64(iterations))
		}
		return merged
	}

	merged := &proto.UnitMetrics{
		Dps:           metric(func(p *proto.UnitMetrics) *proto.DistributionMetrics { return p.Dps }),
		Hps:           metric(func(p *proto.UnitMetrics) *proto.DistributionMetrics { return p.Hps }),
		Threat:        metric(func(p *proto.UnitMetrics) *proto.DistributionMetrics { return p.Threat }),
		Dtps:          metric(func(p *proto.UnitMetrics) *proto.DistributionMetrics { return p.Dtps }),
		Tmi:           metric(func(p *proto.UnitMetrics) *proto.DistributionMetrics { return p.Tmi }),
		ChanceOfDeath: players[0].ChanceOfDeath,
	}
	return &proto.RaidSimResult{RaidMetrics: &proto.RaidMetrics{Parties: []*proto.PartyMetrics{{Players: []*proto.UnitMetrics{merged}}}}}
}

// Appends each extra segment's iterations to the runs it continues. The merged request is new, so
// the input can take another segment and be merged again.
func mergeStatWeightSegments(swcr *proto.StatWeightsCalcRequest) *proto.StatWeightsCalcRequest {
	if len(swcr.ExtraSegments) == 0 {
		return swcr
	}

	baseRuns := []*proto.RaidSimResult{swcr.BaseResult}
	lowRuns := make(map[int32][]*proto.RaidSimResult)
	highRuns := make(map[int32][]*proto.RaidSimResult)
	for _, segment := range swcr.ExtraSegments {
		if segment.BaseResult != nil {
			baseRuns = append(baseRuns, segment.BaseResult)
		}
		for _, statResult := range segment.StatSimResults {
			stat := statResult.StatData.UnitStat
			if statResult.ResultLow != nil {
				lowRuns[stat] = append(lowRuns[stat], statResult.ResultLow)
			}
			highRuns[stat] = append(highRuns[stat], statResult.ResultHigh)
		}
	}

	merged := &proto.StatWeightsCalcRequest{
		BaseResult:       concatStatWeightRuns(baseRuns),
		EpReferenceStat:  swcr.EpReferenceStat,
		HpsReferenceStat: swcr.HpsReferenceStat,
	}
	for _, statResult := range swcr.StatSimResults {
		stat := statResult.StatData.UnitStat
		mergedResult := &proto.StatWeightsStatResultData{
			StatData:   statResult.StatData,
			ResultHigh: concatStatWeightRuns(append([]*proto.RaidSimResult{statResult.ResultHigh}, highRuns[stat]...)),
		}
		if statResult.ResultLow != nil {
			mergedResult.ResultLow = concatStatWeightRuns(append([]*proto.RaidSimResult{statResult.ResultLow}, lowRuns[stat]...))
		}
		merged.StatSimResults = append(merged.StatSimResults, mergedResult)
	}
	return merged
}

// Whether a run's values equal the baseline's first iterations, which ran at the same seeds.
func matchesBaselinePrefix(mod *proto.UnitMetrics, baseline *proto.UnitMetrics) bool {
	for _, pair := range [][2]*proto.DistributionMetrics{{mod.Dps, baseline.Dps}, {mod.Hps, baseline.Hps}, {mod.Tmi, baseline.Tmi}} {
		modValues := pair[0].AllValues
		if !slices.Equal(modValues, pair[1].AllValues[:len(modValues)]) {
			return false
		}
	}
	return true
}

func computeStatWeights(swcr *proto.StatWeightsCalcRequest) *proto.StatWeightsResult {
	result, err := calcStatWeights(mergeStatWeightSegments(swcr))
	if err != nil {
		return &proto.StatWeightsResult{Error: err}
	}
	return result.ToProto()
}

// Each stat run pairs its iterations with the baseline's at the same seeds, so a run that stopped
// early pairs with the baseline's first iterations.
func calcStatWeights(swcr *proto.StatWeightsCalcRequest) (*StatWeightsResult, *proto.ErrorOutcome) {
	haveRefStat := false
	for _, statResult := range swcr.StatSimResults {
		if statResult.StatData.UnitStat == int32(swcr.EpReferenceStat) {
			haveRefStat = true
			break
		}
	}
	if !haveRefStat {
		return nil, &proto.ErrorOutcome{Message: "No result for reference stat exists!"}
	}

	result := NewStatWeightsResult()
	baseDps := swcr.BaseResult.RaidMetrics.Parties[0].Players[0].Dps
	result.BaseDps = &proto.DistributionMetrics{
		Avg:            baseDps.Avg,
		Stdev:          baseDps.Stdev,
		AggregatorData: baseDps.AggregatorData,
	}

	for _, statResult := range swcr.StatSimResults {
		stat := stats.UnitStatFromIdx(int(statResult.StatData.UnitStat))

		baselinePlayer := swcr.BaseResult.RaidMetrics.Parties[0].Players[0]
		modPlayerHigh := statResult.ResultHigh.RaidMetrics.Parties[0].Players[0]
		hasLowMeasurement := statResult.ResultLow != nil
		var modPlayerLow *proto.UnitMetrics
		if hasLowMeasurement {
			modPlayerLow = statResult.ResultLow.RaidMetrics.Parties[0].Players[0]
		} else {
			modPlayerLow = baselinePlayer
		}
		// Every metric of a run has the same number of iterations.
		if len(modPlayerHigh.Dps.AllValues) > len(baselinePlayer.Dps.AllValues) || len(modPlayerLow.Dps.AllValues) > len(baselinePlayer.Dps.AllValues) {
			return nil, &proto.ErrorOutcome{Message: "A stat run has more iterations than the baseline!"}
		}

		modLow, modHigh := statResult.StatData.ModLow, statResult.StatData.ModHigh
		if stat.IsPseudoStat() {
			statName := proto.PseudoStat_name[int32(stat.PseudoStatIdx())]
			if strings.Contains(statName, "SchoolHit") {
				modLow *= HitRatingPerHitPercent
				modHigh *= HitRatingPerHitPercent
			}
		}

		// Check for hard caps. Hard caps will have results identical to the baseline because RNG is fixed.
		// When we find a hard-capped stat, just skip it (will return 0).
		if matchesBaselinePrefix(modPlayerHigh, baselinePlayer) {
			continue
		}

		calcWeightResults := func(baselineMetrics *proto.DistributionMetrics, modLowMetrics *proto.DistributionMetrics, modHighMetrics *proto.DistributionMetrics, weightResults *StatWeightValues) int {
			var lo, hi aggregator
			if hasLowMeasurement {
				for i := range modLowMetrics.AllValues {
					lo.add(modLowMetrics.AllValues[i] - baselineMetrics.AllValues[i])
				}
				lo.scale(1 / modLow)
			}

			for i := range modHighMetrics.AllValues {
				hi.add(modHighMetrics.AllValues[i] - baselineMetrics.AllValues[i])
			}
			hi.scale(1 / modHigh)

			var aggr = &hi
			if hasLowMeasurement {
				aggr = aggr.merge(&lo)
			}

			mean, stdev := aggr.meanAndStdDev()

			weightResults.Weights.AddStat(stat, mean)
			weightResults.WeightsStdev.AddStat(stat, stdev)
			return aggr.n
		}

		samples := calcWeightResults(baselinePlayer.Dps, modPlayerLow.Dps, modPlayerHigh.Dps, &result.Dps)
		result.SampleCounts.AddStat(stat, float64(samples))
		calcWeightResults(baselinePlayer.Hps, modPlayerLow.Hps, modPlayerHigh.Hps, &result.Hps)
		calcWeightResults(baselinePlayer.Threat, modPlayerLow.Threat, modPlayerHigh.Threat, &result.Tps)
		calcWeightResults(baselinePlayer.Dtps, modPlayerLow.Dtps, modPlayerHigh.Dtps, &result.Dtps)
		calcWeightResults(baselinePlayer.Tmi, modPlayerLow.Tmi, modPlayerHigh.Tmi, &result.Tmi)

		meanLow := (modPlayerLow.ChanceOfDeath - baselinePlayer.ChanceOfDeath) / modLow
		meanHigh := (modPlayerHigh.ChanceOfDeath - baselinePlayer.ChanceOfDeath) / modHigh

		pDeathAvg := meanHigh
		if hasLowMeasurement {
			pDeathAvg += meanLow
			pDeathAvg /= 2
		}

		result.PDeath.Weights.AddStat(stat, pDeathAvg)
		result.PDeath.WeightsStdev.AddStat(stat, 0)
	}

	referenceStat := stats.Stat(swcr.EpReferenceStat)

	// Compute EP results.
	for _, statData := range swcr.StatSimResults {
		stat := stats.UnitStatFromIdx(int(statData.StatData.UnitStat))

		calcEpResults := func(weightResults *StatWeightValues, refStat stats.Stat) {
			if weightResults.Weights.Stats[refStat] == 0 {
				return
			}
			mean := weightResults.Weights.Get(stat) / weightResults.Weights.Stats[refStat]
			stdev := weightResults.WeightsStdev.Get(stat) / math.Abs(weightResults.Weights.Stats[refStat])
			weightResults.EpValues.AddStat(stat, math.Abs(mean))
			weightResults.EpValuesStdev.AddStat(stat, math.Abs(stdev))
		}

		calcEpResults(&result.Dps, referenceStat)
		calcEpResults(&result.Hps, stats.Stat(swcr.HpsReferenceStat))
		calcEpResults(&result.Tps, referenceStat)
		calcEpResults(&result.Dtps, DTPSReferenceStat)
		calcEpResults(&result.Tmi, DTPSReferenceStat)
		calcEpResults(&result.PDeath, DTPSReferenceStat)
	}

	return result, nil
}

// Per-side iteration totals that bring each stat's shown EP ± down to the target in every targeted
// metric, sized by the metric that needs the most. A metric is skipped for a stat whose weight or
// reference weight in it is 0, a stat at or under the target in all of them is left out, and the
// baseline runs as far as the furthest stat. The sides divide the total because the low and high
// runs pool their samples. A set multiplier caps each run at that multiple of the first pass's
// iterations, and the plan is empty once statWeightMaxTopUpPasses segments have run.
func planStatWeightTopUp(request *proto.StatWeightsTopUpRequest) *proto.StatWeightsTopUpPlan {
	target := request.Target
	calc := mergeStatWeightSegments(request.Calc)
	plan := &proto.StatWeightsTopUpPlan{
		BaselineIterations: int32(len(calc.BaseResult.RaidMetrics.Parties[0].Players[0].Dps.AllValues)),
	}
	if target == nil || len(request.Calc.ExtraSegments) >= statWeightMaxTopUpPasses {
		return plan
	}
	result, err := calcStatWeights(calc)
	if err != nil {
		return plan
	}

	// Without a multiplier the int32 range still bounds the count, so the conversion below stays defined.
	maxIterationsPerSide := float64(math.MaxInt32)
	if target.MaxIterationMultiplier > 0 {
		maxIterationsPerSide = min(maxIterationsPerSide, float64(len(request.Calc.BaseResult.RaidMetrics.Parties[0].Players[0].Dps.AllValues))*float64(target.MaxIterationMultiplier))
	}
	targetMetrics := target.Metrics
	if len(targetMetrics) == 0 {
		targetMetrics = []proto.StatWeightsTargetMetric{proto.StatWeightsTargetMetric_StatWeightsTargetMetricDps}
	}
	metricValues := func(metric proto.StatWeightsTargetMetric) (*StatWeightValues, stats.Stat) {
		switch metric {
		case proto.StatWeightsTargetMetric_StatWeightsTargetMetricHps:
			return &result.Hps, stats.Stat(calc.HpsReferenceStat)
		case proto.StatWeightsTargetMetric_StatWeightsTargetMetricTps:
			return &result.Tps, stats.Stat(calc.EpReferenceStat)
		default:
			return &result.Dps, stats.Stat(calc.EpReferenceStat)
		}
	}

	for _, statResult := range calc.StatSimResults {
		stat := stats.UnitStatFromIdx(int(statResult.StatData.UnitStat))
		samples := result.SampleCounts.Get(stat)
		epStdev := 0.0
		for _, metric := range targetMetrics {
			values, refStat := metricValues(metric)
			refWeight := values.Weights.Stats[refStat]
			if values.Weights.Get(stat) == 0 || refWeight == 0 {
				continue
			}
			metricEpStdev := values.WeightsStdev.Get(stat) / math.Abs(refWeight)
			if statWeightZ90*metricEpStdev/math.Sqrt(samples) > target.EpCi {
				epStdev = max(epStdev, metricEpStdev)
			}
		}
		if epStdev == 0 {
			continue
		}

		sides := 1.0
		if statResult.ResultLow != nil {
			sides = 2
		}
		current := int32(len(statResult.ResultHigh.RaidMetrics.Parties[0].Players[0].Dps.AllValues))
		needed := int32(math.Min(math.Ceil(math.Pow(statWeightZ90*epStdev/target.EpCi, 2)/sides), maxIterationsPerSide))
		if needed <= current {
			continue
		}

		plan.Stats = append(plan.Stats, &proto.StatWeightsStatIterations{UnitStat: statResult.StatData.UnitStat, Iterations: needed})
		plan.BaselineIterations = max(plan.BaselineIterations, needed)
	}
	return plan
}

// Run stat weight sims and compute weights.
func runStatWeights(request *proto.StatWeightsRequest, progress chan *proto.ProgressMetrics, signals simsignals.Signals) *proto.StatWeightsResult {
	requestData := buildStatWeightRequests(request)

	var iterationsTotal int32 = requestData.BaseRequest.SimOptions.Iterations
	var iterationsDone int32 = 0
	var simsTotal int32 = 1
	var simsCompleted int32 = 0

	for _, reqData := range requestData.StatSimRequests {
		if reqData.RequestLow != nil {
			iterationsTotal += reqData.RequestLow.SimOptions.Iterations
			simsTotal += 1
		}
		iterationsTotal += reqData.RequestHigh.SimOptions.Iterations
		simsTotal += 1
	}

	waitForResult := func(srcProgressChannel chan *proto.ProgressMetrics) *proto.RaidSimResult {
		var lastCompleted int32 = 0
		for metrics := range srcProgressChannel {
			iterationsDone += metrics.CompletedIterations - lastCompleted
			lastCompleted = metrics.CompletedIterations

			if progress != nil {
				progress <- &proto.ProgressMetrics{
					TotalIterations:     iterationsTotal,
					CompletedIterations: iterationsDone,
					CompletedSims:       simsCompleted,
					TotalSims:           simsTotal,
				}
			}

			if metrics.FinalRaidResult != nil {
				simsCompleted++
				return metrics.FinalRaidResult
			}
		}
		return nil
	}

	simFunc := runSimConcurrent
	// Don't use go threads in wasm, it just adds more overhead and makes the worker more unresponsive.
	if IsRunningInWasm() {
		simFunc = RunSim
	}

	baseProgress := make(chan *proto.ProgressMetrics, 100)
	go simFunc(requestData.BaseRequest, baseProgress, signals)
	baselineResult := waitForResult(baseProgress)
	if baselineResult.Error != nil {
		return &proto.StatWeightsResult{Error: baselineResult.Error}
	}

	statResults := []*proto.StatWeightsStatResultData{}

	for _, reqData := range requestData.StatSimRequests {
		var lowRes *proto.RaidSimResult
		if reqData.RequestLow != nil {
			lowProgress := make(chan *proto.ProgressMetrics, 100)
			go simFunc(reqData.RequestLow, lowProgress, signals)
			lowRes = waitForResult(lowProgress)
			if lowRes.Error != nil {
				return &proto.StatWeightsResult{Error: lowRes.Error}
			}
		}

		highProgress := make(chan *proto.ProgressMetrics, 100)
		go simFunc(reqData.RequestHigh, highProgress, signals)
		highRes := waitForResult(highProgress)
		if highRes.Error != nil {
			return &proto.StatWeightsResult{Error: highRes.Error}
		}

		statResults = append(statResults, &proto.StatWeightsStatResultData{
			StatData:   reqData.StatData,
			ResultLow:  lowRes,
			ResultHigh: highRes,
		})
	}

	calc := &proto.StatWeightsCalcRequest{
		BaseResult:       baselineResult,
		EpReferenceStat:  requestData.EpReferenceStat,
		HpsReferenceStat: requestData.HpsReferenceStat,
		StatSimResults:   statResults,
	}
	if request.Target == nil {
		return computeStatWeights(calc)
	}

	// Iteration i runs at seed RandomSeed+i, so a run continued at the seed after its last iteration
	// repeats what one longer run would have simulated.
	continueRun := func(simRequest *proto.RaidSimRequest, done int32, iterations int32) *proto.RaidSimResult {
		continued := googleProto.Clone(simRequest).(*proto.RaidSimRequest)
		continued.SimOptions.Iterations = iterations
		continued.SimOptions.RandomSeed += int64(done)
		continuedProgress := make(chan *proto.ProgressMetrics, 100)
		go simFunc(continued, continuedProgress, signals)
		return waitForResult(continuedProgress)
	}

	probeIterations := requestData.BaseRequest.SimOptions.Iterations
	statRequests := make(map[int32]*proto.StatWeightsStatRequestData)
	statIterations := make(map[int32]int32)
	for _, reqData := range requestData.StatSimRequests {
		statRequests[reqData.StatData.UnitStat] = reqData
		statIterations[reqData.StatData.UnitStat] = probeIterations
	}
	baselineIterations := probeIterations

	for {
		plan := planStatWeightTopUp(&proto.StatWeightsTopUpRequest{Calc: calc, Target: request.Target})
		if len(plan.Stats) == 0 {
			break
		}

		baselineDelta := plan.BaselineIterations - baselineIterations
		if baselineDelta > 0 {
			iterationsTotal += baselineDelta
			simsTotal++
		}
		for _, planned := range plan.Stats {
			runs := TernaryInt32(statRequests[planned.UnitStat].RequestLow != nil, 2, 1)
			iterationsTotal += runs * (planned.Iterations - statIterations[planned.UnitStat])
			simsTotal += runs
		}

		segment := &proto.StatWeightsCalcSegment{}
		if baselineDelta > 0 {
			segment.BaseResult = continueRun(requestData.BaseRequest, baselineIterations, baselineDelta)
			if segment.BaseResult.Error != nil {
				return &proto.StatWeightsResult{Error: segment.BaseResult.Error}
			}
			baselineIterations = plan.BaselineIterations
		}
		for _, planned := range plan.Stats {
			reqData := statRequests[planned.UnitStat]
			done := statIterations[planned.UnitStat]
			delta := planned.Iterations - done

			var lowRes *proto.RaidSimResult
			if reqData.RequestLow != nil {
				lowRes = continueRun(reqData.RequestLow, done, delta)
				if lowRes.Error != nil {
					return &proto.StatWeightsResult{Error: lowRes.Error}
				}
			}
			highRes := continueRun(reqData.RequestHigh, done, delta)
			if highRes.Error != nil {
				return &proto.StatWeightsResult{Error: highRes.Error}
			}

			segment.StatSimResults = append(segment.StatSimResults, &proto.StatWeightsStatResultData{
				StatData:   reqData.StatData,
				ResultLow:  lowRes,
				ResultHigh: highRes,
			})
			statIterations[planned.UnitStat] = planned.Iterations
		}
		calc.ExtraSegments = append(calc.ExtraSegments, segment)
	}

	return computeStatWeights(calc)
}
