import { describe, expect, it } from 'vitest';

import { assignColorSlots, buildSweepOffsets } from './stat_sweep';

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

	it('returns only 0 for a step or range that cannot sweep', () => {
		expect(buildSweepOffsets(100, 50, 0)).toEqual([0]);
		expect(buildSweepOffsets(100, 50, -10)).toEqual([0]);
		expect(buildSweepOffsets(100, -50, 10)).toEqual([0]);
		expect(buildSweepOffsets(100, 0, 10)).toEqual([0]);
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
