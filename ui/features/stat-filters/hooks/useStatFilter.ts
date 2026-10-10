import { useSimHost } from '@sim/context/SimHostContext';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import type { UnitStat } from '@sim/proto/stats';
import { subscribeSimField } from '@sim/state/subscriptions';
import { useMemo } from 'react';

import { isCustomised, shownStats, type StatFilterScope, type StatFilterSlot, withSelection } from '../model/stat_filters';

/**
 * One panel's stat filter. `shown` keeps its identity while it holds the same stats, so a write to another filter
 * (a favourite, the minimum item level) rebuilds nothing keyed on it. The writers read the filters at write time, so
 * a write to another field since the last render survives.
 */
export const useStatFilter = (slot: StatFilterSlot, scope: StatFilterScope) => {
	const { sim } = useSimHost();
	const filters = useStoreSubscribe(subscribeSimField(sim, 'filters'), () => sim.getFilters());
	const current = shownStats(filters, slot, scope);
	const key = current.map(stat => stat.getKey()).join(',');
	// eslint-disable-next-line react-hooks/exhaustive-deps -- the joined stat keys are the invalidation key, so an equal list keeps its identity.
	const shown = useMemo(() => current, [key]);

	return {
		shown,
		customised: isCustomised(filters, slot),
		set: (stats: UnitStat[]) => sim.setFilters(withSelection(sim.getFilters(), slot, stats)),
		reset: () => sim.setFilters(withSelection(sim.getFilters(), slot, undefined)),
	};
};
