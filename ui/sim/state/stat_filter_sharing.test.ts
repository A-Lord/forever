import { Stat } from '@generated/proto/common';
import { DatabaseFilters, IndividualSimSettings, SimSettings, StatFilters, StatSelection } from '@generated/proto/ui';
import { UnitStat } from '@sim/proto/stats';
import { describe, expect, it } from 'vitest';

import { createSimStore, patchSlice } from './sim_store';
import { restoreViewerStatFilters, withoutStatFilters } from './stat_filter_sharing';

const selectionOf = (...stats: Stat[]) => StatSelection.create({ stats: stats.map(stat => UnitStat.fromStat(stat).toProto()) });

const senderFilters = StatFilters.create({ ep: selectionOf(Stat.StatStrength), buffs: selectionOf() });
const viewerFilters = StatFilters.create({ characterStats: selectionOf(Stat.StatIntellect, Stat.StatSpirit) });

const settingsWith = (filters: DatabaseFilters | undefined) =>
	IndividualSimSettings.create({ apiVersion: 3, settings: SimSettings.create({ iterations: 1234, filters }) });

describe('withoutStatFilters', () => {
	it('clears the stat filters and keeps every other filter', () => {
		const proto = settingsWith(DatabaseFilters.create({ minIlvl: 50, showAllEnchants: true, favoriteItems: [7], statFilters: senderFilters }));
		const before = IndividualSimSettings.clone(proto);

		const stripped = withoutStatFilters(proto);

		expect(stripped.settings?.filters?.statFilters).toBeUndefined();
		expect(IndividualSimSettings.equals(proto, before)).toBe(true);
		const expected = IndividualSimSettings.clone(proto);
		expected.settings!.filters!.statFilters = undefined;
		expect(IndividualSimSettings.equals(stripped, expected)).toBe(true);
	});

	it('passes a proto without settings or filters through', () => {
		const noSettings = IndividualSimSettings.create({ apiVersion: 3 });
		const noFilters = settingsWith(undefined);

		expect(IndividualSimSettings.equals(withoutStatFilters(noSettings), noSettings)).toBe(true);
		expect(IndividualSimSettings.equals(withoutStatFilters(noFilters), noFilters)).toBe(true);
	});
});

describe('restoreViewerStatFilters', () => {
	const simWith = (filters: DatabaseFilters) => {
		const store = createSimStore();
		patchSlice(store, 'sim', { filters });
		return {
			getFilters: () => DatabaseFilters.clone(store.getState().sim.filters),
			setFilters: (next: DatabaseFilters) => patchSlice(store, 'sim', { filters: DatabaseFilters.clone(next) }),
		};
	};

	it("puts the viewer's stat filters over the imported ones and keeps every other imported filter", () => {
		const imported = DatabaseFilters.create({ minIlvl: 50, oneHandedWeapons: true, favoriteItems: [7], statFilters: senderFilters });
		const sim = simWith(imported);

		restoreViewerStatFilters(sim, viewerFilters);

		expect(DatabaseFilters.equals(sim.getFilters(), { ...imported, statFilters: viewerFilters })).toBe(true);
	});

	it('clears the imported stat filters when the viewer had none', () => {
		const imported = DatabaseFilters.create({ twoHandedWeapons: true, statFilters: senderFilters });
		const sim = simWith(imported);

		restoreViewerStatFilters(sim, undefined);

		expect(DatabaseFilters.equals(sim.getFilters(), { ...imported, statFilters: undefined })).toBe(true);
	});
});
