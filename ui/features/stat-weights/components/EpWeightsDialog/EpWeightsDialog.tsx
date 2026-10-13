import { Tabs } from '@base-ui/react/tabs';
import { showsEpRatios } from '@features/results/model/sim_results';
import { StatFilterPopover, useStatFilter } from '@features/stat-filters';
import { ErrorOutcomeType, type ProgressMetrics, type StatWeightsResult } from '@generated/proto/api';
import { Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { useSimHost, useSpecConfig } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { useSimStore } from '@sim/hooks/useSimStore';
import { useStatWeights } from '@sim/hooks/useStatWeights';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { Stats } from '@sim/proto/stats';
import type { StatWeightActionSettings } from '@sim/settings/stat_weight_settings';
import { subscribePlayerField, subscribeStatWeightsChange } from '@sim/state/subscriptions';
import { Button } from '@ui-kit/Button';
import { ContentBlock } from '@ui-kit/ContentBlock';
import { Dialog } from '@ui-kit/Dialog';
import { Icon } from '@ui-kit/Icon';
import { ProgressTrackerDialog, type ProgressTrackerHandle, type ProgressTrackerProgress, type ProgressTrackerState } from '@ui-kit/ProgressTrackerDialog';
import { TabNav, TabPanel, TabPanels } from '@ui-kit/TabNav';
import { toastManager } from '@ui-kit/Toast';
import { Tooltip } from '@ui-kit/Tooltip';
import { useCallback, useMemo, useRef, useState } from 'react';

import { trackEvent } from '../../../../tracking/utils';
import { useStatSweepControls } from '../../hooks/useStatSweepControls';
import {
	calculateEp,
	combineScaledEpValues,
	combineScaledWeights,
	emptyStatWeightsResult,
	epWeightsWithoutExcluded,
	precisionRunOptions,
} from '../../model/ep_math';
import { epFilterScope } from '../../model/ep_unit_stats';
import type { StatSweepProgress } from '../../model/stat_sweep';
import { statsTableColumns } from '../../model/stats_table';
import { EpPrecisionOptions } from '../EpPrecisionOptions/EpPrecisionOptions';
import { SavedEpWeights } from '../SavedEpWeights';
import { StatSweep } from '../StatSweep/StatSweep';
import { HandleStatSweepTimeRemaining, type StatSweepTimeRemainingHandle } from '../StatSweep/StatSweepTimeRemaining';
import { EpReferenceOptions } from './EpReferenceOptions';
import { EpWeightsOptions } from './EpWeightsOptions';
import { EpWeightsTable } from './EpWeightsTable';
import { StatsType } from './types';
import { buildEpColumns, EP_TOOLTIP_ID } from './utils';

const WEIGHTS_TAB = 'ep-weights-pane';
const SWEEP_TAB = 'stat-sweep-pane';

const trackerProgress = (metrics: ProgressMetrics): ProgressTrackerProgress => ({
	title: `${metrics.completedSims} / ${metrics.totalSims} ${i18n.t('sidebar.buttons.stat_weights.modal.progress.simulations_complete')}`,
	current: metrics.completedIterations,
	total: metrics.totalIterations,
});

// The bar spans the whole sweep, every stat's steps, read out in iterations at the latest step's count; before any step reports one it counts steps.
const sweepTrackerProgress = ({ stat, stats, step, steps, completed, totalSteps, iterationsPerStep }: StatSweepProgress): ProgressTrackerProgress => {
	const scale = iterationsPerStep || 1;
	const title =
		stats > 1
			? i18n.t('sidebar.buttons.stat_weights.sweep.progress_stats', { stat, stats, current: step, total: steps })
			: i18n.t('sidebar.buttons.stat_weights.sweep.progress', { current: step, total: steps });
	return { title, current: completed * scale, total: totalSteps * scale };
};

export interface EpWeightsDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	settings: StatWeightActionSettings;
}

export const EpWeightsDialog = ({ open, onOpenChange, settings }: EpWeightsDialogProps) => {
	const host = useSimHost();
	const { player, sim } = host;
	const individualConfig = useSpecConfig();

	const epStats = individualConfig.epStats;
	const epReferenceStat = individualConfig.epReferenceStat;
	const epPseudoStats = useMemo(() => individualConfig.epPseudoStats || [], [individualConfig]);
	const epStatSet = useMemo(() => ({ epStats, epPseudoStats }), [epStats, epPseudoStats]);
	const scope = useMemo(() => epFilterScope(epStatSet), [epStatSet]);
	const statFilter = useStatFilter('ep', scope);

	const [tab, setTab] = useState(WEIGHTS_TAB);
	const [statsType, setStatsType] = useState<StatsType>(StatsType.Ep);
	const [iterations, setIterations] = useState(0);
	const [simResult, setSimResult] = useState<StatWeightsResult | null>(null);
	const [progress, setProgress] = useState<ProgressTrackerState>({ stage: 'initializing' });
	const progressRef = useRef<ProgressTrackerHandle>(null);
	const {
		isRunning,
		isAborting,
		abort,
		start: startStatWeights,
	} = useStatWeights({ onProgress: metrics => progressRef.current?.setProgress(trackerProgress(metrics)) });

	const simDisplayMetrics = useDisplayMetrics(sim);
	// A gear planner never computes weights, so none of the computed columns apply: only the
	// editable Current EP column is left (and the calculate button below is dropped).
	const editOnly = !!host.simDisabled;
	const displayMetrics = useMemo(() => (editOnly ? { damage: false, healing: false, threat: false } : simDisplayMetrics), [editOnly, simDisplayMetrics]);
	const { threat: showThreatMetrics } = displayMetrics;
	const showEpRatios = !editOnly && showsEpRatios(displayMetrics);
	const refStats = useStoreSubscribe(subscribePlayerField(player, 'epRefStat'), () => ({
		dps: player.getRefStat('dpsRefStat'),
		heal: player.getRefStat('healRefStat'),
		tank: player.getRefStat('tankRefStat'),
	}));
	const dpsReferenceStat = refStats.dps ?? epReferenceStat;
	const hpsReferenceStat = refStats.heal ?? epReferenceStat;
	const epRatios = useStoreSubscribe(subscribePlayerField(player, 'epRatios'), () => player.getEpRatios());
	const epWeights = usePlayerStore('epWeights');
	const precision = useStoreSubscribe(subscribeStatWeightsChange(settings), () => settings.precision);
	const defaultIterations = useSimStore('iterations');
	const sweepProgressRef = useRef<ProgressTrackerHandle>(null);
	const sweepRemainingRef = useRef<StatSweepTimeRemainingHandle>(null);
	const sweepControls = useStatSweepControls(settings, {
		onProgress: progress => {
			sweepProgressRef.current?.setProgress(sweepTrackerProgress(progress));
			sweepRemainingRef.current?.setSeconds(progress.secondsRemaining);
		},
	});
	const { isRunning: isSweeping, cancel: cancelSweep } = sweepControls.sweep;

	// `calculateEp` writes `epValues` from `weights`, which normalisation leaves alone.
	const result = useMemo(() => (simResult ? calculateEp(simResult, refStats) : null), [simResult, refStats]);
	const prevSimResult = useMemo(() => result ?? emptyStatWeightsResult(), [result]);

	const applyWeights = useCallback(
		(newWeights: Stats) => player.setEpWeights(epWeightsWithoutExcluded(newWeights, player.getEpWeights(), settings)),
		[player, settings],
	);

	const columns = useMemo(
		() =>
			buildEpColumns(
				statsTableColumns({
					getPrevSimResult: () => prevSimResult,
					getDefaultEpWeights: () => individualConfig.defaults.epWeights.toProto(),
					getDpsEpRefStat: () => dpsReferenceStat,
					getHealEpRefStat: () => hpsReferenceStat,
					getTankEpRefStat: () => refStats.tank ?? Stat.StatArmor,
				}),
				weights => applyWeights(Stats.fromProto(weights)),
			),
		[prevSimResult, individualConfig, dpsReferenceStat, hpsReferenceStat, refStats, applyWeights],
	);
	const visibleColumns = useMemo(() => columns.filter(column => !column.metric || displayMetrics[column.metric]), [columns, displayMetrics]);

	const onComputeEp = useCallback(() => {
		const combine = statsType === StatsType.Ep ? combineScaledEpValues : combineScaledWeights;
		applyWeights(combine(prevSimResult, player.getEpRatios()));
	}, [statsType, prevSimResult, player, applyWeights]);

	const handleOpenChange = useCallback(
		(next: boolean) => {
			onOpenChange(next);
			if (!next) {
				abort().catch(console.error);
				cancelSweep();
			}
		},
		[onOpenChange, abort, cancelSweep],
	);

	const onCancel = useCallback(() => {
		if (isAborting) return;
		abort().catch(error => {
			console.error('Error on stat weight abort!');
			console.error(error);
		});
	}, [abort, isAborting]);

	const onCalculate = useCallback(async () => {
		trackEvent({ action: 'sim', category: 'stat_weights', label: 'calculate' });
		if (isRunning) return;
		setProgress({ stage: 'initializing' });

		let result: StatWeightsResult | null = null;
		let runIterations = 0;
		try {
			runIterations = precision.iterations || sim.getIterations();
			setProgress({ stage: 'running' });
			result = await startStatWeights({
				epStats: epStats.filter(stat => !settings.isStatExcludedFromCalc(stat)),
				epPseudoStats: epPseudoStats.filter(pseudoStat => !settings.isPseudoStatExcludedFromCalc(pseudoStat)),
				// A target is measured against these stats' weights, so they have to be the ones the shown ± divides by.
				epReferenceStat: dpsReferenceStat,
				options: { ...precisionRunOptions(precision, displayMetrics), hpsReferenceStat },
			});
			if (result.error) {
				if (result.error.type === ErrorOutcomeType.ErrorOutcomeAborted) toastManager.add({ variant: 'info', body: 'Statweight sim cancelled.' });
				result = null;
			}
		} catch (error: any) {
			console.error(error);
			toastManager.add({
				variant: 'error',
				body: error?.message || 'Something went wrong calculating your stat weights. Reload the page and try again.',
			});
			result = null;
		}

		if (!result) return;
		setIterations(runIterations);
		setSimResult(result);
	}, [sim, settings, precision, displayMetrics, dpsReferenceStat, hpsReferenceStat, epStats, epPseudoStats, isRunning, startStatWeights]);

	return (
		<Dialog
			open={open}
			onOpenChange={handleOpenChange}
			testId="ep-weights-menu"
			size="xl"
			scrollContents
			title={i18n.t('sidebar.buttons.stat_weights.modal.title')}
			footer={
				// A gear planner keeps the editor and drops the calculation: there is no sim to run it on.
				host.simDisabled ? undefined : tab === SWEEP_TAB ? (
					<Button variant="primary" disabled={isSweeping || !sweepControls.selectedStats.length} onClick={sweepControls.onRun}>
						{i18n.t('sidebar.buttons.stat_weights.sweep.run')}
					</Button>
				) : (
					<Button data-testid="calc-weights" disabled={isRunning || isSweeping} onClick={() => void onCalculate()}>
						<Icon name="calculator" />
						{i18n.t('sidebar.buttons.stat_weights.modal.calculate')}
					</Button>
				)
			}>
			<Tabs.Root className="ui-ep-weights-tabs" value={tab} onValueChange={next => setTab(String(next))}>
				{!editOnly && (
					<TabNav
						tabs={[
							{ id: WEIGHTS_TAB, label: i18n.t('sidebar.buttons.stat_weights.modal.tabs.weights') },
							{ id: SWEEP_TAB, label: i18n.t('sidebar.buttons.stat_weights.modal.tabs.sweep') },
						]}
					/>
				)}
				{/* Both panels stay laid out, so the dialog keeps the taller one's height on either tab. */}
				<TabPanels className="ui-ep-weights-panels">
					<TabPanel value={WEIGHTS_TAB} testId={WEIGHTS_TAB} className="ui-ep-weights-pane" preserveHeight>
						<div className="ui-ep-weights-layout">
							<div className="ui-ep-weights-main">
								<p>
									{i18n.t('sidebar.buttons.stat_weights.modal.current_ep_description')}
									{!editOnly && (
										<>
											<br />
											{i18n.t('sidebar.buttons.stat_weights.modal.copy_icon_description')}
										</>
									)}
								</p>
								<EpWeightsTable
									columns={visibleColumns}
									stats={statFilter.shown}
									statsType={statsType}
									result={result}
									iterations={iterations || 1}
									epRatios={epRatios}
									epWeights={epWeights}
									settings={settings}
									player={player}
									epStatSet={epStatSet}
									epReferenceStat={epReferenceStat}
									onComputeEp={onComputeEp}
									showThreatMetrics={showThreatMetrics}
									showEpRatios={showEpRatios}
									displayMetrics={displayMetrics}
									showUpdateColumn={!editOnly}
									statAction={<StatFilterPopover {...statFilter} ariaLabel={i18n.t('stat_filters.triggers.ep')} testId="ep-stat-filter" />}
								/>
							</div>
							<div className="ui-ep-weights-sidebar">
								<ContentBlock config={{ header: { title: i18n.t('sidebar.buttons.stat_weights.modal.settings') } }}>
									<div className="flex flex-col gap-4">
										<div className="ui-ep-weights-settings-grid">
											<EpWeightsOptions statsType={statsType} onStatsTypeChange={setStatsType} />
										</div>
										{!editOnly && (
											<>
												<EpReferenceOptions epStats={epStats} epReferenceStat={epReferenceStat} displayMetrics={displayMetrics} />
												<EpPrecisionOptions
													value={precision}
													onChange={patch => settings.setPrecision(patch)}
													defaultIterations={defaultIterations}
													disabled={isRunning}
												/>
											</>
										)}
									</div>
								</ContentBlock>
								<SavedEpWeights />
							</div>
						</div>
					</TabPanel>
					{!editOnly && (
						<TabPanel value={SWEEP_TAB} testId={SWEEP_TAB} className="ui-ep-weights-pane" preserveHeight>
							<StatSweep settings={settings} controls={sweepControls} />
						</TabPanel>
					)}
				</TabPanels>
			</Tabs.Root>
			<Tooltip id={EP_TOOLTIP_ID} />
			{isRunning && (
				<ProgressTrackerDialog
					ref={progressRef}
					open
					testId="progress-tracker-dialog"
					title={i18n.t('sidebar.buttons.stat_weights.modal.title')}
					state={progress}
					hasProgressBar
					onCancel={onCancel}
				/>
			)}
			{isSweeping && (
				<ProgressTrackerDialog
					ref={sweepProgressRef}
					open
					testId="stat-sweep-progress-dialog"
					title={i18n.t('sidebar.buttons.stat_weights.modal.tabs.sweep')}
					state={{ stage: 'running', message: <HandleStatSweepTimeRemaining ref={sweepRemainingRef} /> }}
					hasProgressBar
					onCancel={cancelSweep}
				/>
			)}
		</Dialog>
	);
};
