import { Stat } from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import type { Player } from '../player/player';
import { UnitStat } from '../proto/stats';
import { createSimStore } from '../state/sim_store';
import { StatWeightActionSettings } from './stat_weight_settings';

const KEY = 'stat-weight-settings-test';

const settingsOn = (storage: Map<string, string>) => {
	const player = {
		storeKey: 0,
		sim: {
			store: createSimStore(),
			env: { storage: { getItem: (key: string) => storage.get(key) ?? null, setItem: (key: string, value: string) => storage.set(key, value) } },
		},
	} as unknown as Player<any>;
	return new StatWeightActionSettings(player, KEY);
};

describe('StatWeightActionSettings precision', () => {
	it('starts on the fixed iteration count with the default target and multiplier', () => {
		expect(settingsOn(new Map()).precision).toEqual({ precisionMode: 'iterations', iterations: 0, targetEpCi: 0.01, maxIterationMultiplier: 8 });
	});

	it('loads back what it stored, next to the excluded stats', () => {
		const storage = new Map<string, string>();
		const saved = settingsOn(storage);
		saved.setStatExcluded(UnitStat.fromStat(Stat.StatStrength), true);
		saved.setPrecision({ precisionMode: 'target', iterations: 4000, targetEpCi: 0.005, maxIterationMultiplier: 4 });

		const loaded = settingsOn(storage);
		loaded.load();

		expect(loaded.precision).toEqual({ precisionMode: 'target', iterations: 4000, targetEpCi: 0.005, maxIterationMultiplier: 4 });
		expect(loaded.excludedStats).toEqual([Stat.StatStrength]);
	});

	it('falls back to the defaults for stored values it cannot use', () => {
		const storage = new Map([[KEY, JSON.stringify({ precisionMode: 'fast', iterations: -5, targetEpCi: 0, maxIterationMultiplier: 'x' })]]);
		const loaded = settingsOn(storage);
		loaded.load();

		expect(loaded.precision).toEqual({ precisionMode: 'iterations', iterations: 0, targetEpCi: 0.01, maxIterationMultiplier: 8 });
	});

	it('does not rewrite storage for a patch that changes nothing', () => {
		const storage = new Map<string, string>();
		const settings = settingsOn(storage);
		settings.setPrecision({ targetEpCi: 0.01 });

		expect(storage.has(KEY)).toBe(false);
	});
});
