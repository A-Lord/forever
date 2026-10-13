import type { DpsCurvePoint, DpsCurveRecord } from '@sim/cache/dps_curve_history';
import type { DpsCurvePrecision } from '@sim/settings/dps_curve_settings';
import type { DpsSample } from '@sim/sim';
import { distinct } from '@sim/utils/collections';
import { stDevToConf90, sum, timeRemainingAtRate } from '@sim/utils/math';

export type DpsCurveStep = Pick<DpsCurvePoint, 'offset' | 'statValue'>;

// Rounds away float noise like 3 × 0.1 = 0.30000000000000004, which would otherwise label an axis tick.
const roundFloat = (value: number) => Math.round(value * 1e6) / 1e6;

/**
 * Multiples of `step` from −`range` to +`range`, ascending, always with 0. ±`range` itself is left out when `range` is
 * not a multiple of `step`.
 */
export function buildCurveOffsets(range: number, step: number): number[] {
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
export const curveSteps = (offsets: readonly number[], totals: readonly number[]): DpsCurveStep[] =>
	offsets.map((offset, index) => ({ offset, statValue: roundFloat(totals[index]) })).filter(step => step.statValue >= 0);

/** A step's total as its change from the run's current value. */
export const curveChange = (statValue: number, baseStatValue: number) => roundFloat(statValue - baseStatValue);

export const pointCi = (point: Pick<DpsCurvePoint, 'dpsStdev' | 'samples'>) => (point.samples ? stDevToConf90(point.dpsStdev, point.samples) : 0);

/** The stored runs plus any run in memory they do not hold yet, newest first. */
export const mergeCurveHistory = (inMemory: readonly DpsCurveRecord[], stored: readonly DpsCurveRecord[]): DpsCurveRecord[] =>
	distinct([...stored, ...inMemory], (a, b) => a.id === b.id).sort((a, b) => b.createdAt - a.createdAt);

/** A step's first pass in Target ± mode; the iterations setting plays no part there. */
export const DPS_CURVE_PROBE_ITERATIONS = 1000;
/** Passes after the first in Target ± mode, each sized from every iteration so far. */
export const DPS_CURVE_MAX_TOP_UP_PASSES = 2;
const Z90 = 1.645;

/** Two runs' DPS as one run over both runs' iterations reports it. */
export function mergeDpsSamples(a: DpsSample, b: DpsSample): DpsSample {
	const n = a.n + b.n;
	const avg = (a.avg * a.n + b.avg * b.n) / n;
	const sumSq = a.sumSq + b.sumSq;
	return { avg, stdev: Math.sqrt(Math.max(0, sumSq / n - avg * avg)), n, sumSq };
}

/** The iterations at which a DPS spread of `stdev` has a 90% ± of `targetCi`. */
export const iterationsForCi = (stdev: number, targetCi: number) => Math.ceil(((Z90 * stdev) / targetCi) ** 2);

/** Runs `iterations` more from `seedOffset` past the run's seed; null when the run stopped. */
export type DpsPassRunner = (iterations: number, seedOffset: number) => Promise<DpsSample | null>;

/**
 * One step's DPS. Iterations mode runs once. Target ± mode runs a fixed first pass, then up to
 * DPS_CURVE_MAX_TOP_UP_PASSES more, each up to the total its ± needs and capped at Max × the first pass. A pass continues at
 * the seed after the last iteration, so the passes add up to one longer run. `onPlanned` hears each total before its pass.
 */
export async function measureDpsStep(
	run: DpsPassRunner,
	precision: DpsCurvePrecision,
	defaultIterations: number,
	onPlanned?: (total: number) => void,
): Promise<DpsSample | null> {
	if (precision.precisionMode !== 'target') {
		const iterations = precision.iterations || defaultIterations;
		onPlanned?.(iterations);
		return run(iterations, 0);
	}
	onPlanned?.(DPS_CURVE_PROBE_ITERATIONS);
	let sample = await run(DPS_CURVE_PROBE_ITERATIONS, 0);
	const cap = precision.maxIterationMultiplier > 0 ? DPS_CURVE_PROBE_ITERATIONS * precision.maxIterationMultiplier : Infinity;
	for (let pass = 0; sample && pass < DPS_CURVE_MAX_TOP_UP_PASSES; pass++) {
		const total = Math.min(iterationsForCi(sample.stdev, precision.targetDpsCi), cap);
		if (total <= sample.n) break;
		onPlanned?.(total);
		const more = await run(total - sample.n, sample.n);
		sample = more && mergeDpsSamples(sample, more);
	}
	return sample;
}

export interface DpsCurveProgress {
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

export const startCurveProgress = (stepCounts: number[]): DpsCurveProgress => ({
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
export function advanceCurveProgress(progress: DpsCurveProgress, index: number, done: number, expected: number, elapsedSeconds: number): DpsCurveProgress {
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

/** The slot a new run takes: the one after the newest saved run's, wrapping at `slotCount`. */
export const nextColorSlot = (history: readonly Pick<DpsCurveRecord, 'colorSlot'>[], slotCount: number): number =>
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

export interface DpsCurveCsvRun {
	label: string;
	statLabel: string;
	baseStatValue: number;
	points: ReadonlyArray<DpsCurvePoint>;
}

export const DPS_CURVE_CSV_COLUMNS = ['run', 'stat', 'base_total', 'change', 'total', 'dps', 'dps_ci90', 'samples'] as const;

const csvField = (value: string | number) => {
	const text = String(value);
	return /[",\r\n]/.test(text) ? `"${text.replaceAll('"', '""')}"` : text;
};

export function dpsCurveCsv(runs: ReadonlyArray<DpsCurveCsvRun>): string {
	const rows = runs.flatMap(run =>
		run.points.map(point => [
			run.label,
			run.statLabel,
			run.baseStatValue,
			curveChange(point.statValue, run.baseStatValue),
			point.statValue,
			point.dps,
			pointCi(point),
			point.samples,
		]),
	);
	return [DPS_CURVE_CSV_COLUMNS, ...rows].map(row => row.map(csvField).join(',')).join('\n') + '\n';
}
