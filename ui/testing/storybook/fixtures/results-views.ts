import { ALL_UNITS, simResultFilter } from '@features/results/components/ResultsFilter/utils';
import type { ReferenceDiffs } from '@features/results/model/reference_diffs';
import { referenceDiffs } from '@features/results/model/reference_diffs';
import { ResultChannel } from '@features/results/model/result_channel';
import type { ResultMetric } from '@features/results/model/topline_metrics';
import { toplineResultMetrics } from '@features/results/model/topline_metrics';
import type { SimResult } from '@sim/proto/sim_result';

import type { StubHostParts } from './models';
import { makeSimResult } from './results';

export interface ResultFixture {
	result: SimResult;
}

/** `makeSimResult()`'s run, for a story's `loaders`. */
export const loadSimResult = async (): Promise<ResultFixture> => ({ result: await makeSimResult() });

export interface DisplayFlags {
	damage?: boolean;
	threat?: boolean;
	healing?: boolean;
}

/**
 * `withSimHost` parts for a container that reads the finished run (`useSimResult`) and the metric
 * display flags (`useDisplayMetrics`). The run is emitted unfiltered, as the results tab first shows it.
 */
export const resultsHostParts = (result: SimResult | null, { damage = true, threat = true, healing = false }: DisplayFlags = {}): StubHostParts => {
	const resultChannel = new ResultChannel();
	resultChannel.emit(result && { result, filter: simResultFilter(ALL_UNITS) });
	return {
		sim: { getShowDamageMetrics: () => damage, getShowThreatMetrics: () => threat, getShowHealingMetrics: () => healing },
		resultChannel,
	};
};

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
