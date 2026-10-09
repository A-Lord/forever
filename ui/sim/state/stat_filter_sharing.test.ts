import { Stat } from '@generated/proto/common';
import { DatabaseFilters, IndividualSimSettings, SimSettings, StatFilters, StatSelection } from '@generated/proto/ui';
import { UnitStat } from '@sim/proto/stats';
import { describe, expect, it } from 'vitest';

import { keepViewerStatFilters, withoutStatFilters } from './stat_filter_sharing';

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

describe('keepViewerStatFilters', () => {
	it("replaces the incoming stat filters with the viewer's", () => {
		const incoming = settingsWith(DatabaseFilters.create({ minIlvl: 50, statFilters: senderFilters }));
		const before = IndividualSimSettings.clone(incoming);

		const kept = keepViewerStatFilters(DatabaseFilters.create({ maxIlvl: 90, statFilters: viewerFilters }), incoming);

		expect(StatFilters.equals(kept.settings!.filters!.statFilters!, viewerFilters)).toBe(true);
		expect(kept.settings!.filters!.minIlvl).toBe(incoming.settings!.filters!.minIlvl);
		expect(kept.settings!.filters!.maxIlvl).toBe(incoming.settings!.filters!.maxIlvl);
		expect(IndividualSimSettings.equals(incoming, before)).toBe(true);
	});

	it("sets the viewer's stat filters when the incoming proto has none", () => {
		const kept = keepViewerStatFilters(DatabaseFilters.create({ statFilters: viewerFilters }), settingsWith(DatabaseFilters.create()));

		expect(StatFilters.equals(kept.settings!.filters!.statFilters!, viewerFilters)).toBe(true);
	});

	it('clears the incoming stat filters when the viewer has none', () => {
		const kept = keepViewerStatFilters(DatabaseFilters.create(), settingsWith(DatabaseFilters.create({ statFilters: senderFilters })));

		expect(kept.settings!.filters!.statFilters).toBeUndefined();
	});

	it('creates the filters only when the incoming proto has settings', () => {
		const current = DatabaseFilters.create({ statFilters: viewerFilters });

		const noFilters = keepViewerStatFilters(current, settingsWith(undefined));
		const noSettings = keepViewerStatFilters(current, IndividualSimSettings.create({ apiVersion: 3 }));

		expect(StatFilters.equals(noFilters.settings!.filters!.statFilters!, viewerFilters)).toBe(true);
		expect(noSettings.settings).toBeUndefined();
	});
});
