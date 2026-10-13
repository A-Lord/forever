import type { MetricSummary, StatSweepPoint, StatSweepRecord } from '@sim/cache/stat_sweep_history';
import type { StatSweepPrecision } from '@sim/settings/stat_sweep_settings';
import type { MetricSample, StatSweepSample } from '@sim/sim';
import type { StatSweepMetric } from '@sim/state/sim_store';
import { distinct } from '@sim/utils/collections';
import { sum, timeRemainingAtRate } from '@sim/utils/math';

export type StatSweepStep = Pick<StatSweepPoint, 'offset' | 'statValue'>;

/** Every metric a step records, in the order the toggle and the CSV list them. */
export const STAT_SWEEP_METRICS: ReadonlyArray<StatSweepMetric> = ['dps', 'hps', 'tps', 'dtps', 'tmi'];

/** DPS and TPS always, HPS with the healing metrics shown, DTPS and TMI with the threat and tank options on. */
export const offeredSweepMetrics = ({ healing, threat }: { healing: boolean; threat: boolean }): StatSweepMetric[] =>
	STAT_SWEEP_METRICS.filter(metric => metric === 'dps' || metric === 'tps' || (metric === 'hps' ? healing : threat));

/** The stored metric while it is offered, else DPS. */
export const shownSweepMetric = (stored: StatSweepMetric, offered: ReadonlyArray<StatSweepMetric>): StatSweepMetric =>
	offered.includes(stored) ? stored : 'dps';

// Rounds away float noise like 3 × 0.1 = 0.30000000000000004, which would otherwise label an axis tick.
const roundFloat = (value: number) => Math.round(value * 1e6) / 1e6;

/**
 * Multiples of `step` from −`range` to +`range`, ascending, always with 0. ±`range` itself is left out when `range` is
 * not a multiple of `step`.
 */
export function buildSweepOffsets(range: number, step: number): number[] {
	if (!(step > 0) || !(range > 0)) return [0];
	const offsets = [0];
	const count = Math.floor(roundFloat(range / step));
	for (let k = 1; k <= count; k++) {
		const offset = roundFloat(k * step);
		offsets.push(offset, -offset);
	}
	return offsets.sort((a, b) => a - b);
}

/** Each offset at the stat total it reaches, `totals` in the same order; an offset whose total falls below 0 is dropped. */
export const sweepSteps = (offsets: readonly number[], totals: readonly number[]): StatSweepStep[] =>
	offsets.map((offset, index) => ({ offset, statValue: roundFloat(totals[index]) })).filter(step => step.statValue >= 0);

/** The stored runs plus any run in memory they do not hold yet, newest first. */
export const mergeSweepHistory = (inMemory: readonly StatSweepRecord[], stored: readonly StatSweepRecord[]): StatSweepRecord[] =>
	distinct([...stored, ...inMemory], (a, b) => a.id === b.id).sort((a, b) => b.createdAt - a.createdAt);

/** A step's first pass in Target ± mode; the iterations setting plays no part there. */
export const STAT_SWEEP_PROBE_ITERATIONS = 1000;
/** Passes after the first in Target ± mode, each sized from every iteration so far. */
export const STAT_SWEEP_MAX_TOP_UP_PASSES = 2;
const Z90 = 1.645;

/** Two runs' metric as one run over both runs' iterations reports it. */
export function mergeMetricSamples(a: MetricSample, b: MetricSample): MetricSample {
	const n = a.n + b.n;
	const avg = (a.avg * a.n + b.avg * b.n) / n;
	const sumSq = a.sumSq + b.sumSq;
	return { avg, stdev: Math.sqrt(Math.max(0, sumSq / n - avg * avg)), n, sumSq };
}

/** Two runs' metrics, each merged as one run over both runs' iterations reports it. */
export const mergeSweepSamples = (a: StatSweepSample, b: StatSweepSample): StatSweepSample =>
	Object.fromEntries(STAT_SWEEP_METRICS.map(metric => [metric, mergeMetricSamples(a[metric], b[metric])])) as StatSweepSample;

/** The iterations at which a metric spread of `stdev` has a 90% ± of `targetCi`. */
export const iterationsForCi = (stdev: number, targetCi: number) => Math.ceil(((Z90 * stdev) / targetCi) ** 2);

/** Runs `iterations` more from `seedOffset` past the run's seed; null when the run stopped. */
export type SweepPassRunner = (iterations: number, seedOffset: number) => Promise<StatSweepSample | null>;

/** A step's metrics over all its passes, with each metric's average over its first pass, which every step runs at the same seed and size. */
export interface StatSweepStepSample {
	metrics: StatSweepSample;
	firstPass: Record<StatSweepMetric, number>;
}

const averages = (sample: StatSweepSample): Record<StatSweepMetric, number> =>
	Object.fromEntries(STAT_SWEEP_METRICS.map(metric => [metric, sample[metric].avg])) as Record<StatSweepMetric, number>;

/**
 * One step's metrics. Iterations mode runs once. Target ± mode runs a fixed first pass, then up to
 * STAT_SWEEP_MAX_TOP_UP_PASSES more, each up to the total the ± of `metric` needs and capped at Max × the first pass. A pass
 * continues at the seed after the last iteration, so the passes add up to one longer run. `onPlanned` hears each total
 * before its pass.
 */
export async function measureSweepStep(
	run: SweepPassRunner,
	precision: StatSweepPrecision,
	metric: StatSweepMetric,
	defaultIterations: number,
	onPlanned?: (total: number) => void,
): Promise<StatSweepStepSample | null> {
	if (precision.precisionMode !== 'target') {
		const iterations = precision.iterations || defaultIterations;
		onPlanned?.(iterations);
		const sample = await run(iterations, 0);
		return sample && { metrics: sample, firstPass: averages(sample) };
	}
	onPlanned?.(STAT_SWEEP_PROBE_ITERATIONS);
	let sample = await run(STAT_SWEEP_PROBE_ITERATIONS, 0);
	if (!sample) return null;
	const firstPass = averages(sample);
	const cap = precision.maxIterationMultiplier > 0 ? STAT_SWEEP_PROBE_ITERATIONS * precision.maxIterationMultiplier : Infinity;
	for (let pass = 0; sample && pass < STAT_SWEEP_MAX_TOP_UP_PASSES; pass++) {
		const done = sample[metric].n;
		const total = Math.min(iterationsForCi(sample[metric].stdev, precision.targetCi), cap);
		if (total <= done) break;
		onPlanned?.(total);
		const more = await run(total - done, done);
		sample = more && mergeSweepSamples(sample, more);
	}
	return sample && { metrics: sample, firstPass };
}

const summaries = (sample: StatSweepSample): Record<StatSweepMetric, MetricSummary> =>
	Object.fromEntries(STAT_SWEEP_METRICS.map(metric => [metric, { avg: sample[metric].avg, stdev: sample[metric].stdev, n: sample[metric].n }])) as Record<
		StatSweepMetric,
		MetricSummary
	>;

/** `steps` in the order a stat runs them: up from change 0 to the top, then down from the first step below 0. */
export const outwardOrder = (steps: readonly StatSweepStep[]): StatSweepStep[] => [
	...steps.filter(step => step.offset >= 0).sort((a, b) => a.offset - b.offset),
	...steps.filter(step => step.offset < 0).sort((a, b) => b.offset - a.offset),
];

export interface StatSweepStatRun {
	/** The finished steps, sorted by change. */
	points: StatSweepPoint[];
	/** The highest step that changed a metric: the step above it gave the same value of every metric, so more of the stat does nothing. */
	cappedStep?: StatSweepStep;
	/** Set when a step came back with no result, which ends the run. */
	stopped: boolean;
}

/**
 * Runs one stat's steps in `outwardOrder`. Every step runs at the same seed, so a stat that no longer changes anything
 * gives the exact first-pass average of every metric of the step below it: that step is kept and the steps above it are skipped, with
 * `onSkip` told how many. The steps below 0 always all run. `measure` hears each step with the count of steps run before
 * it, and `onUpdate` hears the run after each step.
 */
export async function runSweepStat(
	steps: readonly StatSweepStep[],
	measure: (step: StatSweepStep, ran: number) => Promise<StatSweepStepSample | null>,
	onUpdate?: (run: StatSweepStatRun) => void,
	onSkip?: (count: number) => void,
): Promise<StatSweepStatRun> {
	const ordered = outwardOrder(steps);
	const run: StatSweepStatRun = { points: [], stopped: false };
	let below: { step: StatSweepStep; firstPass: Record<StatSweepMetric, number> } | undefined;
	for (const [index, step] of ordered.entries()) {
		if (run.cappedStep && step.offset > 0) continue;
		const sample = await measure(step, run.points.length);
		if (!sample) return { ...run, stopped: true };
		run.points = [...run.points, { ...step, metrics: summaries(sample.metrics) }].sort((a, b) => a.offset - b.offset);
		if (step.offset >= 0) {
			if (step.offset > 0 && below && STAT_SWEEP_METRICS.every(metric => below!.firstPass[metric] === sample.firstPass[metric])) {
				run.cappedStep = below.step;
				const skipped = ordered.slice(index + 1).filter(next => next.offset > 0).length;
				if (skipped) onSkip?.(skipped);
			}
			below = { step, firstPass: sample.firstPass };
		}
		onUpdate?.({ ...run });
	}
	return run;
}

export interface StatSweepProgress {
	/** 1-based stat being run. */
	stat: number;
	stats: number;
	/** 1-based step of that stat being simulated. */
	step: number;
	/** That stat's steps. */
	steps: number;
	/** Steps done across every stat, with the running step's done share; never falls. */
	completed: number;
	/** Steps across every stat. */
	totalSteps: number;
	/** Each stat's steps, in run order. */
	stepCounts: number[];
	/** The running step's expected iterations, 0 until one is known. */
	iterationsPerStep: number;
	secondsRemaining?: number;
}

export const startSweepProgress = (stepCounts: number[]): StatSweepProgress => ({
	stat: 1,
	stats: stepCounts.length,
	step: 1,
	steps: stepCounts[0] ?? 0,
	completed: 0,
	totalSteps: sum(stepCounts),
	stepCounts,
	iterationsPerStep: 0,
});

/**
 * `progress` with step `index` (0-based, counted across every stat) `done` iterations into the `expected` it is likely
 * to run, `elapsedSeconds` into the run.
 */
export function advanceSweepProgress(progress: StatSweepProgress, index: number, done: number, expected: number, elapsedSeconds: number): StatSweepProgress {
	const { stepCounts, totalSteps } = progress;
	let stat = 0;
	let step = index;
	while (stat < stepCounts.length - 1 && step >= stepCounts[stat]) step -= stepCounts[stat++];
	const share = expected > 0 ? Math.min(1, done / expected) : 0;
	const completed = Math.max(progress.completed, index + share);
	return {
		...progress,
		stat: stat + 1,
		step: step + 1,
		steps: stepCounts[stat],
		completed,
		iterationsPerStep: expected || progress.iterationsPerStep,
		secondsRemaining: timeRemainingAtRate(elapsedSeconds, completed, totalSteps) ?? progress.secondsRemaining,
	};
}

/** `progress` with `count` steps of stat `statIndex` (0-based) that will not run taken out of its step count and the total. */
export function skipSweepSteps(progress: StatSweepProgress, statIndex: number, count: number): StatSweepProgress {
	const stepCounts = progress.stepCounts.map((steps, index) => (index === statIndex ? steps - count : steps));
	return { ...progress, stepCounts, steps: stepCounts[progress.stat - 1], totalSteps: progress.totalSteps - count };
}

/** The slot a new run takes: the one after the newest saved run's, wrapping at `slotCount`. */
export const nextColorSlot = (history: readonly Pick<StatSweepRecord, 'colorSlot'>[], slotCount: number): number =>
	history.length ? (history[0].colorSlot + 1) % slotCount : 0;

/**
 * The slot each shown run is drawn in, given oldest first. A run keeps its own slot unless an earlier run holds it, which
 * happens once the slots wrap around; it then takes the next free one.
 */
export function displayColorSlots(runs: readonly { id: string; colorSlot: number }[], slotCount: number): Record<string, number> {
	const shown: Record<string, number> = {};
	const taken = new Set<number>();
	for (const { id, colorSlot } of runs) {
		let slot = colorSlot;
		for (let tried = 0; taken.has(slot) && tried < slotCount; tried++) slot = (slot + 1) % slotCount;
		shown[id] = slot;
		taken.add(slot);
	}
	return shown;
}

export interface StatSweepCsvRun {
	label: string;
	statLabel: string;
	baseStatValue: number;
	points: ReadonlyArray<StatSweepPoint>;
}

export const STAT_SWEEP_CSV_COLUMNS = [
	'run',
	'stat',
	'base_total',
	'added',
	'total',
	...STAT_SWEEP_METRICS.flatMap(metric => [metric, `${metric}_stdev`]),
	'samples',
] as const;

const csvField = (value: string | number) => {
	const text = String(value);
	return /[",\r\n]/.test(text) ? `"${text.replaceAll('"', '""')}"` : text;
};

export function statSweepCsv(runs: ReadonlyArray<StatSweepCsvRun>): string {
	const rows = runs.flatMap(run =>
		run.points.map(point => [
			run.label,
			run.statLabel,
			run.baseStatValue,
			point.offset,
			point.statValue,
			...STAT_SWEEP_METRICS.flatMap(metric => [point.metrics[metric].avg, point.metrics[metric].stdev]),
			// Every metric comes from the same iterations.
			point.metrics.dps.n,
		]),
	);
	return [STAT_SWEEP_CSV_COLUMNS, ...rows].map(row => row.map(csvField).join(',')).join('\n') + '\n';
}
