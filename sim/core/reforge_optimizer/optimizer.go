package reforgeoptimizer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

var reforgeOptimizeRequestID atomic.Uint64

func OptimizeAsync(request *proto.ReforgeOptimizeRequest, signals simsignals.Signals) *proto.ReforgeOptimizeResult {
	requestID := reforgeOptimizeRequestID.Add(1)
	startedAt := time.Now()
	debug := request.GetDebug()
	logAbort := request.GetMode() != proto.ReforgeOptimizeMode_ReforgeOptimizeModeBulk || debug
	if debug {
		log.Printf("[reforgeOptimize:%d] started debug=%t", requestID, debug)
		logRequestInput(requestID, request)
	}

	if request.Raid == nil || len(request.Raid.Parties) == 0 || len(request.Raid.Parties[0].Players) == 0 {
		log.Printf("[reforgeOptimize:%d] failed after %s: missing player", requestID, time.Since(startedAt))
		return optimizeError("Reforge optimizer requires a raid with player 0.")
	}
	if request.Raid.Parties[0].Players[0].Equipment == nil {
		log.Printf("[reforgeOptimize:%d] failed after %s: missing baseline gear", requestID, time.Since(startedAt))
		return optimizeError("Reforge optimizer requires baseline gear.")
	}
	if signals.Abort.IsTriggered() {
		return optimizeAborted()
	}

	solveStartedAt := time.Now()
	optimizer, err := newReforgeOptimizer(request, signals)
	var optimizedGear *proto.EquipmentSpec
	var score float64
	if err == nil {
		optimizedGear, score, err = optimizer.optimizeReforges()
	}
	if err != nil {
		if errors.Is(err, context.Canceled) {
			if logAbort {
				log.Printf("[reforgeOptimize:%d] aborted solving after %s", requestID, time.Since(startedAt))
			}
			return optimizeAborted()
		}
		gear := request.GetRaid().GetParties()[0].GetPlayers()[0].GetEquipment()
		gearJSON, _ := protojson.Marshal(gear)
		log.Printf("[reforgeOptimize:%d] HiGHS failed after %s: %s gear=%s", requestID, time.Since(solveStartedAt), err.Error(), gearJSON)
		return optimizeError(fmt.Sprintf("HiGHS reforge optimizer failed: %s", err.Error()))
	}
	if debug {
		log.Printf("[reforgeOptimize:%d] HiGHS solved in %s score=%.3f", requestID, time.Since(solveStartedAt), score)
	}
	if signals.Abort.IsTriggered() {
		if logAbort {
			log.Printf("[reforgeOptimize:%d] aborted after solving in %s", requestID, time.Since(startedAt))
		}
		return optimizeAborted()
	}

	isBulk := request.GetMode() == proto.ReforgeOptimizeMode_ReforgeOptimizeModeBulk
	var optimizedPlayerStats *proto.PlayerStats
	if !isBulk || debug {
		optimizedRaid := googleProto.Clone(request.Raid).(*proto.Raid)
		optimizedRaid.Parties[0].Players[0].Equipment = optimizedGear
		optimizedResult := computeReforgeStats(&proto.ComputeStatsRequest{Raid: optimizedRaid})
		if optimizedResult.ErrorResult != "" {
			log.Printf("[reforgeOptimize:%d] failed computing optimized stats after %s: %s", requestID, time.Since(startedAt), optimizedResult.ErrorResult)
			return optimizeError(optimizedResult.ErrorResult)
		}
		optimizedPlayerStats = optimizedResult.RaidStats.Parties[0].Players[0]
	}
	if !isBulk {
		log.Printf("[Reforge Optimizer] Reforge optimization completed requestID=%d total=%s score=%.3f", requestID, time.Since(startedAt), score)
	}

	return &proto.ReforgeOptimizeResult{
		OptimizedGear:        optimizedGear,
		OptimizedPlayerStats: optimizedPlayerStats,
		Score:                score,
		PassesDone:           1,
	}
}

type reforgeOptimizer struct {
	request  *proto.ReforgeOptimizeRequest
	settings *proto.ReforgeSettings
	player   *proto.Player
	signals  simsignals.Signals

	frozenSlots    map[proto.ItemSlot]bool
	undershootCaps core.UnitStats

	// statDeps is the player's build-phase StatDependencyManager (ComputeStatDependencies). It
	// resolves every stat conversion the sim models and is used by resolveStatDelta to compute
	// each LP variable's cap-space coefficients (the FULL dependency graph), separately from the
	// EP-calibrated objective coefficients produced by applyReforgeStat.
	statDeps *stats.StatDependencyManager
	sheet    core.SheetAvoidance

	baseRaidProto *proto.Raid
	baseGear      *proto.EquipmentSpec
	baseStats     core.UnitStats
}

// newReforgeOptimizer builds the optimizer context from the request: computes the base stats and
// the stat dependency manager for the input gear.
func newReforgeOptimizer(request *proto.ReforgeOptimizeRequest, signals simsignals.Signals) (*reforgeOptimizer, error) {
	settings := request.GetSettings()
	if settings == nil {
		settings = &proto.ReforgeSettings{}
	} else {
		settings = googleProto.Clone(settings).(*proto.ReforgeSettings)
	}

	request = googleProto.Clone(request).(*proto.ReforgeOptimizeRequest)
	request.Settings = settings

	o := &reforgeOptimizer{
		request:  request,
		settings: settings,
		signals:  signals,

		frozenSlots:    frozenItemSlots(settings),
		undershootCaps: protoToCoreUnitStats(request.GetUndershootCaps()),
	}
	if err := o.setBaseRaid(googleProto.Clone(request.Raid).(*proto.Raid)); err != nil {
		return nil, err
	}
	return o, nil
}

// setBaseRaid makes baseRaid the baseline: its player, gear, base stats and stat dependency manager.
func (o *reforgeOptimizer) setBaseRaid(baseRaid *proto.Raid) error {
	player := baseRaid.Parties[0].Players[0]

	// One environment build yields both FinalStats and the finalized StatDependencyManager,
	// instead of building the character twice for the same base raid.
	baseResult, baseSDM, sheet := computeReforgeStatsAndDeps(&proto.ComputeStatsRequest{Raid: baseRaid})
	if baseResult.ErrorResult != "" {
		return errors.New(baseResult.ErrorResult)
	}
	if o.signals.Abort.IsTriggered() {
		return context.Canceled
	}

	o.player = player
	o.statDeps = baseSDM
	o.sheet = sheet
	o.baseRaidProto = baseRaid
	o.baseGear = cloneEquipmentSpec(player.Equipment)
	o.baseStats = protoToCoreUnitStats(baseResult.RaidStats.Parties[0].Players[0].FinalStats)
	return nil
}

// optimizeReforges picks one static enchant per listed slot. The baseline is the input gear with
// those slots' enchants cleared, so each candidate's coefficients are its full contribution.
func (o *reforgeOptimizer) optimizeReforges() (*proto.EquipmentSpec, float64, error) {
	candidates := o.enchantCandidates(o.baseGear)
	if len(candidates) == 0 {
		return o.baseGear, 0, nil
	}
	if err := o.rebaseWithoutEnchants(candidates); err != nil {
		return nil, 0, err
	}

	reforgeCaps := computeStatCapsDelta(o.baseStats, protoToCoreUnitStats(o.settings.GetStatCaps()))
	var softCapConfigs []*proto.StatCapConfig
	if o.settings.GetUseSoftCapBreakpoints() {
		softCapConfigs = o.request.GetSoftCaps()
	}
	reforgeSoftCaps := computeReforgeSoftCaps(o.baseStats, softCapConfigs)
	weights := checkWeights(protoToCoreUnitStats(o.request.GetPreCapEpWeights()), reforgeCaps, reforgeSoftCaps)

	variables, constraints := o.buildEnchantModel(candidates, weights)
	selected, score, err := o.solveModel(weights, reforgeCaps, reforgeSoftCaps, variables, constraints, optimizerTimeout.Seconds())
	if err != nil {
		return nil, 0, err
	}
	return o.applyEnchantSolution(candidates, selected), score, nil
}

func computeReforgeStats(request *proto.ComputeStatsRequest) *proto.ComputeStatsResult {
	request.SkipRotation = true
	return core.ComputeStats(request)
}

func computeReforgeStatsAndDeps(request *proto.ComputeStatsRequest) (*proto.ComputeStatsResult, *stats.StatDependencyManager, core.SheetAvoidance) {
	request.SkipRotation = true
	return core.ComputeStatsAndDeps(request)
}

func optimizeError(message string) *proto.ReforgeOptimizeResult {
	return &proto.ReforgeOptimizeResult{
		Error: &proto.ErrorOutcome{Message: message},
	}
}

func optimizeAborted() *proto.ReforgeOptimizeResult {
	return &proto.ReforgeOptimizeResult{
		Error: &proto.ErrorOutcome{
			Type:    proto.ErrorOutcomeType_ErrorOutcomeAborted,
			Message: "Reforge optimization aborted.",
		},
	}
}

const (
	optimizerTimeout = 30 * time.Second
)
