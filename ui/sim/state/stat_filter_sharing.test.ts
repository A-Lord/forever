import { Stat } from '@generated/proto/common';
import { DatabaseFilters, IndividualSimSettings, SimSettings, StatFilters, StatSelection } from '@generated/proto/ui';
import { UnitStat } from '@sim/proto/stats';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { restoreViewerStatFilters, withoutStatFilters } from './stat_filter_sharing';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

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
	beforeAll(() => serveDatabase());

	const simWith = async (filters: DatabaseFilters) => {
		const { sim } = await makeIndividualSim();
		sim.setFilters(filters);
		return sim;
	};

	it("puts the viewer's stat filters over the imported ones and keeps every other imported filter", async () => {
		const imported = DatabaseFilters.create({ minIlvl: 50, oneHandedWeapons: true, favoriteItems: [7], statFilters: senderFilters });
		const sim = await simWith(imported);

		restoreViewerStatFilters(sim, viewerFilters);

		expect(DatabaseFilters.equals(sim.getFilters(), { ...imported, statFilters: viewerFilters })).toBe(true);
	});

	it('clears the imported stat filters when the viewer had none', async () => {
		const imported = DatabaseFilters.create({ twoHandedWeapons: true, statFilters: senderFilters });
		const sim = await simWith(imported);

		restoreViewerStatFilters(sim, undefined);

		expect(DatabaseFilters.equals(sim.getFilters(), { ...imported, statFilters: undefined })).toBe(true);
	});
});
