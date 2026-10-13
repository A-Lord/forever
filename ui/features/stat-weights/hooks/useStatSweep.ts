import type { ProgressMetrics } from '@generated/proto/api';
import type { Stat } from '@generated/proto/common';
import { deleteRun, listRuns, saveRun, type StatSweepPoint, type StatSweepRecord } from '@sim/cache/stat_sweep_history';
import { usePlayer, useSim } from '@sim/context/SimHostContext';
import { useSimRun } from '@sim/hooks/useSimRun';
import { Stats, UnitStat } from '@sim/proto/stats';
import type { EpPrecisionSettings } from '@sim/settings/stat_weight_settings';
import { SimRunKind } from '@sim/state/sim_store';
import { randomUUID } from '@sim/utils/misc';
import { toastManager } from '@ui-kit/Toast';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

import { SERIES_SLOT_COUNT } from '../components/StatSweep/colors';
import type { StatSweepSeries } from '../components/StatSweep/StatSweepChart';
import type { StatSweepHistoryEntry } from '../components/StatSweep/StatSweepHistory';
import { precisionRunOptions } from '../model/ep_math';
import {
	advanceSweepProgress,
	buildSweepOffsets,
	displayColorSlots,
	mergeSweepHistory,
	nextColorSlot,
	startSweepProgress,
	type StatSweepProgress,
	sweepPointFromResult,
	sweepSteps,
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

export interface UseStatSweepOptions {
	onProgress?: (progress: StatSweepProgress) => void;
}

interface SweepView {
	history: StatSweepRecord[];
	/** The run in progress, drawn as its points arrive. */
	live: StatSweepRecord | null;
	overlaid: ReadonlyArray<string>;
}

const shownRuns = ({ history, live, overlaid }: SweepView): StatSweepRecord[] => [...(live ? [live] : []), ...history.filter(run => overlaid.includes(run.id))];

// The slot of each shown run, oldest first so a clash moves the newer run.
const shownColorSlots = (view: SweepView, shown: StatSweepRecord[]) =>
	displayColorSlots(
		shown.map(run => ({ id: run.id, colorSlot: run.colorSlot ?? Math.max(view.history.indexOf(run), 0) % SERIES_SLOT_COUNT })).reverse(),
		SERIES_SLOT_COUNT,
	);

export const useStatSweep = ({ onProgress }: UseStatSweepOptions = {}) => {
	const sim = useSim();
	const player = usePlayer();
	// Read at emit time, so a caller may pass a fresh closure every render.
	const progressRef = useRef(onProgress);
	progressRef.current = onProgress;
	const { isRunning, abort } = useSimRun(SimRunKind.StatSweep);
	const [view, setView] = useState<SweepView>({ history: [], live: null, overlaid: [] });
	const history = useRef(view.history);
	history.current = view.history;

	const update = setView;

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
			const { unitStat, refStat, precision } = args;
			// Every step runs at the same seed, so the steps differ only by the swept stat.
			const options = { ...precisionRunOptions(precision), seed: sim.nextRngSeed() };
			const stats = unitStat.isStat() ? [unitStat.getStat()] : [];
			const pseudoStats = unitStat.isPseudoStat() ? [unitStat.getPseudoStat()] : [];
			const record: StatSweepRecord = {
				id: randomUUID(),
				createdAt: Date.now(),
				specId: player.getSpec(),
				label: args.label,
				colorSlot: nextColorSlot(history.current, SERIES_SLOT_COUNT),
				unitStat: unitStat.toJson(),
				refStat,
				range: args.range,
				step: args.step,
				precision,
				baseStatValue: args.baseStatValue,
				points: [],
			};
			update(current => ({ ...current, live: record, overlaid: [...current.overlaid, record.id] }));

			return sim.runs.start(SimRunKind.StatSweep, async () => {
				const points: StatSweepPoint[] = [];
				try {
					const offsets = buildSweepOffsets(args.range, args.step);
					// Reported before the totals resolve, so the progress dialog has its full layout from the start.
					progressRef.current?.(startSweepProgress(offsets.length));
					const totals = await Promise.all(
						offsets.map(offset =>
							sim.getFinalStatsWithBonus(new Stats().withUnitStat(unitStat, offset)).then(finalStats => finalStats.getUnitStat(unitStat)),
						),
					);
					const steps = sweepSteps(offsets, totals);
					const startedAt = Date.now();
					let progress = startSweepProgress(steps.length);
					progressRef.current?.(progress);
					for (const [index, step] of steps.entries()) {
						if (sim.runs.isAborting(SimRunKind.StatSweep)) break;
						const onStepProgress = (metrics: ProgressMetrics) => {
							progress = advanceSweepProgress(progress, index, metrics, (Date.now() - startedAt) / 1000);
							progressRef.current?.(progress);
						};
						const result = await player.computeStatWeights(stats, pseudoStats, refStat, onStepProgress, {
							...options,
							bonusStats: new Stats().withUnitStat(unitStat, step.offset),
						});
						// An aborted step comes back as a result with an error, not as a throw.
						if (result.error) break;
						points.push(sweepPointFromResult(result, unitStat, refStat, step));
						const live = { ...record, points: points.slice() };
						update(current => ({ ...current, live }));
					}
				} catch (error: any) {
					console.error(error);
					toastManager.add({ variant: 'error', body: error?.message || 'Something went wrong running the stat sweep.' });
				}

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

	const colorSlots = useMemo(() => shownColorSlots(view, shownRuns(view)), [view]);
	const series = useMemo<StatSweepSeries[]>(
		() =>
			shownRuns(view).map(run => ({
				id: run.id,
				label: run.label,
				colorSlot: colorSlots[run.id],
				statLabel: UnitStat.fromJson(run.unitStat).getFullName(player.getClass()),
				baseStatValue: run.baseStatValue,
				points: run.points,
			})),
		[view, colorSlots, player],
	);
	const historyEntries = useMemo<StatSweepHistoryEntry[]>(
		() =>
			view.history.map(run => ({
				id: run.id,
				label: run.label,
				createdAt: run.createdAt,
				overlaid: view.overlaid.includes(run.id),
				colorSlot: colorSlots[run.id],
			})),
		[view, colorSlots],
	);

	return { isRunning, series, historyEntries, run, cancel, remove, setOverlaid };
};
