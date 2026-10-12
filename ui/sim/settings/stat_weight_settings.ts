import { PseudoStat, Stat } from '@generated/proto/common';
import { SavedStatWeightSettings } from '@generated/proto/ui';

import { CURRENT_API_VERSION } from '../constants/other';
import type { Player } from '../player/player';
import { UnitStat } from '../proto/stats';
import type { Env } from '../state/env';
import { EpPrecisionMode, patchKeyed, seedKeyed, SimStore, StatWeightsSlice } from '../state/sim_store';
import { subscribeStatWeightsChange } from '../state/subscriptions';

export type EpPrecisionSettings = Pick<StatWeightsSlice, 'precisionMode' | 'iterations' | 'targetEpCi' | 'maxIterationMultiplier'>;

const PRECISION_DEFAULTS: EpPrecisionSettings = { precisionMode: 'iterations', iterations: 0, targetEpCi: 0.01, maxIterationMultiplier: 8 };
const PRECISION_MODES: ReadonlyArray<EpPrecisionMode> = ['iterations', 'target'];

const readNumber = (value: unknown, fallback: number, min: number) => (typeof value === 'number' && Number.isFinite(value) && value >= min ? value : fallback);

const readPrecision = (saved: Record<string, unknown>): EpPrecisionSettings => ({
	precisionMode: PRECISION_MODES.find(mode => mode === saved.precisionMode) ?? PRECISION_DEFAULTS.precisionMode,
	iterations: readNumber(saved.iterations, PRECISION_DEFAULTS.iterations, 0),
	targetEpCi: readNumber(saved.targetEpCi, PRECISION_DEFAULTS.targetEpCi, Number.MIN_VALUE),
	maxIterationMultiplier: readNumber(saved.maxIterationMultiplier, PRECISION_DEFAULTS.maxIterationMultiplier, 1),
});

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

		seedKeyed(this.store, 'statWeights', this.storeKey, { excludedStats: [], excludedPseudoStats: [], ...PRECISION_DEFAULTS, v: { settings: 0 } });
		subscribeStatWeightsChange(this)(() => {
			// SavedStatWeightSettings has no precision fields, so they are stored beside its JSON keys; load() reads the proto with ignoreUnknownFields.
			this.env.storage.setItem(this.storageKey, JSON.stringify({ ...(SavedStatWeightSettings.toJson(this.toProto()) as object), ...this.precision }));
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

	applyDefaults() {
		this.write({ excludedStats: [], excludedPseudoStats: [], ...PRECISION_DEFAULTS });
	}

	load() {
		const storageValue = this.env.storage.getItem(this.storageKey);
		if (storageValue) {
			const settingsProto = SavedStatWeightSettings.fromJsonString(storageValue, { ignoreUnknownFields: true });
			this.write({
				excludedStats: settingsProto.excludedStats || [],
				excludedPseudoStats: settingsProto.excludedPseudoStats || [],
				...readPrecision(JSON.parse(storageValue)),
			});
		}
	}

	toProto(): SavedStatWeightSettings {
		return SavedStatWeightSettings.create({
			apiVersion: CURRENT_API_VERSION,
			excludedStats: this.excludedStats,
			excludedPseudoStats: this.excludedPseudoStats,
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
