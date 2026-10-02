package bulk

import (
	"github.com/wowsims/forever/sim/core/proto"
)

// bulkSimHighStageFloor is the high stage's iteration floor: the requested count, or the stage's own minimum when the request leaves it unset.
func bulkSimHighStageFloor(requested int32) int32 {
	return getBulkSimStageMinIterations(requested, BulkSimStageConfigs[len(BulkSimStageConfigs)-1])
}

func shouldUseSingleStageBulkSim(highStageIterations int32, candidateCount int) bool {
	if candidateCount < BulkSimMinCombinations {
		return true
	}

	highStageIterations = bulkSimHighStageFloor(highStageIterations)
	fullRunIterations := int64(highStageIterations) * int64(candidateCount)
	estimatedMultistageIterationsUpperBound := getBulkSimOptimisationIterationsUpperBound(highStageIterations, candidateCount)
	return estimatedMultistageIterationsUpperBound >= fullRunIterations
}

func getBulkSimOptimisationIterationsUpperBound(highStageIterations int32, candidateCount int) int64 {
	remainingCandidates := candidateCount
	var iterations int64

	for _, stageConfig := range BulkSimStageConfigs {
		if stageConfig.Stage == proto.BulkSimStage_BulkSimStageHigh {
			break
		}
		if !shouldRunBulkSimStage(stageConfig, remainingCandidates) {
			continue
		}

		stageIterations := getBulkSimStageMinIterations(highStageIterations, stageConfig)
		iterations += int64(stageIterations) * int64(remainingCandidates+1)
		remainingCandidates = min(remainingCandidates, getBulkSimStageMaxSurvivors(stageConfig, remainingCandidates))
	}

	return iterations + int64(highStageIterations)*int64(remainingCandidates+1)
}

// estimateBulkSimIterations is what a batch of candidateCount gear sets sims at each stage's minimum, the baseline included,
// with every culling stage keeping its most survivors. The error target can raise a stage above its minimum, and the adaptive
// passes and the finalist stage add iterations on top, so a run usually sims more.
func estimateBulkSimIterations(candidateCount int) int64 {
	highStageIterations := BulkSimStageConfigs[len(BulkSimStageConfigs)-1].MinIterations
	if shouldUseSingleStageBulkSim(highStageIterations, candidateCount) {
		return int64(highStageIterations) * int64(candidateCount+1)
	}

	return getBulkSimOptimisationIterationsUpperBound(highStageIterations, candidateCount)
}
