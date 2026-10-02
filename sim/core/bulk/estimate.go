package bulk

import (
	"github.com/wowsims/forever/sim/core/proto"
)

func bulkSimHighStageFloor() int32 {
	return BulkSimStageConfigs[len(BulkSimStageConfigs)-1].MinIterations
}

func shouldUseSingleStageBulkSim(candidateCount int) bool {
	if candidateCount < BulkSimMinCombinations {
		return true
	}

	fullRunIterations := int64(bulkSimHighStageFloor()) * int64(candidateCount)
	estimatedMultistageIterationsUpperBound := getBulkSimOptimisationIterationsUpperBound(candidateCount)
	return estimatedMultistageIterationsUpperBound >= fullRunIterations
}

func getBulkSimOptimisationIterationsUpperBound(candidateCount int) int64 {
	remainingCandidates := candidateCount
	var iterations int64

	for _, stageConfig := range BulkSimStageConfigs {
		if stageConfig.Stage == proto.BulkSimStage_BulkSimStageHigh {
			break
		}
		if !shouldRunBulkSimStage(stageConfig, remainingCandidates) {
			continue
		}

		iterations += int64(stageConfig.MinIterations) * int64(remainingCandidates+1)
		remainingCandidates = min(remainingCandidates, getBulkSimStageMaxSurvivors(stageConfig, remainingCandidates))
	}

	return iterations + int64(bulkSimHighStageFloor())*int64(remainingCandidates+1)
}

// estimateBulkSimIterations is what a batch of candidateCount gear sets sims at each stage's minimum, the baseline included,
// with every culling stage keeping its most survivors. The error target can raise a stage above its minimum, and the adaptive
// passes and the finalist stage add iterations on top, so a run usually sims more.
func estimateBulkSimIterations(candidateCount int) int64 {
	if shouldUseSingleStageBulkSim(candidateCount) {
		return int64(bulkSimHighStageFloor()) * int64(candidateCount+1)
	}

	return getBulkSimOptimisationIterationsUpperBound(candidateCount)
}
