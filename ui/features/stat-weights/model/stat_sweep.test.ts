import { DistributionMetrics, type StatWeightsResult } from '@generated/proto/api';
import { Stat } from '@generated/proto/common';
import { Stats, UnitStat } from '@sim/proto/stats';
import { describe, expect, it } from 'vitest';

import { emptyStatWeightsResult } from './ep_math';
import { assignColorSlots, buildSweepOffsets, mergeSweepHistory, sweepPointFromResult, type SweepRun, sweepStatMultiplier } from './stat_sweep';

describe('buildSweepOffsets', () => {
	it('steps from -range to +range through 0', () => {
		expect(buildSweepOffsets(200, 30, 10)).toEqual([-30, -20, -10, 0, 10, 20, 30]);
	});

	it('drops offsets that would take the stat below 0 and keeps the one that lands on 0', () => {
		expect(buildSweepOffsets(20, 30, 10)).toEqual([-20, -10, 0, 10, 20, 30]);
		expect(buildSweepOffsets(0, 20, 10)).toEqual([0, 10, 20]);
	});

	it('stops at the last multiple of step inside the range', () => {
		expect(buildSweepOffsets(100, 25, 10)).toEqual([-20, -10, 0, 10, 20]);
		expect(buildSweepOffsets(100, 5, 10)).toEqual([0]);
	});

	it('keeps fractional steps exact', () => {
		expect(buildSweepOffsets(1, 0.3, 0.1)).toEqual([-0.3, -0.2, -0.1, 0, 0.1, 0.2, 0.3]);
	});

	it('keeps offsets whose total, scaled by the multiplier, stays at or above 0', () => {
		expect(buildSweepOffsets(20, 30, 10, 1.1)).toEqual([-10, 0, 10, 20, 30]);
		expect(buildSweepOffsets(22, 30, 10, 1.1)).toEqual([-20, -10, 0, 10, 20, 30]);
	});

	it('returns only 0 for a step or range that cannot sweep', () => {
		expect(buildSweepOffsets(100, 50, 0)).toEqual([0]);
		expect(buildSweepOffsets(100, 50, -10)).toEqual([0]);
		expect(buildSweepOffsets(100, -50, 10)).toEqual([0]);
		expect(buildSweepOffsets(100, 0, 10)).toEqual([0]);
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
		expect(sweepPointFromResult(result(), HIT, Stat.StatStrength, 120, -20)).toEqual({
			offset: -20,
			statValue: 100,
			ep: 4,
			epCi: (1.645 * 2) / 20,
			dps: 1500,
			dpsCi: (1.645 * 100) / 50,
			samples: 400,
		});
	});

	it('places the point at the base value plus the offset scaled by the multiplier', () => {
		expect(sweepPointFromResult(result(), HIT, Stat.StatStrength, 120, -20, 1.1).statValue).toBe(98);
	});

	it('reports no ± for a stat with no samples', () => {
		const empty = result();
		empty.sampleCounts = new Stats().toProto();

		expect(sweepPointFromResult(empty, HIT, Stat.StatStrength, 120, 0).epCi).toBe(0);
	});
});

describe('sweepStatMultiplier', () => {
	it('divides the rise in the total by the bonus points', () => {
		expect(sweepStatMultiplier(200, 211, 10)).toBeCloseTo(1.1);
	});

	it('falls back to 1 when the total does not rise', () => {
		expect(sweepStatMultiplier(200, 200, 10)).toBe(1);
		expect(sweepStatMultiplier(200, 190, 10)).toBe(1);
		expect(sweepStatMultiplier(200, 210, 0)).toBe(1);
	});
});

describe('mergeSweepHistory', () => {
	const sweepRun = (id: string, createdAt: number) => ({ id, createdAt }) as SweepRun;

	it('keeps a run finished before the stored list arrived, newest first', () => {
		const merged = mergeSweepHistory([sweepRun('new', 3)], [sweepRun('old', 1), sweepRun('older', 0)]);
		expect(merged.map(run => run.id)).toEqual(['new', 'old', 'older']);
	});

	it('lists a run held both in memory and in storage once', () => {
		const merged = mergeSweepHistory([sweepRun('a', 2)], [sweepRun('a', 2), sweepRun('b', 1)]);
		expect(merged.map(run => run.id)).toEqual(['a', 'b']);
	});
});

describe('assignColorSlots', () => {
	it('gives new runs the lowest free slots in order', () => {
		expect(assignColorSlots({}, ['a', 'b', 'c'], 8)).toEqual({ a: 0, b: 1, c: 2 });
	});

	it('keeps a shown run on its slot when another run leaves or joins', () => {
		const first = assignColorSlots({}, ['a', 'b', 'c'], 8);
		const withoutA = assignColorSlots(first, ['b', 'c'], 8);
		expect(withoutA).toEqual({ b: 1, c: 2 });
		expect(assignColorSlots(withoutA, ['b', 'c', 'd'], 8)).toEqual({ b: 1, c: 2, d: 0 });
	});

	it('leaves out a run once every slot is taken', () => {
		expect(assignColorSlots({}, ['a', 'b', 'c'], 2)).toEqual({ a: 0, b: 1 });
	});
});
