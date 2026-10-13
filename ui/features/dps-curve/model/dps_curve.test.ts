import type { DpsCurveRecord } from '@sim/cache/dps_curve_history';
import type { DpsCurvePrecision } from '@sim/settings/dps_curve_settings';
import type { DpsSample } from '@sim/sim';
import { describe, expect, it, vi } from 'vitest';

import {
	advanceCurveProgress,
	buildCurveOffsets,
	curveSteps,
	displayColorSlots,
	DPS_CURVE_MAX_TOP_UP_PASSES,
	DPS_CURVE_PROBE_ITERATIONS,
	dpsCurveCsv,
	type DpsPassRunner,
	iterationsForCi,
	measureDpsStep,
	mergeCurveHistory,
	mergeDpsSamples,
	nextColorSlot,
	startCurveProgress,
} from './dps_curve';

// A run of `values`, as the sim reports it: population stdev and the sum of squares.
const sampleOf = (values: number[]): DpsSample => {
	const n = values.length;
	const avg = values.reduce((total, value) => total + value, 0) / n;
	const sumSq = values.reduce((total, value) => total + value * value, 0);
	return { avg, stdev: Math.sqrt(sumSq / n - avg * avg), n, sumSq };
};

const precision = (patch: Partial<DpsCurvePrecision> = {}): DpsCurvePrecision => ({
	precisionMode: 'iterations',
	iterations: 0,
	targetDpsCi: 1,
	maxIterationMultiplier: 0,
	...patch,
});

describe('buildCurveOffsets', () => {
	it('steps from -range to +range through 0', () => {
		expect(buildCurveOffsets(30, 10)).toEqual([-30, -20, -10, 0, 10, 20, 30]);
	});

	it('stops at the last multiple of step inside the range', () => {
		expect(buildCurveOffsets(25, 10)).toEqual([-20, -10, 0, 10, 20]);
		expect(buildCurveOffsets(5, 10)).toEqual([0]);
	});

	it('keeps fractional steps exact', () => {
		expect(buildCurveOffsets(0.3, 0.1)).toEqual([-0.3, -0.2, -0.1, 0, 0.1, 0.2, 0.3]);
	});

	it('returns only 0 for a step or range that cannot step', () => {
		expect(buildCurveOffsets(50, 0)).toEqual([0]);
		expect(buildCurveOffsets(50, -10)).toEqual([0]);
		expect(buildCurveOffsets(-50, 10)).toEqual([0]);
	});
});

describe('curveSteps', () => {
	it('places each offset at its total and drops those whose total falls below 0, keeping one that lands on 0', () => {
		expect(curveSteps([-30, -20, -10, 0, 10], [-13, 0, 11, 22, 33])).toEqual([
			{ offset: -20, statValue: 0 },
			{ offset: -10, statValue: 11 },
			{ offset: 0, statValue: 22 },
			{ offset: 10, statValue: 33 },
		]);
	});

	it('rounds float noise out of the totals', () => {
		expect(curveSteps([10], [0.1 + 0.2])).toEqual([{ offset: 10, statValue: 0.3 }]);
	});
});

describe('mergeDpsSamples', () => {
	it('gives the mean, spread and count of one run over both runs', () => {
		const first = [1000, 1100, 1250];
		const second = [900, 1300];

		const merged = mergeDpsSamples(sampleOf(first), sampleOf(second));
		const whole = sampleOf([...first, ...second]);

		expect(merged.n).toBe(whole.n);
		expect(merged.avg).toBeCloseTo(whole.avg, 9);
		expect(merged.stdev).toBeCloseTo(whole.stdev, 9);
		expect(merged.sumSq).toBeCloseTo(whole.sumSq, 6);
	});
});

describe('iterationsForCi', () => {
	it('sizes the run whose 90% ± of the spread meets the target', () => {
		expect(iterationsForCi(100, 1)).toBe(Math.ceil(164.5 ** 2));
		expect(iterationsForCi(100, 2)).toBe(Math.ceil(82.25 ** 2));
	});
});

describe('measureDpsStep', () => {
	const runner = (stdev: number) => vi.fn(async (iterations: number) => ({ avg: 1000, stdev, n: iterations, sumSq: iterations * (stdev ** 2 + 1000 ** 2) }));

	it('runs the iterations setting once from the seed, or the sim iterations when it is 0', async () => {
		const run = runner(100);

		expect((await measureDpsStep(run, precision({ iterations: 3000 }), 12500))?.n).toBe(3000);
		expect((await measureDpsStep(run, precision(), 12500))?.n).toBe(12500);
		expect(run.mock.calls).toEqual([
			[3000, 0],
			[12500, 0],
		]);
	});

	it('probes, then tops up to the total the target needs, continuing at the seed after the probe', async () => {
		const run = runner(10);
		const planned: number[] = [];

		const sample = await measureDpsStep(run, precision({ precisionMode: 'target', targetDpsCi: 0.5, iterations: 3000 }), 12500, total =>
			planned.push(total),
		);

		const total = iterationsForCi(10, 0.5);
		expect(run.mock.calls).toEqual([
			[DPS_CURVE_PROBE_ITERATIONS, 0],
			[total - DPS_CURVE_PROBE_ITERATIONS, DPS_CURVE_PROBE_ITERATIONS],
		]);
		expect(planned).toEqual([DPS_CURVE_PROBE_ITERATIONS, total]);
		expect(sample?.n).toBe(total);
	});

	it('stops after the probe when it already meets the target', async () => {
		const run = runner(1);

		await measureDpsStep(run, precision({ precisionMode: 'target', targetDpsCi: 1 }), 12500);

		expect(run).toHaveBeenCalledTimes(1);
	});

	it('caps a step at Max × the probe', async () => {
		const run = runner(100);

		const sample = await measureDpsStep(run, precision({ precisionMode: 'target', targetDpsCi: 0.1, maxIterationMultiplier: 4 }), 12500);

		expect(sample?.n).toBe(4 * DPS_CURVE_PROBE_ITERATIONS);
	});

	it('resizes each top-up from every iteration so far, and stops after the last pass', async () => {
		// The spread grows with each pass, so every plan asks for more.
		let stdev = 10;
		const run = vi.fn<DpsPassRunner>(async iterations => {
			stdev *= 2;
			return { avg: 1000, stdev, n: iterations, sumSq: iterations * (stdev ** 2 + 1000 ** 2) };
		});

		await measureDpsStep(run, precision({ precisionMode: 'target', targetDpsCi: 0.5 }), 12500);

		expect(run).toHaveBeenCalledTimes(1 + DPS_CURVE_MAX_TOP_UP_PASSES);
		const seedOffsets = run.mock.calls.map(call => call[1]);
		const iterations = run.mock.calls.map(call => call[0]);
		expect(seedOffsets).toEqual([0, iterations[0], iterations[0] + iterations[1]]);
	});

	it('reports nothing when a pass stops', async () => {
		const run = vi
			.fn<(iterations: number, seedOffset: number) => Promise<DpsSample | null>>()
			.mockResolvedValueOnce(sampleOf([900, 1100]))
			.mockResolvedValueOnce(null);

		expect(await measureDpsStep(run, precision({ precisionMode: 'target', targetDpsCi: 0.1 }), 12500)).toBeNull();
	});
});

describe('advanceCurveProgress', () => {
	it('counts the steps done plus the running step share of its expected iterations, and estimates the rest at the rate so far', () => {
		const progress = advanceCurveProgress(startCurveProgress([4]), 1, 500, 1000, 30);

		expect(progress).toMatchObject({ stat: 1, stats: 1, step: 2, steps: 4, completed: 1.5, totalSteps: 4, iterationsPerStep: 1000, secondsRemaining: 50 });
	});

	it("places a step counted across every stat at its stat and that stat's own step", () => {
		const start = startCurveProgress([3, 2, 4]);

		expect(advanceCurveProgress(start, 5, 500, 1000, 55)).toMatchObject({ stat: 3, step: 1, steps: 4, completed: 5.5, secondsRemaining: 35 });
		expect(advanceCurveProgress(start, 4, 0, 1000, 1)).toMatchObject({ stat: 2, step: 2, steps: 2 });
	});

	it('never falls back when a step plans more iterations than it expected', () => {
		const before = advanceCurveProgress(startCurveProgress([4]), 1, 900, 1000, 30);
		const grown = advanceCurveProgress(before, 1, 1000, 5000, 31);

		expect(grown.completed).toBe(before.completed);
		expect(grown.iterationsPerStep).toBe(5000);
	});

	it('has no estimate before anything completes', () => {
		expect(advanceCurveProgress(startCurveProgress([4]), 0, 0, 1000, 5).secondsRemaining).toBeUndefined();
	});
});

describe('mergeCurveHistory', () => {
	const curveRun = (id: string, createdAt: number, label = id) => ({ id, createdAt, label }) as DpsCurveRecord;

	it('keeps a run finished before the stored list arrived, newest first, and lists a run held in both once, as stored', () => {
		const merged = mergeCurveHistory([curveRun('new', 3), curveRun('a', 2, 'memory')], [curveRun('a', 2, 'stored'), curveRun('older', 0)]);

		expect(merged.map(run => [run.id, run.label])).toEqual([
			['new', 'new'],
			['a', 'stored'],
			['older', 'older'],
		]);
	});
});

describe('color slots', () => {
	it('starts at 0 and follows the newest saved run, wrapping at the slot count', () => {
		expect(nextColorSlot([], 8)).toBe(0);
		expect(nextColorSlot([{ colorSlot: 2 }, { colorSlot: 5 }], 8)).toBe(3);
		expect(nextColorSlot([{ colorSlot: 7 }], 8)).toBe(0);
	});

	it('moves the later of two runs sharing a slot to the next free one', () => {
		expect(
			displayColorSlots(
				[
					{ id: 'a', colorSlot: 1 },
					{ id: 'b', colorSlot: 2 },
					{ id: 'c', colorSlot: 1 },
				],
				4,
			),
		).toEqual({ a: 1, b: 2, c: 3 });
	});
});

describe('dpsCurveCsv', () => {
	it('writes one row per point of every run, with the change, the spread, the 90% ± and a quoted label', () => {
		const csv = dpsCurveCsv([
			{
				label: 'Hit, ±100',
				statLabel: 'Hit Rating',
				baseStatValue: 120,
				points: [
					{ offset: -20, statValue: 100, dps: 1500.5, dpsStdev: 100, samples: 10000 },
					{ offset: 0, statValue: 120, dps: 1520, dpsStdev: 0, samples: 10000 },
				],
			},
			{ label: 'Crit', statLabel: 'Crit "Rating"', baseStatValue: 50, points: [{ offset: 10, statValue: 60, dps: 1600, dpsStdev: 200, samples: 400 }] },
		]);

		expect(csv.split('\n')).toEqual([
			'run,stat,base_total,change,total,dps,dps_stdev,dps_ci90,samples',
			'"Hit, ±100",Hit Rating,120,-20,100,1500.5,100,1.645,10000',
			'"Hit, ±100",Hit Rating,120,0,120,1520,0,0,10000',
			'Crit,"Crit ""Rating""",50,10,60,1600,200,16.45,400',
			'',
		]);
	});
});
