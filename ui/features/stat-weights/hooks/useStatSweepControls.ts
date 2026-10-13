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
	const { sweepStats, sweepRange: range, sweepStep: step } = useStoreSubscribe(subscribeStatWeightsChange(settings), () => settings.sweep);
	// The stored stats the spec offers, in the order offered; with none stored, or none of them offered, the first one offered. A cleared list stays empty.
	const selectedStats = useMemo(() => {
		if (sweepStats?.length === 0) return [];
		const offered = unitStats.filter(stat => sweepStats?.some(stored => stored.equals(stat)));
		return offered.length ? offered : unitStats.slice(0, 1);
	}, [unitStats, sweepStats]);

	const currentStats = usePlayerStore('currentStats');
	const refStat = useStoreSubscribe(subscribePlayerField(player, 'epRefStat'), () => player.getRefStat('dpsRefStat')) ?? epReferenceStat;
	const precision = useStoreSubscribe(subscribeStatWeightsChange(settings), () => settings.precision);

	const onRun = () => {
		if (!selectedStats.length) return;
		const finalStats = currentStats?.finalStats ? Stats.fromProto(currentStats.finalStats) : null;
		sweep
			.run({
				stats: selectedStats.map(unitStat => ({
					unitStat,
					label: i18n.t('sidebar.buttons.stat_weights.sweep.run_label', {
						stat: statOptions.find(option => option.value === unitStat.getKey())?.label ?? '',
						range,
					}),
					baseStatValue: finalStats?.getUnitStat(unitStat) ?? 0,
				})),
				range,
				step,
				refStat,
				precision,
			})
			.catch(console.error);
	};

	return { sweep, unitStats, selectedStats, statOptions, range, step, precision, onRun };
};

export type StatSweepControls = ReturnType<typeof useStatSweepControls>;
