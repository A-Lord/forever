import i18n from '@i18n/config';
import { usePlayer, useSim, useSpecConfig } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { useSimStore } from '@sim/hooks/useSimStore';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { Stats, UnitStat } from '@sim/proto/stats';
import type { StatSweepSettings } from '@sim/settings/stat_sweep_settings';
import { subscribeStatSweepChange } from '@sim/state/subscriptions';
import { Button } from '@ui-kit/Button';
import { Dialog } from '@ui-kit/Dialog';
import { ProgressTrackerDialog, type ProgressTrackerHandle, type ProgressTrackerProgress } from '@ui-kit/ProgressTrackerDialog';
import { downloadString } from '@ui-kit/utils/dom';
import { useMemo, useRef } from 'react';

import { useStatSweep } from '../../hooks/useStatSweep';
import { offeredSweepMetrics, shownSweepMetric, statSweepCsv, type StatSweepProgress } from '../../model/stat_sweep';
import { StatSweepChart } from './StatSweepChart';
import { StatSweepHistory } from './StatSweepHistory';
import { StatSweepPanel } from './StatSweepPanel';
import { StatSweepPrecisionOptions } from './StatSweepPrecisionOptions';
import { HandleStatSweepStatList, type StatSweepStatListHandle } from './StatSweepStatList';
import { HandleStatSweepTimeRemaining, type StatSweepTimeRemainingHandle } from './StatSweepTimeRemaining';

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`sidebar.buttons.stat_sweep.${key}`, options);

const CSV_FILE_NAME = 'wowsims-stat-sweep.csv';

// The bar spans every stat's steps, read out in iterations at the running step's expected count; before a step has one it counts steps.
const trackerProgress = ({ step, steps, completed, totalSteps, iterationsPerStep }: StatSweepProgress): ProgressTrackerProgress => {
	const scale = iterationsPerStep || 1;
	return { title: t('progress', { current: step, total: steps }), current: completed * scale, total: totalSteps * scale };
};

export interface StatSweepDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	settings: StatSweepSettings;
}

export const StatSweepDialog = ({ open, onOpenChange, settings }: StatSweepDialogProps) => {
	const player = usePlayer();
	const { epStats, epPseudoStats } = useSpecConfig();
	const progressRef = useRef<ProgressTrackerHandle>(null);
	const remainingRef = useRef<StatSweepTimeRemainingHandle>(null);
	const statListRef = useRef<StatSweepStatListHandle>(null);
	const sweep = useStatSweep({
		onProgress: progress => {
			progressRef.current?.setProgress(trackerProgress(progress));
			remainingRef.current?.setSeconds(progress.secondsRemaining);
			statListRef.current?.setRunning(progress.stat - 1);
		},
	});
	const defaultIterations = useSimStore('iterations');
	const currentStats = usePlayerStore('currentStats');
	const displayMetrics = useDisplayMetrics(useSim());
	const metrics = useMemo(() => offeredSweepMetrics(displayMetrics), [displayMetrics]);

	const unitStats = useMemo(
		() => [...epStats.map(stat => UnitStat.fromStat(stat)), ...(epPseudoStats ?? []).map(pseudoStat => UnitStat.fromPseudoStat(pseudoStat))],
		[epStats, epPseudoStats],
	);
	const statOptions = useMemo(() => unitStats.map(stat => ({ value: stat.getKey(), label: stat.getFullName(player.getClass()) })), [unitStats, player]);
	const { stats: storedStats, range, step } = useStoreSubscribe(subscribeStatSweepChange(settings), () => settings.setup);
	const precision = useStoreSubscribe(subscribeStatSweepChange(settings), () => settings.precision);
	const metric = shownSweepMetric(
		useStoreSubscribe(subscribeStatSweepChange(settings), () => settings.metric),
		metrics,
	);
	const showSpread = useStoreSubscribe(subscribeStatSweepChange(settings), () => settings.showSpread);
	// The stored stats the spec offers, in the order offered; with none stored, or none of them offered, the first one offered. A cleared list stays empty.
	const selectedStats = useMemo(() => {
		if (storedStats?.length === 0) return [];
		const offered = unitStats.filter(stat => storedStats?.some(stored => stored.equals(stat)));
		return offered.length ? offered : unitStats.slice(0, 1);
	}, [unitStats, storedStats]);
	const selectedKeys = useMemo(() => selectedStats.map(stat => stat.getKey()), [selectedStats]);
	const selectedLabels = useMemo(() => selectedKeys.map(key => statOptions.find(option => option.value === key)?.label ?? ''), [selectedKeys, statOptions]);

	const onRun = () => {
		if (!selectedStats.length) return;
		const finalStats = currentStats?.finalStats ? Stats.fromProto(currentStats.finalStats) : null;
		sweep
			.run({
				stats: selectedStats.map(unitStat => ({
					unitStat,
					label: t('run_label', { stat: statOptions.find(option => option.value === unitStat.getKey())?.label ?? '', range }),
					baseStatValue: finalStats?.getUnitStat(unitStat) ?? 0,
				})),
				range,
				step,
				metric,
				precision,
			})
			.catch(console.error);
	};

	const onExportCsv = sweep.series.some(run => run.points.length) ? () => downloadString(statSweepCsv(sweep.series), CSV_FILE_NAME, 'text/csv') : undefined;

	return (
		<Dialog
			open={open}
			onOpenChange={next => {
				onOpenChange(next);
				if (!next) sweep.cancel();
			}}
			testId="stat-sweep-dialog"
			size="xl"
			maxWidthClassName="max-w-modal-inset-w w-modal-inset-w"
			className="h-modal-inset-h"
			scrollContents
			title={t('title')}
			footer={
				<Button variant="primary" disabled={sweep.isRunning || !selectedStats.length} onClick={onRun}>
					{t('run')}
				</Button>
			}>
			<StatSweepPanel
				statOptions={statOptions}
				stats={selectedKeys}
				onStatsChange={keys => settings.setSetup({ stats: unitStats.filter(stat => keys.includes(stat.getKey())) })}
				range={range}
				onRangeChange={range => settings.setSetup({ range })}
				step={step}
				onStepChange={step => settings.setSetup({ step })}
				fractional={selectedStats.some(stat => stat.isPseudoStat())}
				precision={
					<StatSweepPrecisionOptions
						value={precision}
						onChange={patch => settings.setPrecision(patch)}
						metric={metric}
						defaultIterations={defaultIterations}
						disabled={sweep.isRunning}
					/>
				}
				chart={
					<StatSweepChart
						series={sweep.series}
						metric={metric}
						metrics={metrics}
						onMetricChange={next => settings.setMetric(next)}
						showSpread={showSpread}
						onShowSpreadChange={next => settings.setShowSpread(next)}
					/>
				}
				history={<StatSweepHistory runs={sweep.historyEntries} onOverlayChange={sweep.setOverlaid} onDelete={sweep.remove} onExportCsv={onExportCsv} />}
				running={sweep.isRunning}
			/>
			{sweep.isRunning && (
				<ProgressTrackerDialog
					ref={progressRef}
					open
					testId="stat-sweep-progress-dialog"
					title={t('title')}
					state={{
						stage: 'running',
						message: (
							<>
								<HandleStatSweepStatList ref={statListRef} labels={selectedLabels} />
								<HandleStatSweepTimeRemaining ref={remainingRef} />
							</>
						),
					}}
					hasProgressBar
					onCancel={sweep.cancel}
				/>
			)}
		</Dialog>
	);
};
