import { useSimHost } from '@sim/context/SimHostContext';
import type { UnitStat } from '@sim/proto/stats';
import { useMemo } from 'react';
import { useStore } from 'zustand';

import { isCustomised, shownStats, type StatFilterScope, type StatFilterSlot, statsKey, withSelection } from '../model/stat_filters';

/**
 * One panel's stat filter, shaped as the stat filter popover's props. The host re-renders only when this slot's shown
 * stats or its customised flag change, and `shown` keeps its identity while it holds the same stats, so a write to
 * another filter (a favourite, the minimum item level) rebuilds nothing. The writers read the filters at write time,
 * so a write to another field since the last render survives.
 */
export const useStatFilter = (slot: StatFilterSlot, scope: StatFilterScope) => {
	const { sim, player } = useSimHost();
	const key = useStore(sim.store, state => statsKey(shownStats(state.sim.filters, slot, scope)));
	const customised = useStore(sim.store, state => isCustomised(state.sim.filters, slot));
	const shown = useMemo(() => {
		const keys = new Set(key.split(','));
		return scope.universe.filter(stat => keys.has(stat.getKey()));
	}, [key, scope.universe]);

	return {
		universe: scope.universe,
		shown,
		customised,
		playerClass: player.getClass(),
		onChange: (stats: UnitStat[]) => sim.setFilters(withSelection(sim.getFilters(), slot, stats)),
		onReset: () => sim.setFilters(withSelection(sim.getFilters(), slot, undefined)),
	};
};
