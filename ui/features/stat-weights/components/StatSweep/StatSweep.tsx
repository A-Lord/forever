import i18n from '@i18n/config';
import { usePlayer, useSpecConfig } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { useSimStore } from '@sim/hooks/useSimStore';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { Stats, UnitStat } from '@sim/proto/stats';
import type { StatWeightActionSettings } from '@sim/settings/stat_weight_settings';
import { subscribePlayerField, subscribeStatWeightsChange } from '@sim/state/subscriptions';
import { useMemo, useState } from 'react';

import { useStatSweep } from '../../hooks/useStatSweep';
import { EpPrecisionOptions } from '../EpPrecisionOptions/EpPrecisionOptions';
import { StatSweepChart } from './StatSweepChart';
import { StatSweepHistory } from './StatSweepHistory';
import { StatSweepPanel } from './StatSweepPanel';

export interface StatSweepProps {
	settings: StatWeightActionSettings;
}

const DEFAULT_RANGE = 100;
const DEFAULT_STEP = 10;

export const StatSweep = ({ settings }: StatSweepProps) => {
	const player = usePlayer();
	const { epStats, epPseudoStats, epReferenceStat } = useSpecConfig();
	const sweep = useStatSweep();

	const unitStats = useMemo(
		() => [...epStats.map(stat => UnitStat.fromStat(stat)), ...(epPseudoStats ?? []).map(pseudoStat => UnitStat.fromPseudoStat(pseudoStat))],
		[epStats, epPseudoStats],
	);
	const statOptions = useMemo(() => unitStats.map(stat => ({ value: stat.getKey(), label: stat.getFullName(player.getClass()) })), [unitStats, player]);
	const [statKey, setStatKey] = useState(() => unitStats[0]?.getKey() ?? '');
	const [range, setRange] = useState(DEFAULT_RANGE);
	const [step, setStep] = useState(DEFAULT_STEP);
	const unitStat = unitStats.find(stat => stat.getKey() === statKey) ?? unitStats[0];
	const statLabel = statOptions.find(option => option.value === unitStat?.getKey())?.label ?? '';

	const currentStats = usePlayerStore('currentStats');
	const currentValue = unitStat && currentStats?.finalStats ? Stats.fromProto(currentStats.finalStats).getUnitStat(unitStat) : 0;
	const refStat = useStoreSubscribe(subscribePlayerField(player, 'epRefStat'), () => player.getRefStat('dpsRefStat')) ?? epReferenceStat;
	const precision = useStoreSubscribe(subscribeStatWeightsChange(settings), () => settings.precision);
	const defaultIterations = useSimStore('iterations');

	const onRun = () => {
		if (!unitStat) return;
		sweep
			.run({
				unitStat,
				label: i18n.t('sidebar.buttons.stat_weights.sweep.run_label', { stat: statLabel, range }),
				range,
				step,
				refStat,
				baseStatValue: currentValue,
				precision,
			})
			.catch(console.error);
	};

	const series = useMemo(() => sweep.shown.map(({ run, colorSlot }) => ({ id: run.id, label: run.label, colorSlot, points: run.points })), [sweep.shown]);
	const historyEntries = sweep.history.map(run => ({
		id: run.id,
		label: run.label,
		createdAt: run.createdAt,
		overlaid: sweep.overlaid.includes(run.id),
		colorSlot: sweep.colorSlots[run.id],
	}));

	return (
		<StatSweepPanel
			statOptions={statOptions}
			stat={unitStat?.getKey() ?? ''}
			onStatChange={setStatKey}
			range={range}
			onRangeChange={setRange}
			step={step}
			onStepChange={setStep}
			fractional={unitStat?.isPseudoStat()}
			precision={
				<EpPrecisionOptions
					value={precision}
					onChange={patch => settings.setPrecision(patch)}
					defaultIterations={defaultIterations}
					disabled={sweep.isRunning}
					idPrefix="stat-sweep-precision"
				/>
			}
			chart={<StatSweepChart series={series} statLabel={statLabel} currentValue={currentValue} />}
			history={<StatSweepHistory runs={historyEntries} onOverlayChange={sweep.setOverlaid} onDelete={sweep.remove} />}
			running={sweep.isRunning}
			progress={sweep.progress}
			onRun={onRun}
			onCancel={sweep.cancel}
		/>
	);
};
