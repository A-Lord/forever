import { useSim, useSimHost } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { textClassName } from '@sim/proto/utils';

import { referenceDiffs } from '../../model/reference_diffs';
import type { SimResultsManager } from '../../model/results_manager';
import { toplineResultMetrics } from '../../model/topline_metrics';
import { SimResultSummaryView } from './SimResultSummaryView';

export interface SimResultSummaryProps {
	results: SimResultsManager;
}

/** Renders nothing until a run completes, so the panel's result zone is empty at load. */
export const SimResultSummary = ({ results }: SimResultSummaryProps) => {
	const { cssScheme } = useSimHost().config;
	const displayMetrics = useDisplayMetrics(useSim());
	const { current, reference } = useStoreSubscribe(results.subscribe, results.getData);

	if (!current) return null;

	return (
		<SimResultSummaryView
			metrics={toplineResultMetrics(current.simResult, undefined, { displayMetrics })}
			referenceDiffs={reference ? referenceDiffs(current.simResult, reference.simResult) : undefined}
			pinClassName={textClassName(cssScheme)}
			onSetReference={() => results.setReference()}
			onSwapReference={() => results.swapReference()}
			onClearReference={() => results.clearReference()}
		/>
	);
};
