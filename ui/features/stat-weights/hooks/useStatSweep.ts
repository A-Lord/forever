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

export interface StatSweepStatArgs {
	unitStat: UnitStat;
	label: string;
	/** The stat's current total. */
	baseStatValue: number;
}

export interface StatSweepArgs {
	/** Swept one after another, each on its own and saved as its own run. */
	stats: StatSweepStatArgs[];
	range: number;
	step: number;
	refStat: Stat;
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
			const { refStat, precision } = args;
			// Every step runs at the same seed, so the steps differ only by the swept stat.
			const options = { ...precisionRunOptions(precision), seed: sim.nextRngSeed() };
			// The history read here holds none of this sweep's runs, so each stat takes the slot after the one before.
			const firstSlot = nextColorSlot(history.current, SERIES_SLOT_COUNT);

			// Moves a stat's run into the history and saves it, or drops it when no step finished.
			const keep = async (record: StatSweepRecord, points: StatSweepPoint[]) => {
				if (!points.length) {
					update(current => ({ ...current, live: null, overlaid: current.overlaid.filter(id => id !== record.id) }));
					return;
				}
				const finished = { ...record, points };
				update(current => ({ ...current, live: null, history: [finished, ...current.history] }));
				await saveRun(finished).catch(console.error);
			};

			return sim.runs.start(SimRunKind.StatSweep, async () => {
				try {
					const offsets = buildSweepOffsets(args.range, args.step);
					// Reported before the totals resolve, so the progress dialog has its full layout from the start.
					progressRef.current?.(startSweepProgress(args.stats.map(() => offsets.length)));
					// Every stat's totals before the first step, so the bar and the estimate span the whole sweep.
					const batch = await Promise.all(
						args.stats.map(({ unitStat }) =>
							Promise.all(
								offsets.map(offset =>
									sim.getFinalStatsWithBonus(new Stats().withUnitStat(unitStat, offset)).then(finalStats => finalStats.getUnitStat(unitStat)),
								),
							).then(totals => sweepSteps(offsets, totals)),
						),
					);
					const startedAt = Date.now();
					let progress = startSweepProgress(batch.map(steps => steps.length));
					progressRef.current?.(progress);
					let stepsBefore = 0;
					for (const [statIndex, { unitStat, label, baseStatValue }] of args.stats.entries()) {
						if (sim.runs.isAborting(SimRunKind.StatSweep)) break;
						const stats = unitStat.isStat() ? [unitStat.getStat()] : [];
						const pseudoStats = unitStat.isPseudoStat() ? [unitStat.getPseudoStat()] : [];
						const record: StatSweepRecord = {
							id: randomUUID(),
							createdAt: Date.now(),
							specId: player.getSpec(),
							label,
							colorSlot: (firstSlot + statIndex) % SERIES_SLOT_COUNT,
							unitStat: unitStat.toJson(),
							refStat,
							range: args.range,
							step: args.step,
							precision,
							baseStatValue,
							points: [],
						};
						update(current => ({ ...current, live: record, overlaid: [...current.overlaid, record.id] }));
						const points: StatSweepPoint[] = [];
						let stopped = false;
						try {
							for (const [index, step] of batch[statIndex].entries()) {
								if (sim.runs.isAborting(SimRunKind.StatSweep)) {
									stopped = true;
									break;
								}
								const onStepProgress = (metrics: ProgressMetrics) => {
									progress = advanceSweepProgress(progress, stepsBefore + index, metrics, (Date.now() - startedAt) / 1000);
									progressRef.current?.(progress);
								};
								const result = await player.computeStatWeights(stats, pseudoStats, refStat, onStepProgress, {
									...options,
									bonusStats: new Stats().withUnitStat(unitStat, step.offset),
								});
								// An aborted step comes back as a result with an error, not as a throw.
								if (result.error) {
									stopped = true;
									break;
								}
								points.push(sweepPointFromResult(result, unitStat, refStat, step));
								const live = { ...record, points: points.slice() };
								update(current => ({ ...current, live }));
							}
						} finally {
							await keep(record, points);
						}
						if (stopped) break;
						stepsBefore += batch[statIndex].length;
					}
				} catch (error: any) {
					console.error(error);
					toastManager.add({ variant: 'error', body: error?.message || 'Something went wrong running the stat sweep.' });
				}
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
