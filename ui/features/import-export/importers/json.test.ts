import { Player } from '@generated/proto/api';
import { Stat } from '@generated/proto/common';
import { DatabaseFilters, IndividualSimSettings, SimSettings, StatFilters, StatSelection } from '@generated/proto/ui';
import { UnitStat } from '@sim/proto/stats';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { JSON_IMPORTER } from './json';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

const selectionOf = (...stats: Stat[]) => StatSelection.create({ stats: stats.map(stat => UnitStat.fromStat(stat).toProto()) });
const viewerStatFilters = StatFilters.create({ characterStats: selectionOf(Stat.StatIntellect), consumables: selectionOf(Stat.StatStamina) });
const senderStatFilters = StatFilters.create({ ep: selectionOf(Stat.StatStrength), buffs: selectionOf() });

const document = (settings?: SimSettings) =>
	IndividualSimSettings.toJsonString(IndividualSimSettings.create({ apiVersion: 3, player: Player.create(), settings }));

// The viewer's filters before the import: a hand filter off, a favourite and their own stat filters.
const viewer = async (statFilters: StatFilters | undefined) => {
	const built = await makeIndividualSim();
	const filters = built.sim.getFilters();
	built.sim.setFilters({ ...filters, twoHandedWeapons: false, favoriteItems: [19019], statFilters });
	return built;
};

// What the same import leaves on a viewer the stat filters never touch, less the stat filters.
const withoutStatFiltersAfter = async (data: string) => {
	const { sim, host } = await viewer(undefined);
	host.fromProto(IndividualSimSettings.fromJsonString(data));
	return { ...sim.getFilters(), statFilters: undefined };
};

beforeAll(() => serveDatabase());

describe('JSON_IMPORTER', () => {
	it("keeps the viewer's stat filters over the ones the document carries", async () => {
		const { sim, host } = await viewer(viewerStatFilters);
		const data = document(SimSettings.create({ filters: DatabaseFilters.create({ minIlvl: 50, statFilters: senderStatFilters }) }));

		await JSON_IMPORTER.onImport(host, data);

		expect(StatFilters.equals(sim.getFilters().statFilters!, viewerStatFilters)).toBe(true);
		expect(sim.getFilters().minIlvl).toBe(50);
	});

	it.each([
		['settings without filters', document(SimSettings.create({ iterations: 1000 }))],
		['no settings at all', document()],
	])("keeps the viewer's stat filters through a document with %s, and every other filter as before", async (_name, data) => {
		const { sim, host } = await viewer(viewerStatFilters);

		await JSON_IMPORTER.onImport(host, data);

		expect(StatFilters.equals(sim.getFilters().statFilters!, viewerStatFilters)).toBe(true);
		expect(DatabaseFilters.equals({ ...sim.getFilters(), statFilters: undefined }, await withoutStatFiltersAfter(data))).toBe(true);
	});

	it('leaves a viewer with no stat filters without any', async () => {
		const { sim, host } = await viewer(undefined);

		await JSON_IMPORTER.onImport(host, document(SimSettings.create({ filters: DatabaseFilters.create({ statFilters: senderStatFilters }) })));

		expect(sim.getFilters().statFilters).toBeUndefined();
	});
});
