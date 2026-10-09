import { useSimHost } from '@sim/context/SimHostContext';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import type { UnitStat } from '@sim/proto/stats';
import { subscribeSimField } from '@sim/state/subscriptions';
import { useMemo } from 'react';

import { isCustomised, shownStats, type StatFilterScope, type StatFilterSlot, withSelection } from '../model/stat_filters';

/**
 * One panel's stat filter. `scope` keys the memoised `shown`, so pass a stable (memoised) object. The writers read the
 * filters at write time, so a write to another field since the last render survives.
 */
export const useStatFilter = (slot: StatFilterSlot, scope: StatFilterScope) => {
	const { sim } = useSimHost();
	const filters = useStoreSubscribe(subscribeSimField(sim, 'filters'), () => sim.getFilters());
	const shown = useMemo(() => shownStats(filters, slot, scope), [filters, slot, scope]);

	return {
		shown,
		customised: isCustomised(filters, slot),
		set: (stats: UnitStat[]) => sim.setFilters(withSelection(sim.getFilters(), slot, stats)),
		reset: () => sim.setFilters(withSelection(sim.getFilters(), slot, undefined)),
	};
};
