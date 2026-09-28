import type { StatWeightsResult, StatWeightValues } from '@generated/proto/api';
import type { Stat } from '@generated/proto/common';
import type { DisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import type { Player } from '@sim/player/player';
import { scaledEpValue, type Stats, type UnitStat } from '@sim/proto/stats';
import type { StatWeightActionSettings } from '@sim/settings/stat_weight_settings';
import { subscribeStatWeightsChange } from '@sim/state/subscriptions';

import { StatWeightRowView } from './StatWeightRowView';

export interface StatWeightRowProps {
	stat: UnitStat;
	result: StatWeightsResult | null;
	iterations: number;
	epRatios: number[];
	epWeights: Stats;
	settings: StatWeightActionSettings;
	player: Player<any>;
	epReferenceStat: Stat;
	includable: boolean;
	showThreatMetrics: boolean;
	displayMetrics: DisplayMetrics;
	showUpdateColumn?: boolean;
}

export const StatWeightRow = ({
	stat,
	result,
	iterations,
	epRatios,
	epWeights,
	settings,
	player,
	epReferenceStat,
	includable,
	showThreatMetrics,
	displayMetrics,
	showUpdateColumn = true,
}: StatWeightRowProps) => {
	const excluded = useStoreSubscribe(subscribeStatWeightsChange(settings), () => settings.isUnitStatExcludedFromCalc(stat));
	const rowResult = excluded ? null : result;
	const allMetrics: { statWeights: StatWeightValues | undefined; metric: keyof DisplayMetrics }[] = [
		{ statWeights: rowResult?.dps, metric: 'damage' },
		{ statWeights: rowResult?.hps, metric: 'healing' },
		{ statWeights: rowResult?.tps, metric: 'threat' },
		{ statWeights: rowResult?.dtps, metric: 'threat' },
		{ statWeights: rowResult?.tmi, metric: 'threat' },
		{ statWeights: rowResult?.pDeath, metric: 'threat' },
	];
	const metrics = allMetrics
		.map(({ statWeights, metric }, ratioIndex) => ({ statWeights, metric, ratioIndex, epRatio: epRatios[ratioIndex] }))
		.filter(entry => displayMetrics[entry.metric]);
	const currentEp = epWeights.getUnitStat(stat);

	return (
		<StatWeightRowView
			stat={stat}
			fullName={stat.getFullName(player.getClass())}
			shortName={stat.getShortName(player.playerClass.classID)}
			metrics={metrics}
			iterations={iterations}
			epDelta={scaledEpValue(stat, epRatios, rowResult) - currentEp}
			currentEp={currentEp}
			onCurrentEpChange={value => player.setEpWeights(player.getEpWeights().withUnitStat(stat, value))}
			include={
				includable
					? {
							value: !excluded,
							disabled: stat.isStat() && epReferenceStat === stat.getStat(),
							onChange: included => settings.setStatExcluded(stat, !included),
						}
					: undefined
			}
			showThreatMetrics={showThreatMetrics}
			showUpdateColumn={showUpdateColumn}
		/>
	);
};
