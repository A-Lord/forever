import { useSimHost } from '@sim/context/SimHostContext';
import type { UnitStat } from '@sim/proto/stats';
import { parseStringArray, useLocalStorageState } from '@ui-kit/hooks/useLocalStorageState';
import { useMemo, useState } from 'react';
import { useStore } from 'zustand';

import { isCustomised, shownStatsReader, type StatFilterScope, type StatFilterSlot, withSelection } from '../model/stat_filters';
import { defaultOpenFilterGroups } from '../model/stat_groups';

/**
 * One panel's stat filter, shaped as the stat filter popover's props. The host re-renders only when this slot's shown
 * stats or its customised flag change, and `shown` keeps its identity while it holds the same stats, so a write to
 * another filter (a favourite, the minimum item level) rebuilds nothing. The writers read the filters at write time,
 * so a write to another field since the last render survives. The open groups are kept in the browser per panel; with
 * none stored, each opening of the popover opens the groups holding a shown stat at that moment and keeps them open
 * until it closes, so unticking a group's last stat does not fold the group away.
 */
export const useStatFilter = (slot: StatFilterSlot, scope: StatFilterScope) => {
	const host = useSimHost();
	const { sim, player } = host;
	const readShown = useMemo(() => shownStatsReader(slot, scope), [slot, scope]);
	const shown = useStore(sim.store, state => readShown(state.sim.filters));
	const customised = useStore(sim.store, state => isCustomised(state.sim.filters, slot));
	const [storedGroups, setStoredGroups] = useLocalStorageState<string[] | null>(
		host.getStorageKey(`__${slot}StatFilterGroups`),
		parseStringArray,
		() => null,
	);
	const [openedGroups, setOpenedGroups] = useState<string[]>([]);

	return {
		universe: scope.universe,
		shown,
		customised,
		openGroups: storedGroups ?? openedGroups,
		onOpenGroupsChange: setStoredGroups,
		onPopoverOpenChange: (open: boolean) => {
			if (open && storedGroups === null) setOpenedGroups(defaultOpenFilterGroups(scope.universe, shown));
		},
		playerClass: player.getClass(),
		onChange: (stats: UnitStat[]) => sim.setFilters(withSelection(sim.getFilters(), slot, stats)),
		onReset: () => sim.setFilters(withSelection(sim.getFilters(), slot, undefined)),
	};
};
