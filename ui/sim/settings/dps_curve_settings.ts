import { DpsCurvePrecisionMode as SavedPrecisionMode, SavedDpsCurveSettings } from '@generated/proto/ui';

import { CURRENT_API_VERSION } from '../constants/other';
import type { Player } from '../player/player';
import { UnitStat } from '../proto/stats';
import type { Env } from '../state/env';
import { type DpsCurveSlice, patchKeyed, seedKeyed, type SimStore } from '../state/sim_store';
import { subscribeDpsCurveChange } from '../state/subscriptions';

export type DpsCurveSetup = Pick<DpsCurveSlice, 'stats' | 'range' | 'step'>;

export type DpsCurvePrecision = Pick<DpsCurveSlice, 'precisionMode' | 'iterations' | 'targetDpsCi' | 'maxIterationMultiplier'>;

type DpsCurveValues = Omit<DpsCurveSlice, 'v'>;

const DEFAULTS: DpsCurveValues = {
	stats: null,
	range: 100,
	step: 10,
	precisionMode: 'iterations',
	iterations: 0,
	targetDpsCi: 1,
	maxIterationMultiplier: 0,
};

// Proto3 cannot tell an unset field from 0, or an unset list from an empty one, so a zero range, step or target and
// no stored stat read as the default.
const fromProto = (saved: SavedDpsCurveSettings): DpsCurveValues => {
	const stats = saved.stats.map(stat => UnitStat.fromProto(stat)).filter(stat => stat.isStat() || stat.isPseudoStat());
	return {
		stats: stats.length ? stats : DEFAULTS.stats,
		range: saved.range || DEFAULTS.range,
		step: saved.step || DEFAULTS.step,
		precisionMode: saved.precisionMode === SavedPrecisionMode.DpsCurvePrecisionModeTarget ? 'target' : 'iterations',
		iterations: saved.iterations,
		targetDpsCi: saved.targetDpsCi || DEFAULTS.targetDpsCi,
		maxIterationMultiplier: saved.maxIterationMultiplier,
	};
};

const sameStats = (a: UnitStat[] | null, b: UnitStat[] | null) => (a && b ? a.length === b.length && a.every((stat, index) => stat.equals(b[index])) : a === b);

// DPS curve settings. Values live in the sim store (`dpsCurve[player.storeKey]`) with a version counter; persists itself
// to localStorage on every change.
export class DpsCurveSettings {
	private readonly storageKey: string;
	private readonly env: Env;
	readonly store: SimStore;
	readonly storeKey: number;

	constructor(player: Player<any>, storageKey: string) {
		this.storageKey = storageKey;
		this.env = player.sim.env;
		this.store = player.sim.store;
		this.storeKey = player.storeKey;

		seedKeyed(this.store, 'dpsCurve', this.storeKey, { ...DEFAULTS, v: { settings: 0 } });
		subscribeDpsCurveChange(this)(() => {
			this.env.storage.setItem(this.storageKey, SavedDpsCurveSettings.toJsonString(this.toProto()));
		});
	}

	private get slice(): DpsCurveSlice {
		return this.store.getState().dpsCurve[this.storeKey];
	}

	private write(patch: Partial<DpsCurveValues>) {
		patchKeyed(this.store, 'dpsCurve', this.storeKey, patch, ['settings']);
	}

	private update(patch: Partial<DpsCurveValues>) {
		const current = this.slice;
		const unchanged = (Object.keys(patch) as Array<keyof DpsCurveValues>).every(key =>
			key === 'stats' ? sameStats(patch.stats ?? null, current.stats) : patch[key] === current[key],
		);
		if (!unchanged) this.write(patch);
	}

	get setup(): DpsCurveSetup {
		const { stats, range, step } = this.slice;
		return { stats, range, step };
	}

	setSetup(patch: Partial<DpsCurveSetup>) {
		this.update(patch);
	}

	get precision(): DpsCurvePrecision {
		const { precisionMode, iterations, targetDpsCi, maxIterationMultiplier } = this.slice;
		return { precisionMode, iterations, targetDpsCi, maxIterationMultiplier };
	}

	setPrecision(patch: Partial<DpsCurvePrecision>) {
		this.update(patch);
	}

	applyDefaults() {
		this.write(DEFAULTS);
	}

	load() {
		const storageValue = this.env.storage.getItem(this.storageKey);
		if (storageValue) {
			this.write(fromProto(SavedDpsCurveSettings.fromJsonString(storageValue, { ignoreUnknownFields: true })));
		}
	}

	toProto(): SavedDpsCurveSettings {
		const { stats, range, step, precisionMode, iterations, targetDpsCi, maxIterationMultiplier } = this.slice;
		return SavedDpsCurveSettings.create({
			apiVersion: CURRENT_API_VERSION,
			stats: stats?.map(stat => stat.toProto()) ?? [],
			range,
			step,
			precisionMode: precisionMode === 'target' ? SavedPrecisionMode.DpsCurvePrecisionModeTarget : SavedPrecisionMode.DpsCurvePrecisionModeIterations,
			iterations,
			targetDpsCi,
			maxIterationMultiplier,
		});
	}
}
