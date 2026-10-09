import { Stat } from '@generated/proto/common';
import { IndividualSimSettings, StatFilters, StatSelection } from '@generated/proto/ui';
import { SimSettingCategories } from '@sim/constants/sim_settings';
import { UnitStat } from '@sim/proto/stats';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { createSettingsJson } from './json';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase());

describe('createSettingsJson', () => {
	it('leaves the stat filters out of a document that carries the other filters', async () => {
		const { sim, host } = await makeIndividualSim();
		const stats = StatSelection.create({ stats: [UnitStat.fromStat(Stat.StatStrength).toProto()] });
		sim.setFilters({ ...sim.getFilters(), minIlvl: 50, statFilters: StatFilters.create({ ep: stats }) });

		const json = createSettingsJson(host, [SimSettingCategories.Gear, SimSettingCategories.UISettings]);
		const filters = IndividualSimSettings.fromJsonString(json).settings!.filters!;

		expect(json).not.toContain('statFilters');
		expect(filters.minIlvl).toBe(50);
		expect(sim.getFilters().statFilters).toBeDefined();
	});
});
