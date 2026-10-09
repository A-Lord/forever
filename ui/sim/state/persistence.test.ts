import { Player } from '@generated/proto/api';
import { Stat } from '@generated/proto/common';
import { DatabaseFilters, IndividualSimSettings, SimSettings, StatFilters, StatSelection } from '@generated/proto/ui';
import { UnitStat } from '@sim/proto/stats';
import { makeIndividualSim, memoryEnv, serveDatabase } from '@testing/sim/individual_sim';
import pako from 'pako';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import type { StatWeightActionSettings } from '../settings/stat_weight_settings';
import { loadIndividualSettings } from './persistence';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

const STORAGE_KEY = 'test__currentSettings__';

const selectionOf = (...stats: Stat[]) => StatSelection.create({ stats: stats.map(stat => UnitStat.fromStat(stat).toProto()) });
const savedStatFilters = StatFilters.create({ characterStats: selectionOf(Stat.StatIntellect), consumables: selectionOf(Stat.StatStamina) });
const linkedStatFilters = StatFilters.create({ ep: selectionOf(Stat.StatStrength), buffs: selectionOf() });

const envelope = (settings?: SimSettings) => IndividualSimSettings.create({ apiVersion: 3, player: Player.create(), settings });

const savedWith = (statFilters: StatFilters | undefined) =>
	envelope(SimSettings.create({ filters: DatabaseFilters.create({ twoHandedWeapons: false, favoriteItems: [19019], statFilters }) }));

// A link's hash, and the `i` parameter that names its categories: `u` is the one that carries settings.
const linkLocation = (settings: IndividualSimSettings, categories: string | null) => ({
	hash: '#' + btoa(String.fromCharCode(...pako.deflate(IndividualSimSettings.toBinary(settings)))),
	search: categories == null ? '' : `?i=${categories}`,
});

const load = async (location: { hash: string; search: string } = { hash: '', search: '' }, saved = savedWith(savedStatFilters)) => {
	const env = memoryEnv(location);
	env.storage.setItem(STORAGE_KEY, IndividualSimSettings.toJsonString(saved));
	const { sim, player, host } = await makeIndividualSim({ env });
	loadIndividualSettings(host, {
		storageKey: STORAGE_KEY,
		player,
		autosaveSubscribe: () => () => {},
		statWeightSettings: { load: () => {} } as unknown as StatWeightActionSettings,
	});
	return sim.getFilters();
};

const lessStatFilters = (filters: DatabaseFilters) => ({ ...filters, statFilters: undefined });

beforeAll(() => serveDatabase());

describe('loadIndividualSettings', () => {
	it('restores the autosaved stat filters', async () => {
		expect(StatFilters.equals((await load()).statFilters!, savedStatFilters)).toBe(true);
	});

	it('keeps the autosaved stat filters over the ones a link carries, and takes its other filters', async () => {
		const linked = envelope(SimSettings.create({ filters: DatabaseFilters.create({ minIlvl: 50, statFilters: linkedStatFilters }) }));

		const filters = await load(linkLocation(linked, 'gu'));

		expect(StatFilters.equals(filters.statFilters!, savedStatFilters)).toBe(true);
		expect(filters.minIlvl).toBe(50);
	});

	it.each([
		['settings without filters', envelope(SimSettings.create({ iterations: 1000 }))],
		['no settings at all', envelope()],
	])('keeps the autosaved stat filters through a link with %s, and every other filter as before', async (_name, linked) => {
		const filters = await load(linkLocation(linked, 'gu'));

		const withoutStatFiltersInPlay = await load(linkLocation(linked, 'gu'), savedWith(undefined));
		expect(StatFilters.equals(filters.statFilters!, savedStatFilters)).toBe(true);
		expect(DatabaseFilters.equals(lessStatFilters(filters), lessStatFilters(withoutStatFiltersInPlay))).toBe(true);
	});

	it('leaves every autosaved filter as it was under a link that carries no settings category', async () => {
		const linked = envelope(SimSettings.create({ filters: DatabaseFilters.create({ minIlvl: 50, statFilters: linkedStatFilters }) }));

		expect(DatabaseFilters.equals(await load(linkLocation(linked, null)), await load())).toBe(true);
	});
});
