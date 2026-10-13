import type { DpsCurvePoint } from '@sim/cache/dps_curve_history';

import type { DpsCurveSeries } from '../../../features/dps-curve/components/DpsCurveDialog/DpsCurveChart';
import { buildCurveOffsets, curveSteps } from '../../../features/dps-curve/model/dps_curve';

const HIT_CAP = 150;
const DPS_PER_POINT = 2.1;

// DPS climbs until the cap and then flattens.
const dpsAt = (value: number) => 1180 + DPS_PER_POINT * Math.min(value, HIT_CAP) - 0.004 * (Math.min(value, HIT_CAP) - 80) ** 2;

const pointsFrom = (base: number, samples: number, dps: (value: number) => number): DpsCurvePoint[] => {
	const offsets = buildCurveOffsets(100, 10);
	return curveSteps(
		offsets,
		offsets.map(offset => base + offset),
	).map(({ offset, statValue }) => ({ offset, statValue, dps: dps(statValue), dpsStdev: 450, samples }));
};

/** A ±100 hit curve in steps of 10 from `base`, against a hit cap at 150. */
export const makeHitCurvePoints = (base: number, samples = 20_000): DpsCurvePoint[] => pointsFrom(base, samples, dpsAt);

/** A ±100 crit curve in steps of 10 from `base`, with no cap in reach. */
export const makeCritCurvePoints = (base: number, samples = 20_000): DpsCurvePoint[] => pointsFrom(base, samples, value => 1300 + 1.1 * value);

export const curveSeries = (
	id: string,
	statLabel: string,
	base: number,
	points: DpsCurvePoint[],
	colorSlot = 0,
	label = `${statLabel} ±100`,
): DpsCurveSeries => ({ id, label, colorSlot, statLabel, baseStatValue: base, points });
