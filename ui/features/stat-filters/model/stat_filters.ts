import { DatabaseFilters, StatFilters, StatSelection } from '@generated/proto/ui';
import { UnitStat } from '@sim/proto/stats';

export type StatFilterSlot = keyof StatFilters;

export interface StatFilterScope {
	/** The spec's list for the panel, shown while its slot is unset. */
	defaults: ReadonlyArray<UnitStat>;
	/** Every stat the panel can toggle, in display order. */
	universe: ReadonlyArray<UnitStat>;
}

const contains = (stats: ReadonlyArray<UnitStat>, stat: UnitStat) => stats.some(other => other.equals(stat));

export const statsKey = (stats: ReadonlyArray<UnitStat>): string => stats.map(stat => stat.getKey()).join(',');

export const isCustomised = (filters: DatabaseFilters, slot: StatFilterSlot): boolean => filters.statFilters?.[slot] !== undefined;

/** The slot's stored stats, or the defaults while it is unset, kept to the universe and in its order. */
export const shownStats = (filters: DatabaseFilters, slot: StatFilterSlot, scope: StatFilterScope): UnitStat[] => {
	const selection = filters.statFilters?.[slot];
	const wanted = selection ? selection.stats.map(stat => UnitStat.fromProto(stat)) : scope.defaults;
	return scope.universe.filter(stat => contains(wanted, stat));
};

/** A copy of the filters with the slot set to the given stats, or cleared back to the defaults for `undefined`. */
export const withSelection = (filters: DatabaseFilters, slot: StatFilterSlot, stats: ReadonlyArray<UnitStat> | undefined): DatabaseFilters => {
	const next = DatabaseFilters.clone(filters);
	if (!stats && !next.statFilters) return next;
	next.statFilters ??= StatFilters.create();
	next.statFilters[slot] = stats ? StatSelection.create({ stats: stats.map(stat => stat.toProto()) }) : undefined;
	return next;
};
