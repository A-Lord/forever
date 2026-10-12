import { Tabs } from '@base-ui/react/tabs';
import { showsEpRatios } from '@features/results/model/sim_results';
import { StatFilterPopover, useStatFilter } from '@features/stat-filters';
import { ErrorOutcomeType, type StatWeightsResult } from '@generated/proto/api';
import { Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { useSimHost, useSpecConfig } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { useSimRun } from '@sim/hooks/useSimRun';
import { useSimStore } from '@sim/hooks/useSimStore';
import { useStatWeights } from '@sim/hooks/useStatWeights';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { Stats } from '@sim/proto/stats';
import type { StatWeightActionSettings } from '@sim/settings/stat_weight_settings';
import { SimRunKind } from '@sim/state/sim_store';
import { subscribePlayerField, subscribeStatWeightsChange } from '@sim/state/subscriptions';
import { Button } from '@ui-kit/Button';
import { Dialog } from '@ui-kit/Dialog';
import { Icon } from '@ui-kit/Icon';
import { ProgressTrackerDialog, type ProgressTrackerHandle, type ProgressTrackerState } from '@ui-kit/ProgressTrackerDialog';
import { TabNav, TabPanel, TabPanels } from '@ui-kit/TabNav';
import { toastManager } from '@ui-kit/Toast';
import { Tooltip } from '@ui-kit/Tooltip';
import { useCallback, useMemo, useRef, useState } from 'react';

import { trackEvent } from '../../../../tracking/utils';
import {
	calculateEp,
	combineScaledEpValues,
	combineScaledWeights,
	emptyStatWeightsResult,
	epWeightsWithoutExcluded,
	precisionRunOptions,
} from '../../model/ep_math';
import { epFilterScope } from '../../model/ep_unit_stats';
import { statsTableColumns } from '../../model/stats_table';
import { EpPrecisionOptions } from '../EpPrecisionOptions/EpPrecisionOptions';
import { SavedEpWeights } from '../SavedEpWeights';
import { StatSweep } from '../StatSweep/StatSweep';
import { EpReferenceOptions } from './EpReferenceOptions';
import { EpWeightsOptions } from './EpWeightsOptions';
import { EpWeightsTable } from './EpWeightsTable';
import { StatsType } from './types';
import { buildEpColumns, EP_TOOLTIP_ID } from './utils';

const WEIGHTS_TAB = 'ep-weights-pane';
const SWEEP_TAB = 'stat-sweep-pane';

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
	} = useStatWeights({
		onProgress: metrics =>
			progressRef.current?.setProgress({
				title: `${metrics.completedSims} / ${metrics.totalSims} ${i18n.t('sidebar.buttons.stat_weights.modal.progress.simulations_complete')}`,
				current: metrics.completedIterations,
				total: metrics.totalIterations,
			}),
	});

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
	const epRatios = useStoreSubscribe(subscribePlayerField(player, 'epRatios'), () => player.getEpRatios());
	const epWeights = usePlayerStore('epWeights');
	const precision = useStoreSubscribe(subscribeStatWeightsChange(settings), () => settings.precision);
	const defaultIterations = useSimStore('iterations');
	const { isRunning: isSweeping, abort: abortSweep } = useSimRun(SimRunKind.StatSweep);

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
					getDpsEpRefStat: () => refStats.dps ?? epReferenceStat,
					getHealEpRefStat: () => refStats.heal ?? epReferenceStat,
					getTankEpRefStat: () => refStats.tank ?? Stat.StatArmor,
				}),
				weights => applyWeights(Stats.fromProto(weights)),
			),
		[prevSimResult, individualConfig, refStats, epReferenceStat, applyWeights],
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
				abortSweep().catch(console.error);
			}
		},
		[onOpenChange, abort, abortSweep],
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
				// The target is measured against this stat's weight, so it has to be the one the shown ± divides by.
				epReferenceStat: precision.precisionMode === 'target' ? (refStats.dps ?? epReferenceStat) : epReferenceStat,
				options: precisionRunOptions(precision, { displayMetrics, hpsReferenceStat: refStats.heal }),
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
	}, [sim, settings, precision, displayMetrics, refStats, epStats, epPseudoStats, epReferenceStat, isRunning, startStatWeights]);

	return (
		<Dialog
			open={open}
			onOpenChange={handleOpenChange}
			testId="ep-weights-menu"
			size={showThreatMetrics ? 'xl' : 'lg'}
			scrollContents
			title={i18n.t('sidebar.buttons.stat_weights.modal.title')}
			footer={
				// A gear planner keeps the editor and drops the calculation: there is no sim to run it on.
				// The sweep tab runs from its own button.
				host.simDisabled || tab === SWEEP_TAB ? undefined : (
					<Button data-testid="calc-weights" disabled={isRunning || isSweeping} onClick={() => void onCalculate()}>
						<Icon name="calculator" />
						{i18n.t('sidebar.buttons.stat_weights.modal.calculate')}
					</Button>
				)
			}>
			<Tabs.Root value={tab} onValueChange={next => setTab(String(next))}>
				{!editOnly && (
					<TabNav
						tabs={[
							{ id: WEIGHTS_TAB, label: i18n.t('sidebar.buttons.stat_weights.modal.tabs.weights') },
							{ id: SWEEP_TAB, label: i18n.t('sidebar.buttons.stat_weights.modal.tabs.sweep') },
						]}
					/>
				)}
				<TabPanels>
					<TabPanel value={WEIGHTS_TAB} testId={WEIGHTS_TAB}>
						<div className="flex flex-col gap-4 lg:flex-row lg:items-start">
							<div className="order-1 w-full lg:order-0">
								<EpWeightsOptions statsType={statsType} onStatsTypeChange={setStatsType}>
									<StatFilterPopover {...statFilter} ariaLabel={i18n.t('stat_filters.triggers.ep')} testId="ep-stat-filter" />
								</EpWeightsOptions>
								{!editOnly && (
									<>
										<EpReferenceOptions epStats={epStats} epReferenceStat={epReferenceStat} displayMetrics={displayMetrics} />
										<div className="mb-2">
											<EpPrecisionOptions
												value={precision}
												onChange={patch => settings.setPrecision(patch)}
												defaultIterations={defaultIterations}
												disabled={isRunning}
											/>
										</div>
									</>
								)}
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
								/>
							</div>
							<div className="ui-ep-weights-sidebar order-0 min-w-42 lg:sticky lg:top-0 lg:z-sticky lg:order-1">
								<SavedEpWeights />
							</div>
						</div>
					</TabPanel>
					{!editOnly && (
						<TabPanel value={SWEEP_TAB} testId={SWEEP_TAB}>
							<StatSweep settings={settings} />
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
		</Dialog>
	);
};
