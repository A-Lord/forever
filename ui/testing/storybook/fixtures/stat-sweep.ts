import type { StatSweepPoint } from '@sim/cache/stat_sweep_history';

import { buildSweepOffsets, sweepSteps } from '../../../features/stat-weights/model/stat_sweep';

const HIT_CAP = 150;
const EP_BELOW_CAP = 4.7;
const EP_ABOVE_CAP = 0.05;
const DPS_PER_POINT = 2.1;

// EP holds near its uncapped value, then falls off over the last ~20 points before the cap; DPS climbs until the cap and then flattens.
const epAt = (value: number) => EP_ABOVE_CAP + (EP_BELOW_CAP - EP_ABOVE_CAP) / (1 + Math.exp((value - HIT_CAP) / 6));
const dpsAt = (value: number) => 1180 + DPS_PER_POINT * Math.min(value, HIT_CAP) - 0.004 * (Math.min(value, HIT_CAP) - 80) ** 2;

/** A ±100 hit sweep in steps of 10 from `base`, against a hit cap at 150. */
export const makeHitSweepPoints = (base: number, samples = 20_000): StatSweepPoint[] => {
	const offsets = buildSweepOffsets(100, 10);
	return sweepSteps(
		offsets,
		offsets.map(offset => base + offset),
	).map(({ offset, statValue }) => {
		const ep = epAt(statValue);
		return {
			offset,
			statValue,
			ep,
			epCi: 0.006 + ep * 0.002,
			dps: dpsAt(statValue),
			dpsCi: 1.8,
			samples,
		};
	});
};
