import type { SimResultsManager } from '@features/results/model/results_manager';
import type { WarningsRegistry } from '@features/results/model/warnings';
import { useSimHost } from '@sim/context/SimHostContext';
import { useSimReady } from '@sim/hooks/useSimReady';
import { Spinner } from '@ui-kit/Spinner';
import { useSyncExternalStore } from 'react';

import { AbortButton } from './AbortButton';
import { GearPlannerNotice } from './GearPlannerNotice';
import { ResultsPanelStage, type ResultsPanelStore } from './results_panel_store';
import { SimProgress } from './SimProgress';
import { SimResultSummary } from './SimResultSummary';
import { SimWarnings } from './SimWarnings';
import { UnlaunchedNotice } from './UnlaunchedNotice';

export interface SimResultsPanelProps {
	panel: ResultsPanelStore;
	warnings: WarningsRegistry;
	results: SimResultsManager | null;
}

export const SimResultsPanel = ({ panel, warnings, results }: SimResultsPanelProps) => {
	const host = useSimHost();
	const ready = useSimReady();
	const stage = useSyncExternalStore(panel.subscribe, panel.getStage);
	const abortHandler = useSyncExternalStore(panel.subscribe, panel.getAbortHandler);
	const buttonsVisible = useSyncExternalStore(panel.subscribe, panel.getButtonsVisible);

	return (
		<div className="mx-auto flex w-fit flex-col" data-testid="results-viewer">
			{(stage === ResultsPanelStage.Pending || stage === ResultsPanelStage.Running) &&
				(stage === ResultsPanelStage.Running ? <SimProgress panel={panel} /> : <Spinner size="lg" className="m-auto" />)}
			{stage === ResultsPanelStage.Result && results && <SimResultSummary results={results} />}
			{/* `stopping` is its own state, which a remount on `buttonsVisible` would drop — so the button
				stays mounted and only `hidden` moves with visibility; `abortHandler` is what removes it. */}
			{abortHandler && <AbortButton onAbort={abortHandler} hidden={!buttonsVisible} testId="button-zone" />}
			<SimWarnings warnings={warnings} ready={ready} />
			{host.disabled && <UnlaunchedNotice />}
			{host.simDisabled && !host.disabled && <GearPlannerNotice />}
		</div>
	);
};
