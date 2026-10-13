import type { MetricSummary, StatSweepPoint } from '@sim/cache/stat_sweep_history';
import type { StatSweepMetric } from '@sim/state/sim_store';

import type { StatSweepSeries } from '../../../features/stat-sweep/components/StatSweepDialog/StatSweepChart';
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
