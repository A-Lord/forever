import type { ErrorOutcome } from '@generated/proto/api';
import { Class, Spec, Stat } from '@generated/proto/common';
import type { MetricSummary, StatSweepPoint } from '@sim/cache/stat_sweep_history';
import type { Stats } from '@sim/proto/stats';
import { StatSweepSettings } from '@sim/settings/stat_sweep_settings';
import type { MetricSample, StatSweepSample, StatSweepStepOptions } from '@sim/sim';
import type { IndividualSimHost } from '@sim/sim_host';
import { SimRuns } from '@sim/sim_runs';
import { createSimStore, patchSlice, PLAYER_FIELDS, seedKeyed, type SimStore, type StatSweepMetric, zeroVersions } from '@sim/state/sim_store';

import type { StatSweepSeries } from '../../../features/stat-sweep/components/StatSweepChart';
import { buildSweepOffsets, sweepSteps } from '../../../features/stat-sweep/model/stat_sweep';

const HIT_CAP = 150;
const DPS_PER_POINT = 2.1;

// DPS climbs until the cap and then flattens.
const dpsAt = (value: number) => 1180 + DPS_PER_POINT * Math.min(value, HIT_CAP) - 0.004 * (Math.min(value, HIT_CAP) - 80) ** 2;

type MetricsAt = (value: number, samples: number) => Record<StatSweepMetric, MetricSummary>;

const pointsFrom = (base: number, samples: number, metricsAt: MetricsAt): StatSweepPoint[] => {
	const offsets = buildSweepOffsets(100, 10);
	return sweepSteps(
		offsets,
		offsets.map(offset => base + offset),
	).map(({ offset, statValue }) => ({ offset, statValue, metrics: metricsAt(statValue, samples) }));
};

// A damage dealer: threat follows damage, and nothing is healed or taken.
const dealerMetrics =
	(dps: (value: number) => number): MetricsAt =>
	(value, n) => ({
		dps: { avg: dps(value), stdev: 450, n },
		hps: { avg: 0, stdev: 0, n },
		tps: { avg: dps(value) * 0.71, stdev: 320, n },
		dtps: { avg: 0, stdev: 0, n },
		tmi: { avg: 0, stdev: 0, n },
	});

// A tank whose damage taken and TMI fall as the stat rises.
const tankMetrics: MetricsAt = (value, n) => ({
	dps: { avg: 640 + 0.2 * value, stdev: 90, n },
	hps: { avg: 0, stdev: 0, n },
	tps: { avg: 1450 + 0.35 * value, stdev: 160, n },
	dtps: { avg: 1250 - 1.6 * value, stdev: 210, n },
	tmi: { avg: 18 - 0.03 * value, stdev: 4.5, n },
});

/** A ±100 hit sweep in steps of 10 from `base`, against a hit cap at 150. */
export const makeHitSweepPoints = (base: number, samples = 20_000): StatSweepPoint[] => pointsFrom(base, samples, dealerMetrics(dpsAt));

/** A ±100 crit sweep in steps of 10 from `base`, with no cap in reach. */
export const makeCritSweepPoints = (base: number, samples = 20_000): StatSweepPoint[] =>
	pointsFrom(
		base,
		samples,
		dealerMetrics(value => 1300 + 1.1 * value),
	);

/** A ±100 defense sweep in steps of 10 from `base`, for a tank. */
export const makeDefenseSweepPoints = (base: number, samples = 20_000): StatSweepPoint[] => pointsFrom(base, samples, tankMetrics);

export const sweepSeries = (
	id: string,
	statLabel: string,
	base: number,
	points: StatSweepPoint[],
	colorSlot = 0,
	label = `${statLabel} ±100`,
): StatSweepSeries => ({ id, label, colorSlot, statLabel, baseStatValue: base, points });

const flatSample = (avg: number, stdev: number, n: number): MetricSample => ({ avg, stdev, n, sumSq: n * (stdev ** 2 + avg ** 2) });

/** One sweep step's metrics over `iterations`, each moving with `dps`; TPS has the widest spread. */
export const makeSweepSample = (iterations: number, dps = 1000, tmi = dps / 100): StatSweepSample => ({
	dps: flatSample(dps, 100, iterations),
	hps: flatSample(10, 5, iterations),
	tps: flatSample(2 * dps, 400, iterations),
	dtps: flatSample(3000 - dps, 300, iterations),
	tmi: flatSample(tmi, 2, iterations),
});

/** One DPS per bonus point of Strength or Agility, so every step's DPS differs. */
export const sweepStepDps = ({ bonusStats }: StatSweepStepOptions) => 1000 + bonusStats.getStat(Stat.StatStrength) + bonusStats.getStat(Stat.StatAgility);

export interface StatSweepHostParts {
	simulateStatSweepStep?: (options: StatSweepStepOptions) => Promise<StatSweepSample | ErrorOutcome>;
	abortType?: (typeMask: number) => Promise<void>;
	simDisabled?: boolean;
	showPowerTools?: boolean;
}

export interface StatSweepHostFixture {
	host: IndividualSimHost<any>;
	store: SimStore;
	settings: StatSweepSettings;
}

/**
 * A host for the stat sweep: a warrior weighing Strength and Agility with no stats of its own, so a step's totals are its
 * bonus, on a sim that draws seed 7 and runs `simulateStatSweepStep` for each pass, by default at `sweepStepDps`.
 */
export const makeStatSweepHost = ({
	simulateStatSweepStep,
	abortType,
	simDisabled = false,
	showPowerTools = false,
}: StatSweepHostParts = {}): StatSweepHostFixture => {
	const store = createSimStore();
	patchSlice(store, 'ui', { showPowerTools });
	const storage = new Map<string, string>();
	const env = { storage: { getItem: (key: string) => storage.get(key) ?? null, setItem: (key: string, value: string) => storage.set(key, value) } };
	seedKeyed(store, 'players', 0, { v: zeroVersions(PLAYER_FIELDS) } as never);
	const sim = {
		store,
		env,
		runs: new SimRuns(store, { abortType: abortType ?? (async () => {}) }),
		waitForInit: () => Promise.resolve(),
		getIterations: () => 1000,
		nextRngSeed: () => 7,
		getShowDamageMetrics: () => store.getState().ui.showDamageMetrics,
		getShowHealingMetrics: () => store.getState().ui.showHealingMetrics,
		getShowThreatMetrics: () => store.getState().ui.showThreatMetrics,
		getFinalStatsWithBonus: async (bonusStats: Stats) => bonusStats,
		simulateStatSweepStep: simulateStatSweepStep ?? (async (options: StatSweepStepOptions) => makeSweepSample(options.iterations, sweepStepDps(options))),
	};
	const player = { storeKey: 0, sim, getClass: () => Class.ClassWarrior, getSpec: () => Spec.SpecUnknown };
	const settings = new StatSweepSettings(player as never, 'stat-sweep-fixture');
	const host = {
		player,
		sim,
		simDisabled,
		individualConfig: { epStats: [Stat.StatStrength, Stat.StatAgility], epPseudoStats: [], epReferenceStat: Stat.StatStrength },
	};
	return { host: host as never, store, settings };
};
