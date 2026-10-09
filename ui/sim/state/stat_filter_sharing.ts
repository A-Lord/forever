import { DatabaseFilters, IndividualSimSettings, StatFilters } from '@generated/proto/ui';

// Stat filters belong to the viewer: an exported build carries none and an imported one keeps the viewer's.

export const withoutStatFilters = (proto: IndividualSimSettings): IndividualSimSettings => {
	const next = IndividualSimSettings.clone(proto);
	if (next.settings?.filters) next.settings.filters.statFilters = undefined;
	return next;
};

export const keepViewerStatFilters = (current: DatabaseFilters, proto: IndividualSimSettings): IndividualSimSettings => {
	const next = IndividualSimSettings.clone(proto);
	if (!next.settings) return next;
	if (current.statFilters) {
		next.settings.filters ??= DatabaseFilters.create();
		next.settings.filters.statFilters = StatFilters.clone(current.statFilters);
	} else if (next.settings.filters) {
		next.settings.filters.statFilters = undefined;
	}
	return next;
};
