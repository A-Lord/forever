import i18n from '@i18n/config';
import { usePlayer, useSpecConfig } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { useSimStore } from '@sim/hooks/useSimStore';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { Stats, UnitStat } from '@sim/proto/stats';
import type { StatWeightActionSettings } from '@sim/settings/stat_weight_settings';
import { subscribePlayerField, subscribeStatWeightsChange } from '@sim/state/subscriptions';
import { useMemo } from 'react';

import { useStatSweep } from '../../hooks/useStatSweep';
import { EP_UNIT_STATS } from '../../model/ep_unit_stats';
import { EpPrecisionOptions } from '../EpPrecisionOptions/EpPrecisionOptions';
import { StatSweepChart } from './StatSweepChart';
import { StatSweepHistory } from './StatSweepHistory';
import { StatSweepPanel } from './StatSweepPanel';

export interface StatSweepProps {
	settings: StatWeightActionSettings;
}

export const StatSweep = ({ settings }: StatSweepProps) => {
	const player = usePlayer();
	const { epStats, epPseudoStats, epReferenceStat } = useSpecConfig();
	const sweep = useStatSweep();

	// Only the stats the EP math fills a weight for.
	const unitStats = useMemo(
		() =>
			[...epStats.map(stat => UnitStat.fromStat(stat)), ...(epPseudoStats ?? []).map(pseudoStat => UnitStat.fromPseudoStat(pseudoStat))].filter(stat =>
				EP_UNIT_STATS.some(epStat => epStat.equals(stat)),
			),
		[epStats, epPseudoStats],
	);
	const statOptions = useMemo(() => unitStats.map(stat => ({ value: stat.getKey(), label: stat.getFullName(player.getClass()) })), [unitStats, player]);
	const { sweepStat, sweepRange: range, sweepStep: step } = useStoreSubscribe(subscribeStatWeightsChange(settings), () => settings.sweep);
	const unitStat = (sweepStat && unitStats.find(stat => stat.equals(sweepStat))) || unitStats[0];
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

	return (
		<StatSweepPanel
			statOptions={statOptions}
			stat={unitStat?.getKey() ?? ''}
			onStatChange={key => settings.setSweep({ sweepStat: unitStats.find(stat => stat.getKey() === key) ?? null })}
			range={range}
			onRangeChange={sweepRange => settings.setSweep({ sweepRange })}
			step={step}
			onStepChange={sweepStep => settings.setSweep({ sweepStep })}
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
			chart={<StatSweepChart series={sweep.series} />}
			history={<StatSweepHistory runs={sweep.historyEntries} onOverlayChange={sweep.setOverlaid} onDelete={sweep.remove} />}
			running={sweep.isRunning}
			progress={sweep.progress}
			onRun={onRun}
			onCancel={sweep.cancel}
		/>
	);
};
