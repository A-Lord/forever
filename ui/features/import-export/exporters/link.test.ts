import { Stat } from '@generated/proto/common';
import { StatFilters, StatSelection } from '@generated/proto/ui';
import { SimSettingCategories } from '@sim/constants/sim_settings';
import { UnitStat } from '@sim/proto/stats';
import { tryParseUrlLocation } from '@sim/state/sim_links';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { createLink } from './link';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase());

describe('createLink', () => {
	it('leaves the stat filters out of a link that carries the other filters', async () => {
		const { sim, host } = await makeIndividualSim();
		const stats = StatSelection.create({ stats: [UnitStat.fromStat(Stat.StatStrength).toProto()] });
		sim.setFilters({ ...sim.getFilters(), minIlvl: 50, statFilters: StatFilters.create({ ep: stats }) });

		const url = new URL(createLink(host, [SimSettingCategories.Gear, SimSettingCategories.UISettings]));
		const filters = tryParseUrlLocation(url)!.settings.settings!.filters!;

		expect(filters.statFilters).toBeUndefined();
		expect(filters.minIlvl).toBe(50);
		expect(sim.getFilters().statFilters).toBeDefined();
	});
});
