import { DistributionMetrics, type StatWeightsResult } from '@generated/proto/api';
import { Stat } from '@generated/proto/common';
import type { StatSweepRecord } from '@sim/cache/stat_sweep_history';
import { Stats, UnitStat } from '@sim/proto/stats';
import { describe, expect, it } from 'vitest';

import { emptyStatWeightsResult } from './ep_math';
import {
	advanceSweepProgress,
	buildSweepOffsets,
	displayColorSlots,
	mergeSweepHistory,
	nextColorSlot,
	startSweepProgress,
	sweepPointFromResult,
	sweepSteps,
} from './stat_sweep';

describe('advanceSweepProgress', () => {
	it('counts the steps done plus the running step share, and estimates the rest at the rate so far', () => {
		const progress = advanceSweepProgress(startSweepProgress(4), 1, { completedIterations: 500, totalIterations: 1000 }, 30);

		expect(progress).toEqual({ step: 2, steps: 4, completed: 1.5, iterationsPerStep: 1000, secondsRemaining: 50 });
	});

	it('never falls back, keeping the last iteration count and estimate while a new step has not reported a total', () => {
		const before = advanceSweepProgress(startSweepProgress(4), 1, { completedIterations: 900, totalIterations: 1000 }, 30);
		const grown = advanceSweepProgress(before, 1, { completedIterations: 900, totalIterations: 3000 }, 31);
		const fresh = advanceSweepProgress(grown, 2, { completedIterations: 0, totalIterations: 0 }, 32);

		expect(grown.completed).toBe(before.completed);
		expect(fresh).toMatchObject({ step: 3, completed: 2, iterationsPerStep: 3000 });
		expect(fresh.secondsRemaining).toBeDefined();
	});

	it('has no estimate before anything completes', () => {
		expect(advanceSweepProgress(startSweepProgress(4), 0, { completedIterations: 0, totalIterations: 1000 }, 5).secondsRemaining).toBeUndefined();
	});
});

describe('buildSweepOffsets', () => {
	it('steps from -range to +range through 0', () => {
		expect(buildSweepOffsets(30, 10)).toEqual([-30, -20, -10, 0, 10, 20, 30]);
	});

	it('stops at the last multiple of step inside the range', () => {
		expect(buildSweepOffsets(25, 10)).toEqual([-20, -10, 0, 10, 20]);
		expect(buildSweepOffsets(5, 10)).toEqual([0]);
	});

	it('keeps fractional steps exact', () => {
		expect(buildSweepOffsets(0.3, 0.1)).toEqual([-0.3, -0.2, -0.1, 0, 0.1, 0.2, 0.3]);
	});

	it('returns only 0 for a step or range that cannot sweep', () => {
		expect(buildSweepOffsets(50, 0)).toEqual([0]);
		expect(buildSweepOffsets(50, -10)).toEqual([0]);
		expect(buildSweepOffsets(-50, 10)).toEqual([0]);
		expect(buildSweepOffsets(0, 10)).toEqual([0]);
	});
});

describe('sweepSteps', () => {
	it('places each offset at its total and drops those whose total falls below 0, keeping one that lands on 0', () => {
		expect(sweepSteps([-30, -20, -10, 0, 10], [-13, 0, 11, 22, 33])).toEqual([
			{ offset: -20, statValue: 0 },
			{ offset: -10, statValue: 11 },
			{ offset: 0, statValue: 22 },
			{ offset: 10, statValue: 33 },
		]);
	});

	it('rounds float noise out of the totals', () => {
		expect(sweepSteps([10], [0.1 + 0.2])).toEqual([{ offset: 10, statValue: 0.3 }]);
	});
});

describe('sweepPointFromResult', () => {
	const HIT = UnitStat.fromStat(Stat.StatHitRating);
	const result = (): StatWeightsResult => {
		const swept = emptyStatWeightsResult();
		swept.dps!.weights = Stats.fromMap({ [Stat.StatStrength]: 2, [Stat.StatHitRating]: 8 }).toProto();
		swept.dps!.weightsStdev = Stats.fromMap({ [Stat.StatHitRating]: 4 }).toProto();
		swept.sampleCounts = Stats.fromMap({ [Stat.StatHitRating]: 400 }).toProto();
		swept.baseDps = DistributionMetrics.create({ avg: 1500, stdev: 100, aggregatorData: { n: 2500, sumSq: 0 } });
		return swept;
	};

	it('reads the EP against the reference stat and sizes both ± by their own sample counts', () => {
		expect(sweepPointFromResult(result(), HIT, Stat.StatStrength, { offset: -20, statValue: 98 })).toEqual({
			offset: -20,
			statValue: 98,
			ep: 4,
			epCi: (1.645 * 2) / 20,
			dps: 1500,
			dpsCi: (1.645 * 100) / 50,
			samples: 400,
		});
	});

	it('reports no ± for a stat with no samples', () => {
		const empty = result();
		empty.sampleCounts = new Stats().toProto();

		expect(sweepPointFromResult(empty, HIT, Stat.StatStrength, { offset: 0, statValue: 120 }).epCi).toBe(0);
	});
});

describe('mergeSweepHistory', () => {
	const sweepRun = (id: string, createdAt: number, label = id) => ({ id, createdAt, label }) as StatSweepRecord;

	it('keeps a run finished before the stored list arrived, newest first', () => {
		const merged = mergeSweepHistory([sweepRun('new', 3)], [sweepRun('old', 1), sweepRun('older', 0)]);
		expect(merged.map(run => run.id)).toEqual(['new', 'old', 'older']);
	});

	it('lists a run held both in memory and in storage once, as stored', () => {
		const merged = mergeSweepHistory([sweepRun('a', 2, 'memory')], [sweepRun('a', 2, 'stored'), sweepRun('b', 1)]);
		expect(merged.map(run => [run.id, run.label])).toEqual([
			['a', 'stored'],
			['b', 'b'],
		]);
	});
});

describe('nextColorSlot', () => {
	it('starts at 0 and follows the newest saved run, wrapping at the slot count', () => {
		expect(nextColorSlot([], 8)).toBe(0);
		expect(nextColorSlot([{ colorSlot: 2 }, { colorSlot: 5 }], 8)).toBe(3);
		expect(nextColorSlot([{ colorSlot: 7 }], 8)).toBe(0);
	});
});

describe('displayColorSlots', () => {
	it('keeps every run on its own slot while they differ', () => {
		expect(
			displayColorSlots(
				[
					{ id: 'a', colorSlot: 3 },
					{ id: 'b', colorSlot: 1 },
				],
				8,
			),
		).toEqual({ a: 3, b: 1 });
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
