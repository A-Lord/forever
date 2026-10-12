import type { StatSweepPoint, StatSweepPrecision, StatSweepRecord, StatSweepUnitStat } from '@sim/cache/stat_sweep_history';

export type SweepPoint = StatSweepPoint;
export type SweepPrecision = StatSweepPrecision;
export type SweepRun = StatSweepRecord;
export type SweepUnitStat = StatSweepUnitStat;

// Rounds away float steps like 3 × 0.1 = 0.30000000000000004, which would otherwise label an axis tick.
const roundOffset = (value: number) => Math.round(value * 1e6) / 1e6;

/**
 * Multiples of `step` from −`range` to +`range`, ascending, always with 0. An offset that would take the stat below 0 is
 * dropped, and so is ±`range` itself when `range` is not a multiple of `step`.
 */
export function buildSweepOffsets(current: number, range: number, step: number): number[] {
	if (!(step > 0) || !(range > 0)) return [0];
	const offsets = [0];
	const count = Math.floor(roundOffset(range / step));
	for (let k = 1; k <= count; k++) {
		const offset = roundOffset(k * step);
		offsets.push(offset);
		if (current - offset >= 0) offsets.push(-offset);
	}
	return offsets.sort((a, b) => a - b);
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
