import type { StatWeightsResult } from '@generated/proto/api';
import type { Stat } from '@generated/proto/common';
import type { StatSweepPoint, StatSweepRecord } from '@sim/cache/stat_sweep_history';
import type { UnitStat } from '@sim/proto/stats';
import { distinct } from '@sim/utils/collections';
import { stDevToConf90 } from '@sim/utils/math';

import { calculateEp } from './ep_math';

export type StatSweepStep = Pick<StatSweepPoint, 'offset' | 'statValue'>;

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

/** A step's total as its change from the run's current value. */
export const sweepChange = (statValue: number, baseStatValue: number) => roundFloat(statValue - baseStatValue);

/** The stored runs plus any run in memory they do not hold yet, newest first. */
export const mergeSweepHistory = (inMemory: readonly StatSweepRecord[], stored: readonly StatSweepRecord[]): StatSweepRecord[] =>
	distinct([...stored, ...inMemory], (a, b) => a.id === b.id).sort((a, b) => b.createdAt - a.createdAt);

/** One step's point: the swept stat's DPS EP against `refStat` and the baseline DPS, each with its 90% ±. */
export function sweepPointFromResult(result: StatWeightsResult, unitStat: UnitStat, refStat: Stat, { offset, statValue }: StatSweepStep): StatSweepPoint {
	const dps = calculateEp(result, { dps: refStat, heal: undefined, tank: undefined }).dps!;
	const samples = (result.sampleCounts && unitStat.getProtoValue(result.sampleCounts)) || 0;
	const baseDps = result.baseDps;
	const dpsSamples = baseDps?.aggregatorData?.n;
	return {
		offset,
		statValue,
		ep: unitStat.getProtoValue(dps.epValues!) || 0,
		epCi: samples ? stDevToConf90(unitStat.getProtoValue(dps.epValuesStdev!) || 0, samples) : 0,
		dps: baseDps?.avg ?? 0,
		dpsCi: baseDps && dpsSamples ? stDevToConf90(baseDps.stdev, dpsSamples) : 0,
		samples,
	};
}

/** The slot a new run takes: the one after the newest saved run's, wrapping at `slotCount`. */
export const nextColorSlot = (history: readonly Pick<StatSweepRecord, 'colorSlot'>[], slotCount: number): number =>
	history.length ? ((history[0].colorSlot ?? 0) + 1) % slotCount : 0;

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
