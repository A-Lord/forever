import { PseudoStat, Stat } from '@generated/proto/common';
import { SavedStatWeightSettings } from '@generated/proto/ui';
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
	it('starts on the fixed iteration count with the default target and no multiplier cap', () => {
		expect(settingsOn(new Map()).precision).toEqual({ precisionMode: 'iterations', iterations: 0, targetEpCi: 0.01, maxIterationMultiplier: 0 });
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

	it('reads unset precision fields as the defaults', () => {
		const storage = new Map([[KEY, SavedStatWeightSettings.toJsonString(SavedStatWeightSettings.create({ excludedStats: [Stat.StatAgility] }))]]);
		const loaded = settingsOn(storage);
		loaded.load();

		expect(loaded.precision).toEqual({ precisionMode: 'iterations', iterations: 0, targetEpCi: 0.01, maxIterationMultiplier: 0 });
		expect(loaded.excludedStats).toEqual([Stat.StatAgility]);
	});

	it('loads a cleared multiplier back as no cap', () => {
		const storage = new Map<string, string>();
		const saved = settingsOn(storage);
		saved.setPrecision({ precisionMode: 'target', maxIterationMultiplier: 4 });
		saved.setPrecision({ maxIterationMultiplier: 0 });

		const loaded = settingsOn(storage);
		loaded.load();

		expect(loaded.precision.maxIterationMultiplier).toBe(0);
	});

	it('reads a stored mode name the proto does not list as the fixed count, and keeps the other fields', () => {
		const stored = JSON.stringify({
			excludedStats: ['StatAgility'],
			precisionMode: 'target',
			iterations: 4000,
			targetEpCi: 0.005,
			maxIterationMultiplier: 4,
		});
		const loaded = settingsOn(new Map([[KEY, stored]]));
		loaded.load();

		expect(loaded.precision).toEqual({ precisionMode: 'iterations', iterations: 4000, targetEpCi: 0.005, maxIterationMultiplier: 4 });
		expect(loaded.excludedStats).toEqual([Stat.StatAgility]);
	});

	it('does not rewrite storage for a patch that changes nothing', () => {
		const storage = new Map<string, string>();
		const settings = settingsOn(storage);
		settings.setPrecision({ targetEpCi: 0.01 });

		expect(storage.has(KEY)).toBe(false);
	});
});

describe('StatWeightActionSettings sweep', () => {
	it('starts with no stat, a range of 100 and a step of 10', () => {
		expect(settingsOn(new Map()).sweep).toEqual({ sweepStat: null, sweepRange: 100, sweepStep: 10 });
	});

	it('loads back the stored stat, range and step, a pseudo stat in fractions of a point', () => {
		const storage = new Map<string, string>();
		settingsOn(storage).setSweep({ sweepStat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellHitPercent), sweepRange: 2.5, sweepStep: 0.5 });

		const loaded = settingsOn(storage);
		loaded.load();

		expect(loaded.sweep.sweepStat?.equals(UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellHitPercent))).toBe(true);
		expect([loaded.sweep.sweepRange, loaded.sweep.sweepStep]).toEqual([2.5, 0.5]);
	});

	it('reads a stored stat that names neither a stat nor a pseudo stat as no stat', () => {
		const loaded = settingsOn(new Map([[KEY, JSON.stringify({ sweepStat: { apiVersion: 1 }, sweepRange: 50 })]]));
		loaded.load();

		expect(loaded.sweep).toEqual({ sweepStat: null, sweepRange: 50, sweepStep: 10 });
	});

	it('does not rewrite storage for the same stat in a new instance', () => {
		const storage = new Map<string, string>();
		const settings = settingsOn(storage);
		settings.setSweep({ sweepStat: UnitStat.fromStat(Stat.StatHitRating) });
		storage.clear();
		settings.setSweep({ sweepStat: UnitStat.fromStat(Stat.StatHitRating), sweepRange: 100 });

		expect(storage.has(KEY)).toBe(false);
	});
});
