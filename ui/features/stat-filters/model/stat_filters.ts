import { DatabaseFilters, StatFilters, StatSelection } from '@generated/proto/ui';
import { UnitStat } from '@sim/proto/stats';

export type StatFilterSlot = 'ep' | 'characterStats' | 'buffs' | 'debuffs' | 'consumables';

export interface StatFilterScope {
	/** The spec's list for the panel, shown while its slot is unset. */
	defaults: ReadonlyArray<UnitStat>;
	/** Every stat the panel can toggle, in display order. */
	universe: ReadonlyArray<UnitStat>;
}

const slotFields: Record<StatFilterSlot, keyof StatFilters> = {
	ep: 'ep',
	characterStats: 'characterStats',
	buffs: 'buffs',
	debuffs: 'debuffs',
	consumables: 'consumables',
};

const contains = (stats: ReadonlyArray<UnitStat>, stat: UnitStat) => stats.some(other => other.equals(stat));

const selectionOf = (filters: DatabaseFilters, slot: StatFilterSlot): StatSelection | undefined => filters.statFilters?.[slotFields[slot]];

export const isCustomised = (filters: DatabaseFilters, slot: StatFilterSlot): boolean => selectionOf(filters, slot) !== undefined;

/** The slot's stored stats, or the defaults while it is unset, kept to the universe and in its order. */
export const shownStats = (filters: DatabaseFilters, slot: StatFilterSlot, scope: StatFilterScope): UnitStat[] => {
	const selection = selectionOf(filters, slot);
	const wanted = selection ? selection.stats.map(stat => UnitStat.fromProto(stat)) : scope.defaults;
	return scope.universe.filter(stat => contains(wanted, stat));
};

/** A copy of the filters with the slot set to the given stats, or cleared back to the defaults for `undefined`. */
export const withSelection = (filters: DatabaseFilters, slot: StatFilterSlot, stats: ReadonlyArray<UnitStat> | undefined): DatabaseFilters => {
	const next = DatabaseFilters.clone(filters);
	if (!stats && !next.statFilters) return next;
	next.statFilters ??= StatFilters.create();
	next.statFilters[slotFields[slot]] = stats ? StatSelection.create({ stats: stats.map(stat => stat.toProto()) }) : undefined;
	return next;
};
