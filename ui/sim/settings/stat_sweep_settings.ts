import { SavedStatSweepSettings, StatSweepMetric as SavedMetric, StatSweepPrecisionMode as SavedPrecisionMode } from '@generated/proto/ui';

import { CURRENT_API_VERSION } from '../constants/other';
import type { Player } from '../player/player';
import { UnitStat } from '../proto/stats';
import type { Env } from '../state/env';
import { patchKeyed, seedKeyed, type SimStore, type StatSweepMetric, type StatSweepSlice } from '../state/sim_store';
import { subscribeStatSweepChange } from '../state/subscriptions';

export type StatSweepSetup = Pick<StatSweepSlice, 'stats' | 'range' | 'step'>;

export type StatSweepPrecision = Pick<StatSweepSlice, 'precisionMode' | 'iterations' | 'targetCi' | 'maxIterationMultiplier'>;

type StatSweepValues = Omit<StatSweepSlice, 'v'>;

const DEFAULTS: StatSweepValues = {
	stats: null,
	range: 100,
	step: 10,
	metric: 'dps',
	precisionMode: 'iterations',
	iterations: 0,
	targetCi: 1,
	maxIterationMultiplier: 0,
};

const SAVED_METRICS: Record<StatSweepMetric, SavedMetric> = {
	dps: SavedMetric.StatSweepMetricDps,
	hps: SavedMetric.StatSweepMetricHps,
	tps: SavedMetric.StatSweepMetricTps,
	dtps: SavedMetric.StatSweepMetricDtps,
	tmi: SavedMetric.StatSweepMetricTmi,
};

const metricFromProto = (saved: SavedMetric): StatSweepMetric =>
	(Object.keys(SAVED_METRICS) as StatSweepMetric[]).find(metric => SAVED_METRICS[metric] === saved) ?? DEFAULTS.metric;

// Proto3 cannot tell an unset field from 0, or an unset list from an empty one, so a zero range, step or target and
// no stored stat read as the default.
const fromProto = (saved: SavedStatSweepSettings): StatSweepValues => {
	const stats = saved.stats.map(stat => UnitStat.fromProto(stat)).filter(stat => stat.isStat() || stat.isPseudoStat());
	return {
		stats: stats.length ? stats : DEFAULTS.stats,
		range: saved.range || DEFAULTS.range,
		step: saved.step || DEFAULTS.step,
		metric: metricFromProto(saved.metric),
		precisionMode: saved.precisionMode === SavedPrecisionMode.StatSweepPrecisionModeTarget ? 'target' : 'iterations',
		iterations: saved.iterations,
		targetCi: saved.targetCi || DEFAULTS.targetCi,
		maxIterationMultiplier: saved.maxIterationMultiplier,
	};
};

const sameStats = (a: UnitStat[] | null, b: UnitStat[] | null) => (a && b ? a.length === b.length && a.every((stat, index) => stat.equals(b[index])) : a === b);

// Stat sweep settings. Values live in the sim store (`statSweep[player.storeKey]`) with a version counter; persists itself
// to localStorage on every change.
export class StatSweepSettings {
	private readonly storageKey: string;
	private readonly env: Env;
	readonly store: SimStore;
	readonly storeKey: number;

	constructor(player: Player<any>, storageKey: string) {
		this.storageKey = storageKey;
		this.env = player.sim.env;
		this.store = player.sim.store;
		this.storeKey = player.storeKey;

		seedKeyed(this.store, 'statSweep', this.storeKey, { ...DEFAULTS, v: { settings: 0 } });
		subscribeStatSweepChange(this)(() => {
			this.env.storage.setItem(this.storageKey, SavedStatSweepSettings.toJsonString(this.toProto()));
		});
	}

	private get slice(): StatSweepSlice {
		return this.store.getState().statSweep[this.storeKey];
	}

	private write(patch: Partial<StatSweepValues>) {
		patchKeyed(this.store, 'statSweep', this.storeKey, patch, ['settings']);
	}

	private update(patch: Partial<StatSweepValues>) {
		const current = this.slice;
		const unchanged = (Object.keys(patch) as Array<keyof StatSweepValues>).every(key =>
			key === 'stats' ? sameStats(patch.stats ?? null, current.stats) : patch[key] === current[key],
		);
		if (!unchanged) this.write(patch);
	}

	get setup(): StatSweepSetup {
		const { stats, range, step } = this.slice;
		return { stats, range, step };
	}

	setSetup(patch: Partial<StatSweepSetup>) {
		this.update(patch);
	}

	/** The stored pick, which the dialog plots only while the display settings offer it. */
	get metric(): StatSweepMetric {
		return this.slice.metric;
	}

	setMetric(metric: StatSweepMetric) {
		this.update({ metric });
	}

	get precision(): StatSweepPrecision {
		const { precisionMode, iterations, targetCi, maxIterationMultiplier } = this.slice;
		return { precisionMode, iterations, targetCi, maxIterationMultiplier };
	}

	setPrecision(patch: Partial<StatSweepPrecision>) {
		this.update(patch);
	}

	applyDefaults() {
		this.write(DEFAULTS);
	}

	load() {
		const storageValue = this.env.storage.getItem(this.storageKey);
		if (storageValue) {
			this.write(fromProto(SavedStatSweepSettings.fromJsonString(storageValue, { ignoreUnknownFields: true })));
		}
	}

	toProto(): SavedStatSweepSettings {
		const { stats, range, step, metric, precisionMode, iterations, targetCi, maxIterationMultiplier } = this.slice;
		return SavedStatSweepSettings.create({
			apiVersion: CURRENT_API_VERSION,
			stats: stats?.map(stat => stat.toProto()) ?? [],
			range,
			step,
			metric: SAVED_METRICS[metric],
			precisionMode: precisionMode === 'target' ? SavedPrecisionMode.StatSweepPrecisionModeTarget : SavedPrecisionMode.StatSweepPrecisionModeIterations,
			iterations,
			targetCi,
			maxIterationMultiplier,
		});
	}
}
