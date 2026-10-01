import type { StatWeightsResult } from '@generated/proto/api';
import type { Stat } from '@generated/proto/common';
import type { DisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import type { Player } from '@sim/player/player';
import type { Stats, UnitStat } from '@sim/proto/stats';
import type { StatWeightActionSettings } from '@sim/settings/stat_weight_settings';
import { useCallback } from 'react';

import type { EpStatSet } from '../../model/ep_unit_stats';
import { isEpStat } from '../../model/ep_unit_stats';
import { EpWeightsTableView } from './EpWeightsTableView';
import { StatWeightRow } from './StatWeightRow';
import type { EpColumn, StatsType } from './types';

export interface EpWeightsTableProps {
	columns: EpColumn[];
	stats: UnitStat[];
	statsType: StatsType;
	result: StatWeightsResult | null;
	iterations: number;
	epRatios: number[];
	epWeights: Stats;
	settings: StatWeightActionSettings;
	player: Player<any>;
	epStatSet: EpStatSet;
	epReferenceStat: Stat;
	onComputeEp: () => void;
	showThreatMetrics: boolean;
	showEpRatios: boolean;
	displayMetrics: DisplayMetrics;
	showUpdateColumn?: boolean;
}

export const EpWeightsTable = ({
	columns,
	stats,
	statsType,
	result,
	iterations,
	epRatios,
	epWeights,
	settings,
	player,
	epStatSet,
	epReferenceStat,
	onComputeEp,
	showThreatMetrics,
	showEpRatios,
	displayMetrics,
	showUpdateColumn = true,
}: EpWeightsTableProps) => {
	const onEpRatioChange = useCallback(
		(ratioIndex: number, value: number) => {
			const next = player.getEpRatios();
			next[ratioIndex] = value;
			player.setEpRatios(next);
		},
		[player],
	);

	return (
		<EpWeightsTableView
			columns={columns}
			statsType={statsType}
			epRatios={epRatios}
			onEpRatioChange={onEpRatioChange}
			onComputeEp={onComputeEp}
			showThreatMetrics={showThreatMetrics}
			showEpRatios={showEpRatios}
			showUpdateColumn={showUpdateColumn}>
			{stats.map(stat => (
				<StatWeightRow
					key={stat.getFullName(player.getClass())}
					stat={stat}
					result={result}
					iterations={iterations}
					epRatios={epRatios}
					epWeights={epWeights}
					settings={settings}
					player={player}
					epReferenceStat={epReferenceStat}
					includable={isEpStat(stat, epStatSet)}
					showThreatMetrics={showThreatMetrics}
					displayMetrics={displayMetrics}
					showUpdateColumn={showUpdateColumn}
				/>
			))}
		</EpWeightsTableView>
	);
};
