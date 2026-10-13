import i18n from '@i18n/config';
import { usePlayer, useSpecConfig } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { useSimStore } from '@sim/hooks/useSimStore';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { Stats, UnitStat } from '@sim/proto/stats';
import type { DpsCurveSettings } from '@sim/settings/dps_curve_settings';
import { subscribeDpsCurveChange } from '@sim/state/subscriptions';
import { Button } from '@ui-kit/Button';
import { Dialog } from '@ui-kit/Dialog';
import { ProgressTrackerDialog, type ProgressTrackerHandle, type ProgressTrackerProgress } from '@ui-kit/ProgressTrackerDialog';
import { downloadString } from '@ui-kit/utils/dom';
import { useMemo, useRef } from 'react';

import { useDpsCurve } from '../../hooks/useDpsCurve';
import { dpsCurveCsv, type DpsCurveProgress } from '../../model/dps_curve';
import { DpsCurveChart } from './DpsCurveChart';
import { DpsCurveHistory } from './DpsCurveHistory';
import { DpsCurvePanel } from './DpsCurvePanel';
import { DpsCurvePrecisionOptions } from './DpsCurvePrecisionOptions';
import { type DpsCurveStatListHandle, HandleDpsCurveStatList } from './DpsCurveStatList';
import { type DpsCurveTimeRemainingHandle, HandleDpsCurveTimeRemaining } from './DpsCurveTimeRemaining';

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`sidebar.buttons.dps_curve.${key}`, options);

const CSV_FILE_NAME = 'wowsims-dps-curve.csv';

// The bar spans every stat's steps, read out in iterations at the running step's expected count; before a step has one it counts steps.
const trackerProgress = ({ step, steps, completed, totalSteps, iterationsPerStep }: DpsCurveProgress): ProgressTrackerProgress => {
	const scale = iterationsPerStep || 1;
	return { title: t('progress', { current: step, total: steps }), current: completed * scale, total: totalSteps * scale };
};

export interface DpsCurveDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	settings: DpsCurveSettings;
}

export const DpsCurveDialog = ({ open, onOpenChange, settings }: DpsCurveDialogProps) => {
	const player = usePlayer();
	const { epStats, epPseudoStats } = useSpecConfig();
	const progressRef = useRef<ProgressTrackerHandle>(null);
	const remainingRef = useRef<DpsCurveTimeRemainingHandle>(null);
	const statListRef = useRef<DpsCurveStatListHandle>(null);
	const curve = useDpsCurve({
		onProgress: progress => {
			progressRef.current?.setProgress(trackerProgress(progress));
			remainingRef.current?.setSeconds(progress.secondsRemaining);
			statListRef.current?.setRunning(progress.stat - 1);
		},
	});
	const defaultIterations = useSimStore('iterations');
	const currentStats = usePlayerStore('currentStats');

	const unitStats = useMemo(
		() => [...epStats.map(stat => UnitStat.fromStat(stat)), ...(epPseudoStats ?? []).map(pseudoStat => UnitStat.fromPseudoStat(pseudoStat))],
		[epStats, epPseudoStats],
	);
	const statOptions = useMemo(() => unitStats.map(stat => ({ value: stat.getKey(), label: stat.getFullName(player.getClass()) })), [unitStats, player]);
	const { stats: storedStats, range, step } = useStoreSubscribe(subscribeDpsCurveChange(settings), () => settings.setup);
	const precision = useStoreSubscribe(subscribeDpsCurveChange(settings), () => settings.precision);
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
		curve
			.run({
				stats: selectedStats.map(unitStat => ({
					unitStat,
					label: t('run_label', { stat: statOptions.find(option => option.value === unitStat.getKey())?.label ?? '', range }),
					baseStatValue: finalStats?.getUnitStat(unitStat) ?? 0,
				})),
				range,
				step,
				precision,
			})
			.catch(console.error);
	};

	const onExportCsv = curve.series.some(run => run.points.length) ? () => downloadString(dpsCurveCsv(curve.series), CSV_FILE_NAME, 'text/csv') : undefined;

	return (
		<Dialog
			open={open}
			onOpenChange={next => {
				onOpenChange(next);
				if (!next) curve.cancel();
			}}
			testId="dps-curve-dialog"
			size="xl"
			scrollContents
			title={t('title')}
			footer={
				<Button variant="primary" disabled={curve.isRunning || !selectedStats.length} onClick={onRun}>
					{t('run')}
				</Button>
			}>
			<DpsCurvePanel
				statOptions={statOptions}
				stats={selectedKeys}
				onStatsChange={keys => settings.setSetup({ stats: unitStats.filter(stat => keys.includes(stat.getKey())) })}
				range={range}
				onRangeChange={range => settings.setSetup({ range })}
				step={step}
				onStepChange={step => settings.setSetup({ step })}
				fractional={selectedStats.some(stat => stat.isPseudoStat())}
				precision={
					<DpsCurvePrecisionOptions
						value={precision}
						onChange={patch => settings.setPrecision(patch)}
						defaultIterations={defaultIterations}
						disabled={curve.isRunning}
					/>
				}
				chart={<DpsCurveChart series={curve.series} />}
				history={<DpsCurveHistory runs={curve.historyEntries} onOverlayChange={curve.setOverlaid} onDelete={curve.remove} onExportCsv={onExportCsv} />}
				running={curve.isRunning}
			/>
			{curve.isRunning && (
				<ProgressTrackerDialog
					ref={progressRef}
					open
					testId="dps-curve-progress-dialog"
					title={t('title')}
					state={{
						stage: 'running',
						message: (
							<>
								<HandleDpsCurveTimeRemaining ref={remainingRef} />
								<HandleDpsCurveStatList ref={statListRef} labels={selectedLabels} />
							</>
						),
					}}
					hasProgressBar
					onCancel={curve.cancel}
				/>
			)}
		</Dialog>
	);
};
