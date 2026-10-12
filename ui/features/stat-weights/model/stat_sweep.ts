import type { StatWeightsResult } from '@generated/proto/api';
import type { Stat } from '@generated/proto/common';
import type { StatSweepPoint, StatSweepPrecision, StatSweepRecord, StatSweepUnitStat } from '@sim/cache/stat_sweep_history';
import type { UnitStat } from '@sim/proto/stats';
import { stDevToConf90 } from '@sim/utils/math';

import { calculateEp } from './ep_math';

export type SweepPoint = StatSweepPoint;
export type SweepPrecision = StatSweepPrecision;
export type SweepRun = StatSweepRecord;
export type SweepUnitStat = StatSweepUnitStat;

// Rounds away float steps like 3 × 0.1 = 0.30000000000000004, which would otherwise label an axis tick.
const roundOffset = (value: number) => Math.round(value * 1e6) / 1e6;

/**
 * Multiples of `step` from −`range` to +`range`, ascending, always with 0. Each offset is a bonus stat, which moves the
 * stat's total by `multiplier` times the offset. An offset that would take the total below 0 is dropped, and so is
 * ±`range` itself when `range` is not a multiple of `step`.
 */
export function buildSweepOffsets(current: number, range: number, step: number, multiplier = 1): number[] {
	if (!(step > 0) || !(range > 0)) return [0];
	const offsets = [0];
	const count = Math.floor(roundOffset(range / step));
	for (let k = 1; k <= count; k++) {
		const offset = roundOffset(k * step);
		offsets.push(offset);
		if (roundOffset(current - multiplier * offset) >= 0) offsets.push(-offset);
	}
	return offsets.sort((a, b) => a - b);
}

/**
 * How far the stat's total moves per bonus point, from its totals without and with `step` bonus points, since % stat
 * multipliers apply to bonus stats. 1 when the total does not rise.
 */
export function sweepStatMultiplier(withoutBonus: number, withBonus: number, step: number): number {
	const multiplier = (withBonus - withoutBonus) / step;
	return multiplier > 0 && Number.isFinite(multiplier) ? multiplier : 1;
}

/** The stored runs plus any run in memory they do not hold yet, newest first. */
export function mergeSweepHistory(inMemory: readonly SweepRun[], stored: readonly SweepRun[]): SweepRun[] {
	const storedIds = new Set(stored.map(run => run.id));
	return [...stored, ...inMemory.filter(run => !storedIds.has(run.id))].sort((a, b) => b.createdAt - a.createdAt);
}

export const toSweepUnitStat = (unitStat: UnitStat): SweepUnitStat =>
	unitStat.isStat() ? { kind: 'stat', index: unitStat.getStat() } : { kind: 'pseudo', index: unitStat.getPseudoStat() };

/**
 * One step's point: the swept stat's DPS EP against `refStat` and the baseline DPS, each with its 90% ±, at the stat's
 * total, which `offset` bonus points move by `multiplier` times the offset.
 */
export function sweepPointFromResult(
	result: StatWeightsResult,
	unitStat: UnitStat,
	refStat: Stat,
	baseStatValue: number,
	offset: number,
	multiplier = 1,
): SweepPoint {
	const dps = calculateEp(result, { dps: refStat, heal: undefined, tank: undefined }).dps!;
	const samples = (result.sampleCounts && unitStat.getProtoValue(result.sampleCounts)) || 0;
	const baseDps = result.baseDps;
	const dpsSamples = baseDps?.aggregatorData?.n || samples;
	return {
		offset,
		statValue: roundOffset(baseStatValue + multiplier * offset),
		ep: unitStat.getProtoValue(dps.epValues!) || 0,
		epCi: samples ? stDevToConf90(unitStat.getProtoValue(dps.epValuesStdev!) || 0, samples) : 0,
		dps: baseDps?.avg ?? 0,
		dpsCi: baseDps && dpsSamples ? stDevToConf90(baseDps.stdev, dpsSamples) : 0,
		samples,
	};
}

/**
 * Gives each run a colour slot that stays its own while it is shown, so ticking or unticking another run never repaints
 * it. A new run takes the lowest free slot; one that finds none is left out.
 */
export function assignColorSlots(previous: Readonly<Record<string, number>>, runIds: readonly string[], slotCount: number): Record<string, number> {
	const assigned: Record<string, number> = {};
	const taken = new Set<number>();
	for (const id of runIds) {
		const slot = previous[id];
		if (slot !== undefined && slot < slotCount) {
			assigned[id] = slot;
			taken.add(slot);
		}
	}
	for (const id of runIds) {
		if (assigned[id] !== undefined) continue;
		let slot = 0;
		while (taken.has(slot)) slot++;
		if (slot >= slotCount) continue;
		assigned[id] = slot;
		taken.add(slot);
	}
	return assigned;
}
