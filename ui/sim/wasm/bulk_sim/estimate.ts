import { BulkSimRequest, BulkSimStage } from '@generated/proto/api';

import { BULK_SIM_MIN_COMBINATIONS } from './constants_auto_gen';
import { bulkSimStageConfigs, getBulkSimStageMaxSurvivors, getBulkSimStageMinIterations, shouldRunBulkSimStage } from './stage';

export const shouldUseSingleStageBulkSim = (request: BulkSimRequest, candidateCount: number): boolean => {
	if (candidateCount < BULK_SIM_MIN_COMBINATIONS) {
		return true;
	}

	const highStageConfig = bulkSimStageConfigs[bulkSimStageConfigs.length - 1]!;
	const highStageIterations = getBulkSimStageMinIterations(request, highStageConfig);
	let remainingCandidates = candidateCount;
	let estimatedOptimisationIterationsUpperBound = 0;

	for (const config of bulkSimStageConfigs) {
		if (config.stage === BulkSimStage.BulkSimStageHigh) {
			break;
		}
		if (!shouldRunBulkSimStage(config, remainingCandidates)) {
			continue;
		}

		estimatedOptimisationIterationsUpperBound += getBulkSimStageMinIterations(request, config) * (remainingCandidates + 1);
		remainingCandidates = Math.min(remainingCandidates, getBulkSimStageMaxSurvivors(config, remainingCandidates) ?? remainingCandidates);
	}

	estimatedOptimisationIterationsUpperBound += highStageIterations * (remainingCandidates + 1);
	return estimatedOptimisationIterationsUpperBound >= highStageIterations * candidateCount;
};
