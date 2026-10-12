import type { Stat } from '@generated/proto/common';
import { deleteRun, listRuns, saveRun } from '@sim/cache/stat_sweep_history';
import { usePlayer, useSim } from '@sim/context/SimHostContext';
import { useSimRun } from '@sim/hooks/useSimRun';
import { Stats, type UnitStat } from '@sim/proto/stats';
import type { EpPrecisionSettings } from '@sim/settings/stat_weight_settings';
import { Sim } from '@sim/sim';
import { SimRunKind } from '@sim/state/sim_store';
import { noop } from '@sim/utils/misc';
import { toastManager } from '@ui-kit/Toast';
import { useCallback, useEffect, useMemo, useState } from 'react';

import { SERIES_SLOT_COUNT } from '../components/StatSweep/colors';
import type { StatSweepProgress } from '../components/StatSweep/StatSweepPanel';
import { precisionRunOptions } from '../model/ep_math';
import {
	assignColorSlots,
	buildSweepOffsets,
	mergeSweepHistory,
	type SweepPoint,
	sweepPointFromResult,
	type SweepRun,
	sweepStatMultiplier,
	toSweepUnitStat,
} from '../model/stat_sweep';

export interface StatSweepArgs {
	unitStat: UnitStat;
	label: string;
	range: number;
	step: number;
	refStat: Stat;
	/** The swept stat's current total. */
	baseStatValue: number;
	precision: EpPrecisionSettings;
}

export interface ShownSweepRun {
	run: SweepRun;
	colorSlot: number;
}

interface SweepView {
	history: SweepRun[];
	/** The run in progress, drawn as its points arrive. */
	live: SweepRun | null;
	overlaid: ReadonlyArray<string>;
	colorSlots: Record<string, number>;
}

const shownRuns = ({ history, live, overlaid }: SweepView): SweepRun[] => [
	...(live ? [live] : []),
	...history.filter(run => run.id !== live?.id && overlaid.includes(run.id)),
];

const withColorSlots = (view: SweepView): SweepView => ({
	...view,
	colorSlots: assignColorSlots(
		view.colorSlots,
		shownRuns(view).map(run => run.id),
		SERIES_SLOT_COUNT,
	),
});

export const useStatSweep = () => {
	const sim = useSim();
	const player = usePlayer();
	const { isRunning, abort } = useSimRun(SimRunKind.StatSweep);
	const [view, setView] = useState<SweepView>({ history: [], live: null, overlaid: [], colorSlots: {} });
	const [progress, setProgress] = useState<StatSweepProgress>();

	const update = useCallback((change: (view: SweepView) => SweepView) => setView(current => withColorSlots(change(current))), []);

	useEffect(() => {
		let active = true;
		listRuns(player.getSpec())
			.then(history => {
				if (!active) return;
				update(current => {
					const merged = mergeSweepHistory(current.history, history);
					// With no run in progress and none ticked yet, the newest stored run is drawn.
					const overlaid = current.live || current.overlaid.length || !merged.length ? current.overlaid : [merged[0].id];
					return { ...current, history: merged, overlaid };
				});
			})
			.catch(console.error);
		return () => {
			active = false;
		};
	}, [player, update]);

	const run = useCallback(
		(args: StatSweepArgs) => {
			const { unitStat, refStat, baseStatValue, precision } = args;
			const totalWithBonus = (bonus: number) =>
				sim.getFinalStatsWithBonus(new Stats().withUnitStat(unitStat, bonus)).then(finalStats => finalStats.getUnitStat(unitStat));
			const measureMultiplier = async () => {
				try {
					const [withoutBonus, withBonus] = await Promise.all([totalWithBonus(0), totalWithBonus(args.step)]);
					return sweepStatMultiplier(withoutBonus, withBonus, args.step);
				} catch (error) {
					console.error(error);
					return 1;
				}
			};
			// Every step runs at the same seed, so the steps differ only by the swept stat.
			const seed = sim.getFixedRngSeed() || Math.floor(Math.random() * Sim.MAX_RNG_SEED) + 1;
			const options = { ...precisionRunOptions(precision), seed };
			const stats = unitStat.isStat() ? [unitStat.getStat()] : [];
			const pseudoStats = unitStat.isPseudoStat() ? [unitStat.getPseudoStat()] : [];
			const createdAt = Date.now();
			const record: SweepRun = {
				id: `${createdAt}-${Math.random().toString(36).slice(2, 10)}`,
				createdAt,
				specId: player.getSpec(),
				label: args.label,
				unitStat: toSweepUnitStat(unitStat),
				refStat,
				range: args.range,
				step: args.step,
				precision: {
					mode: precision.precisionMode,
					iterations: precision.iterations,
					targetEpCi: precision.targetEpCi,
					maxIterationMultiplier: precision.maxIterationMultiplier,
				},
				baseStatValue,
				points: [],
			};
			update(current => ({ ...current, live: record, overlaid: [...current.overlaid, record.id] }));

			return sim.runs.start(SimRunKind.StatSweep, async () => {
				const points: SweepPoint[] = [];
				try {
					const multiplier = await measureMultiplier();
					const offsets = buildSweepOffsets(baseStatValue, args.range, args.step, multiplier);
					for (const [index, offset] of offsets.entries()) {
						if (sim.runs.isAborting(SimRunKind.StatSweep)) break;
						setProgress({ step: index + 1, steps: offsets.length });
						const result = await player.computeStatWeights(stats, pseudoStats, refStat, noop, {
							...options,
							bonusStats: new Stats().withUnitStat(unitStat, offset),
						});
						// An aborted step comes back as a result with an error, not as a throw.
						if (result.error) break;
						points.push(sweepPointFromResult(result, unitStat, refStat, baseStatValue, offset, multiplier));
						const live = { ...record, points: points.slice() };
						update(current => ({ ...current, live }));
					}
				} catch (error: any) {
					console.error(error);
					toastManager.add({ variant: 'error', body: error?.message || 'Something went wrong running the stat sweep.' });
				}

				setProgress(undefined);
				if (!points.length) {
					update(current => ({ ...current, live: null, overlaid: current.overlaid.filter(id => id !== record.id) }));
					return;
				}
				const finished = { ...record, points };
				update(current => ({ ...current, live: null, history: [finished, ...current.history] }));
				await saveRun(finished).catch(console.error);
			});
		},
		[sim, player, update],
	);

	const cancel = useCallback(() => {
		abort().catch(console.error);
	}, [abort]);

	const remove = useCallback(
		(id: string) => {
			update(current => ({
				...current,
				history: current.history.filter(run => run.id !== id),
				overlaid: current.overlaid.filter(other => other !== id),
			}));
			deleteRun(id).catch(console.error);
		},
		[update],
	);

	const setOverlaid = useCallback(
		(id: string, overlaid: boolean) =>
			update(current => ({ ...current, overlaid: overlaid ? [...current.overlaid, id] : current.overlaid.filter(other => other !== id) })),
		[update],
	);

	const shown = useMemo<ShownSweepRun[]>(() => shownRuns(view).map(run => ({ run, colorSlot: view.colorSlots[run.id] })), [view]);

	return {
		isRunning,
		progress,
		history: view.history,
		overlaid: view.overlaid,
		colorSlots: view.colorSlots,
		shown,
		run,
		cancel,
		remove,
		setOverlaid,
	};
};
