import i18n from '@i18n/config';
import { usePlayer, useSpecConfig } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { Stats, UnitStat } from '@sim/proto/stats';
import type { StatWeightActionSettings } from '@sim/settings/stat_weight_settings';
import { subscribePlayerField, subscribeStatWeightsChange } from '@sim/state/subscriptions';
import { useMemo } from 'react';

import { EP_UNIT_STATS } from '../model/ep_unit_stats';
import { useStatSweep, type UseStatSweepOptions } from './useStatSweep';

// The sweep's stored setup and run state, shared by the sweep tab's panel, the dialog footer's Run sweep button and the progress dialog.
export const useStatSweepControls = (settings: StatWeightActionSettings, options?: UseStatSweepOptions) => {
	const player = usePlayer();
	const { epStats, epPseudoStats, epReferenceStat } = useSpecConfig();
	const sweep = useStatSweep(options);

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

	return { sweep, unitStats, unitStat, statOptions, range, step, precision, onRun };
};

export type StatSweepControls = ReturnType<typeof useStatSweepControls>;
