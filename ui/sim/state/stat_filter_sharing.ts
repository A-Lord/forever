import { IndividualSimSettings, StatFilters } from '@generated/proto/ui';

import type { Sim } from '../sim';

// Stat filters belong to the viewer: an exported build carries none and an imported one keeps the viewer's.

export const withoutStatFilters = (proto: IndividualSimSettings): IndividualSimSettings => {
	const next = IndividualSimSettings.clone(proto);
	if (next.settings?.filters) next.settings.filters.statFilters = undefined;
	return next;
};

// Puts the stat filters read before an import back over the filters the import left, every other
// filter as the import set it: loading settings replaces the filters whole, with defaults where the
// settings carry none.
export const restoreViewerStatFilters = (sim: Pick<Sim, 'getFilters' | 'setFilters'>, statFilters: StatFilters | undefined) => {
	const filters = sim.getFilters();
	filters.statFilters = statFilters && StatFilters.clone(statFilters);
	sim.setFilters(filters);
};
