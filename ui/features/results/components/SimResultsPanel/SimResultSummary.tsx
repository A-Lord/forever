import { useSim } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';

import { useResultFreshness } from '../../hooks/useResultFreshness';
import { referenceDiffs } from '../../model/reference_diffs';
import type { SimResultsManager } from '../../model/results_manager';
import { toplineResultMetrics } from '../../model/topline_metrics';
import { SimResultSummaryView } from './SimResultSummaryView';

export interface SimResultSummaryProps {
	results: SimResultsManager;
	onOpenBreakdown?: () => void;
}

/** Renders nothing until a run completes, so the panel's result zone is empty at load. */
export const SimResultSummary = ({ results, onOpenBreakdown }: SimResultSummaryProps) => {
	const displayMetrics = useDisplayMetrics(useSim());
	const { current, reference, swapped } = useStoreSubscribe(results.subscribe, results.getData);
	const freshness = useResultFreshness();

	if (!current) return null;

	return (
		<SimResultSummaryView
			metrics={toplineResultMetrics(current.simResult, undefined, { displayMetrics })}
			referenceDiffs={reference ? referenceDiffs(current.simResult, reference.simResult) : undefined}
			freshness={freshness}
			swapped={swapped}
			onOpenBreakdown={onOpenBreakdown}
			onSetReference={() => results.setReference()}
			onSwapReference={() => results.swapReference()}
			onClearReference={() => results.clearReference()}
		/>
	);
};
