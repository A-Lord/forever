import { PseudoStat, Stat } from '@generated/proto/common';
import { EpPrecisionMode as SavedEpPrecisionMode, SavedStatWeightSettings } from '@generated/proto/ui';

import { CURRENT_API_VERSION } from '../constants/other';
import type { Player } from '../player/player';
import { UnitStat } from '../proto/stats';
import type { Env } from '../state/env';
import { patchKeyed, seedKeyed, SimStore, StatWeightsSlice } from '../state/sim_store';
import { subscribeStatWeightsChange } from '../state/subscriptions';

export type EpPrecisionSettings = Pick<StatWeightsSlice, 'precisionMode' | 'iterations' | 'targetEpCi' | 'maxIterationMultiplier'>;

export type StatSweepSettings = Pick<StatWeightsSlice, 'sweepStats' | 'sweepRange' | 'sweepStep'>;

const PRECISION_DEFAULTS: EpPrecisionSettings = { precisionMode: 'iterations', iterations: 0, targetEpCi: 0.01, maxIterationMultiplier: 0 };
const SWEEP_DEFAULTS: StatSweepSettings = { sweepStats: null, sweepRange: 100, sweepStep: 10 };

// Proto3 cannot tell an unset field from 0, so a zero target, sweep range or sweep step reads as the default.
const precisionFromProto = (saved: SavedStatWeightSettings): EpPrecisionSettings => ({
	precisionMode: saved.precisionMode === SavedEpPrecisionMode.EpPrecisionModeTarget ? 'target' : 'iterations',
	iterations: saved.iterations,
	targetEpCi: saved.targetEpCi || PRECISION_DEFAULTS.targetEpCi,
	maxIterationMultiplier: saved.maxIterationMultiplier,
});

// Proto3 cannot tell an unset list from an empty one either, so no stored stat reads as the default.
const sweepStatsFromProto = (saved: SavedStatWeightSettings): UnitStat[] | null => {
	const stats = saved.sweepStats.map(stat => UnitStat.fromProto(stat)).filter(stat => stat.isStat() || stat.isPseudoStat());
	return stats.length ? stats : null;
};

const sweepFromProto = (saved: SavedStatWeightSettings): StatSweepSettings => ({
	sweepStats: sweepStatsFromProto(saved),
	sweepRange: saved.sweepRange || SWEEP_DEFAULTS.sweepRange,
	sweepStep: saved.sweepStep || SWEEP_DEFAULTS.sweepStep,
});

const sameSweepStats = (a: UnitStat[] | null, b: UnitStat[] | null) =>
	a && b ? a.length === b.length && a.every((stat, index) => stat.equals(b[index])) : a === b;

// Stat-weight modal settings. Values live in the sim store
// (`statWeights[player.storeKey]`) with a version counter; persists itself to
// localStorage on every change.
export class StatWeightActionSettings {
	private readonly storageKey: string;
	private readonly env: Env;
	readonly store: SimStore;
	readonly storeKey: number;

	constructor(player: Player<any>, storageKey: string) {
		this.storageKey = storageKey;
		this.env = player.sim.env;
		this.store = player.sim.store;
		this.storeKey = player.storeKey;

		seedKeyed(this.store, 'statWeights', this.storeKey, {
			excludedStats: [],
			excludedPseudoStats: [],
			...PRECISION_DEFAULTS,
			...SWEEP_DEFAULTS,
			v: { settings: 0 },
		});
		subscribeStatWeightsChange(this)(() => {
			this.env.storage.setItem(this.storageKey, SavedStatWeightSettings.toJsonString(this.toProto()));
		});
	}

	private get slice(): StatWeightsSlice {
		return this.store.getState().statWeights[this.storeKey];
	}

	// One store write per logical change.
	private write(patch: Partial<Omit<StatWeightsSlice, 'v'>>) {
		patchKeyed(this.store, 'statWeights', this.storeKey, patch, ['settings']);
	}

	get excludedStats(): Stat[] {
		return this.slice.excludedStats.slice();
	}

	get excludedPseudoStats(): PseudoStat[] {
		return this.slice.excludedPseudoStats.slice();
	}

	get precision(): EpPrecisionSettings {
		const { precisionMode, iterations, targetEpCi, maxIterationMultiplier } = this.slice;
		return { precisionMode, iterations, targetEpCi, maxIterationMultiplier };
	}

	setPrecision(patch: Partial<EpPrecisionSettings>) {
		const current = this.precision;
		if ((Object.keys(patch) as Array<keyof EpPrecisionSettings>).every(key => patch[key] === current[key])) return;
		this.write(patch);
	}

	get sweep(): StatSweepSettings {
		const { sweepStats, sweepRange, sweepStep } = this.slice;
		return { sweepStats, sweepRange, sweepStep };
	}

	setSweep(patch: Partial<StatSweepSettings>) {
		const current = this.sweep;
		const unchanged = (Object.keys(patch) as Array<keyof StatSweepSettings>).every(key =>
			key === 'sweepStats' ? sameSweepStats(patch.sweepStats ?? null, current.sweepStats) : patch[key] === current[key],
		);
		if (unchanged) return;
		this.write(patch);
	}

	applyDefaults() {
		this.write({ excludedStats: [], excludedPseudoStats: [], ...PRECISION_DEFAULTS, ...SWEEP_DEFAULTS });
	}

	load() {
		const storageValue = this.env.storage.getItem(this.storageKey);
		if (storageValue) {
			const settingsProto = SavedStatWeightSettings.fromJsonString(storageValue, { ignoreUnknownFields: true });
			this.write({
				excludedStats: settingsProto.excludedStats || [],
				excludedPseudoStats: settingsProto.excludedPseudoStats || [],
				...precisionFromProto(settingsProto),
				...sweepFromProto(settingsProto),
			});
		}
	}

	toProto(): SavedStatWeightSettings {
		const { precisionMode, iterations, targetEpCi, maxIterationMultiplier } = this.precision;
		const { sweepStats, sweepRange, sweepStep } = this.sweep;
		return SavedStatWeightSettings.create({
			apiVersion: CURRENT_API_VERSION,
			excludedStats: this.excludedStats,
			excludedPseudoStats: this.excludedPseudoStats,
			precisionMode: precisionMode === 'target' ? SavedEpPrecisionMode.EpPrecisionModeTarget : SavedEpPrecisionMode.EpPrecisionModeIterations,
			iterations,
			targetEpCi,
			maxIterationMultiplier,
			sweepStats: sweepStats?.map(stat => stat.toProto()) ?? [],
			sweepRange,
			sweepStep,
		});
	}

	/**
	 * Check if a stat should be excluded from weight calculation.
	 * @param stat
	 * @returns true if stat should be excluded.
	 */
	isStatExcludedFromCalc(stat: Stat): boolean {
		return !!this.excludedStats.includes(stat);
	}

	/**
	 * Check if a pseudostat should be excluded from weight calculation.
	 * @param pseudoStat
	 * @returns true if pseudostat should be excluded.
	 */
	isPseudoStatExcludedFromCalc(pseudoStat: PseudoStat): boolean {
		return !!this.excludedPseudoStats.includes(pseudoStat);
	}

	/**
	 * Check if a unitstat should be excluded from weight calculation.
	 * @param unitstat
	 * @returns true if unitstat should be excluded.
	 */
	isUnitStatExcludedFromCalc(unitstat: UnitStat): boolean {
		return unitstat.isStat() ? this.isStatExcludedFromCalc(unitstat.getStat()) : this.isPseudoStatExcludedFromCalc(unitstat.getPseudoStat());
	}

	/**
	 * Set whether a stat should be excluded from calculation.
	 * @param stat
	 * @param exclude
	 */
	setStatExcluded(stat: UnitStat, exclude: boolean) {
		const updateStatEntry = <T extends Stat | PseudoStat>(s: T, target: T[]) => {
			const currentIdx = target.indexOf(s);
			if (exclude) {
				if (currentIdx === -1) target.push(s);
			} else if (currentIdx !== -1) {
				target.splice(currentIdx, 1);
			}
			return target;
		};
		if (stat.isStat()) {
			this.write({ excludedStats: updateStatEntry(stat.getStat(), this.excludedStats) });
		} else {
			this.write({ excludedPseudoStats: updateStatEntry(stat.getPseudoStat(), this.excludedPseudoStats) });
		}
	}
}
