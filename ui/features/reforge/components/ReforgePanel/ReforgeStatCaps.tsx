import type { ReforgeOptimizerModel } from '@features/reforge/model/reforge_optimizer';
import type { Player } from '@sim/player/player';
import type { UnitStat } from '@sim/proto/stats';
import { Stats } from '@sim/proto/stats';
import type { ReactNode } from 'react';

import { ReforgeStatCapRow } from './ReforgeStatCapRow';
import { ReforgeStatCapsView } from './ReforgeStatCapsView';
import { isIncludedStat } from './utils';

export interface ReforgeStatCapsProps {
	model: ReforgeOptimizerModel;
	player: Player<any>;
	displayStats: UnitStat[];
	statTooltips: Partial<Record<number, ReactNode>>;
	useCustomEPValues: boolean;
}

/** The hard-cap table, rendered only while custom EP values are on. */
export const ReforgeStatCaps = ({ model, player, displayStats, statTooltips, useCustomEPValues }: ReforgeStatCapsProps) => {
	if (!useCustomEPValues) return null;

	return (
		<ReforgeStatCapsView onReset={() => model.settings.setStatCaps(model.defaults.statCaps || new Stats())}>
			{displayStats.filter(isIncludedStat).map(unitStat => (
				<ReforgeStatCapRow
					key={unitStat.getKey()}
					model={model}
					player={player}
					unitStat={unitStat}
					tooltip={unitStat.hasRootStat() ? statTooltips[unitStat.getRootStat()] : null}
				/>
			))}
		</ReforgeStatCapsView>
	);
};
