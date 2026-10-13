import { UIStat } from '@generated/proto/api';
import { PseudoStat, Stat } from '@generated/proto/common';
import { SavedDpsCurveSettings } from '@generated/proto/ui';
import { describe, expect, it } from 'vitest';

import type { Player } from '../player/player';
import { UnitStat } from '../proto/stats';
import { createSimStore } from '../state/sim_store';
import { DpsCurveSettings } from './dps_curve_settings';

const KEY = 'dps-curve-settings-test';

const settingsOn = (storage: Map<string, string>) => {
	const player = {
		storeKey: 0,
		sim: {
			store: createSimStore(),
			env: { storage: { getItem: (key: string) => storage.get(key) ?? null, setItem: (key: string, value: string) => storage.set(key, value) } },
		},
	} as unknown as Player<any>;
	return new DpsCurveSettings(player, KEY);
};

describe('DpsCurveSettings', () => {
	it('starts with no stats, a range of 100, a step of 10, the global iterations and a target of 1 DPS with no cap', () => {
		const settings = settingsOn(new Map());

		expect(settings.setup).toEqual({ stats: null, range: 100, step: 10 });
		expect(settings.precision).toEqual({ precisionMode: 'iterations', iterations: 0, targetDpsCi: 1, maxIterationMultiplier: 0 });
	});

	it('loads back the stored stats in order, range, step and precision, a pseudo stat in fractions of a point', () => {
		const storage = new Map<string, string>();
		const stats = [UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellHitPercent), UnitStat.fromStat(Stat.StatCritRating)];
		const saved = settingsOn(storage);
		saved.setSetup({ stats, range: 2.5, step: 0.5 });
		saved.setPrecision({ precisionMode: 'target', iterations: 4000, targetDpsCi: 0.5, maxIterationMultiplier: 8 });

		const loaded = settingsOn(storage);
		loaded.load();

		expect(loaded.setup.stats?.map(stat => stat.getKey())).toEqual(stats.map(stat => stat.getKey()));
		expect([loaded.setup.range, loaded.setup.step]).toEqual([2.5, 0.5]);
		expect(loaded.precision).toEqual({ precisionMode: 'target', iterations: 4000, targetDpsCi: 0.5, maxIterationMultiplier: 8 });
	});

	it('reads unset fields as the defaults', () => {
		const loaded = settingsOn(new Map([[KEY, SavedDpsCurveSettings.toJsonString(SavedDpsCurveSettings.create())]]));
		loaded.load();

		expect(loaded.setup).toEqual({ stats: null, range: 100, step: 10 });
		expect(loaded.precision).toEqual({ precisionMode: 'iterations', iterations: 0, targetDpsCi: 1, maxIterationMultiplier: 0 });
	});

	it('drops a stored stat that names neither a stat nor a pseudo stat, and reads none left as no stats', () => {
		const stored = (stats: object[]) => new Map([[KEY, JSON.stringify({ stats, range: 50 })]]);
		const kept = settingsOn(stored([{ apiVersion: 1 }, UIStat.toJson(UnitStat.fromStat(Stat.StatHitRating).toProto()) as object]));
		kept.load();
		const none = settingsOn(stored([{ apiVersion: 1 }]));
		none.load();

		expect(kept.setup.stats?.map(stat => stat.getKey())).toEqual([UnitStat.fromStat(Stat.StatHitRating).getKey()]);
		expect(none.setup).toEqual({ stats: null, range: 50, step: 10 });
	});

	it('does not rewrite storage for the same stats in a new list or an unchanged precision', () => {
		const storage = new Map<string, string>();
		const settings = settingsOn(storage);
		settings.setSetup({ stats: [UnitStat.fromStat(Stat.StatHitRating), UnitStat.fromStat(Stat.StatCritRating)] });
		storage.clear();
		settings.setSetup({ stats: [UnitStat.fromStat(Stat.StatHitRating), UnitStat.fromStat(Stat.StatCritRating)], range: 100 });
		settings.setPrecision({ targetDpsCi: 1 });

		expect(storage.has(KEY)).toBe(false);
	});

	it('goes back to the defaults', () => {
		const settings = settingsOn(new Map());
		settings.setSetup({ stats: [], range: 40 });
		settings.setPrecision({ precisionMode: 'target' });
		settings.applyDefaults();

		expect(settings.setup).toEqual({ stats: null, range: 100, step: 10 });
		expect(settings.precision.precisionMode).toBe('iterations');
	});
});
