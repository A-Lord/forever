import { BulkSimStage } from '@generated/proto/api';

import { BULK_SIM_MIN_COMBINATIONS } from './constants_auto_gen';
import { bulkSimStageConfigs, getBulkSimStageMaxSurvivors, getBulkSimStageMinIterations, shouldRunBulkSimStage } from './stage';

export const bulkSimHighStageFloor = (): number => getBulkSimStageMinIterations(bulkSimStageConfigs[bulkSimStageConfigs.length - 1]!);

export const shouldUseSingleStageBulkSim = (candidateCount: number): boolean => {
	if (candidateCount < BULK_SIM_MIN_COMBINATIONS) {
		return true;
	}

	const highStageIterations = bulkSimHighStageFloor();
	let remainingCandidates = candidateCount;
	let estimatedOptimisationIterationsUpperBound = 0;

	for (const config of bulkSimStageConfigs) {
		if (config.stage === BulkSimStage.BulkSimStageHigh) {
			break;
		}
		if (!shouldRunBulkSimStage(config, remainingCandidates)) {
			continue;
		}

		estimatedOptimisationIterationsUpperBound += getBulkSimStageMinIterations(config) * (remainingCandidates + 1);
		remainingCandidates = Math.min(remainingCandidates, getBulkSimStageMaxSurvivors(config, remainingCandidates) ?? remainingCandidates);
	}

	estimatedOptimisationIterationsUpperBound += highStageIterations * (remainingCandidates + 1);
	return estimatedOptimisationIterationsUpperBound >= highStageIterations * candidateCount;
};
