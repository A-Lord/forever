import type { ReferenceDiffs } from '@features/results/model/reference_diffs';
import { referenceDiffs } from '@features/results/model/reference_diffs';
import type { ResultMetric } from '@features/results/model/topline_metrics';
import { toplineResultMetrics } from '@features/results/model/topline_metrics';

import { makeSimResult } from './results';

export interface ToplineFixture {
	metrics: Array<ResultMetric>;
	/** The same run against one 5% weaker, so every compared metric carries a delta. */
	referenceDiffs: ReferenceDiffs;
}

/** The topline metrics of `makeSimResult()`'s run, for a story's `loaders`. */
export const loadTopline = async (): Promise<ToplineFixture> => {
	const current = await makeSimResult();
	const reference = await makeSimResult({ dpsScale: 0.95 });
	return { metrics: toplineResultMetrics(current), referenceDiffs: referenceDiffs(current, reference) };
};
