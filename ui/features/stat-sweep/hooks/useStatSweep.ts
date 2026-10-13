import { deleteRun, listRuns, saveRun, type StatSweepRecord } from '@sim/cache/stat_sweep_history';
import { usePlayer, useSim } from '@sim/context/SimHostContext';
import { useSimRun } from '@sim/hooks/useSimRun';
import { Stats, UnitStat } from '@sim/proto/stats';
import type { StatSweepPrecision } from '@sim/settings/stat_sweep_settings';
import { SimRunKind, type StatSweepMetric } from '@sim/state/sim_store';
import { randomUUID } from '@sim/utils/misc';
import { toastManager } from '@ui-kit/Toast';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

import { SERIES_SLOT_COUNT } from '../components/StatSweepDialog/colors';
import { runLabel, type StatSweepSeries } from '../components/StatSweepDialog/StatSweepChart';
import type { StatSweepHistoryEntry } from '../components/StatSweepDialog/StatSweepHistory';
import {
	advanceSweepProgress,
	buildSweepOffsets,
	displayColorSlots,
	measureSweepStep,
	mergeSweepHistory,
	nextColorSlot,
	runSweepStat,
	skipSweepSteps,
	startSweepProgress,
	type StatSweepProgress,
	type StatSweepStep,
	type SweepPassRunner,
	sweepSteps,
} from '../model/stat_sweep';

export interface StatSweepStatArgs {
	unitStat: UnitStat;
	label: string;
	/** The stat's current total. */
	baseStatValue: number;
}

export interface StatSweepArgs {
	/** Run one after another, each on its own and saved as its own run. */
	stats: StatSweepStatArgs[];
	range: number;
	step: number;
	/** Sizes Target ±. */
	metric: StatSweepMetric;
	precision: StatSweepPrecision;
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
const shownColorSlots = (shown: StatSweepRecord[]) => displayColorSlots(shown.slice().reverse(), SERIES_SLOT_COUNT);

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

	useEffect(() => {
		let active = true;
		listRuns(player.getSpec())
			.then(stored => {
				if (!active) return;
				setView(current => {
					const merged = mergeSweepHistory(current.history, stored);
					// With no run in progress and none ticked yet, the newest stored run is drawn.
					const overlaid = current.live || current.overlaid.length || !merged.length ? current.overlaid : [merged[0].id];
					return { ...current, history: merged, overlaid };
				});
			})
			.catch(console.error);
		return () => {
			active = false;
		};
	}, [player]);

	const run = useCallback(
		(args: StatSweepArgs) => {
			const { precision, metric } = args;
			// Every step of every stat runs at this seed, so the steps differ only by the stat.
			const seed = sim.nextRngSeed();
			const defaultIterations = sim.getIterations();
			// The history read here holds none of this run's stats, so each stat takes the slot after the one before.
			const firstSlot = nextColorSlot(history.current, SERIES_SLOT_COUNT);

			// Moves a stat's run into the history and saves it, or drops it when no step finished.
			const keep = async (finished: StatSweepRecord) => {
				if (!finished.points.length) {
					setView(current => ({ ...current, live: null, overlaid: current.overlaid.filter(id => id !== finished.id) }));
					return;
				}
				setView(current => ({ ...current, live: null, history: [finished, ...current.history] }));
				await saveRun(finished).catch(console.error);
			};

			// Only this batch's runs are drawn; the user may tick older ones again afterwards.
			setView(current => ({ ...current, overlaid: [] }));

			return sim.runs.start(SimRunKind.StatSweep, async () => {
				try {
					const offsets = buildSweepOffsets(args.range, args.step);
					// Reported before the totals resolve, so the progress dialog has its full layout from the start.
					progressRef.current?.(startSweepProgress(args.stats.map(() => offsets.length)));
					// Every stat's totals before the first step, so the bar and the estimate span the whole run.
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
					// The last finished step's iterations: in Target ± mode a step's total is known only after its first pass, and
					// neighbouring steps need about as many, so the bar sizes a step by it until its own plan is larger.
					let lastStepIterations = 0;
					for (const [statIndex, { unitStat, label, baseStatValue }] of args.stats.entries()) {
						if (sim.runs.isAborting(SimRunKind.StatSweep)) break;
						const record: StatSweepRecord = {
							id: randomUUID(),
							createdAt: Date.now(),
							specId: player.getSpec(),
							label,
							colorSlot: (firstSlot + statIndex) % SERIES_SLOT_COUNT,
							unitStat: unitStat.toJson(),
							range: args.range,
							step: args.step,
							metric,
							precision,
							baseStatValue,
							points: [],
						};
						setView(current => ({ ...current, live: record, overlaid: [...current.overlaid, record.id] }));
						let finished: StatSweepRecord = record;
						let stopped = false;
						try {
							const measure = async (step: StatSweepStep, index: number) => {
								if (sim.runs.isAborting(SimRunKind.StatSweep)) return null;
								const bonusStats = new Stats().withUnitStat(unitStat, step.offset);
								let planned = 0;
								const report = (done: number) => {
									progress = advanceSweepProgress(
										progress,
										stepsBefore + index,
										done,
										Math.max(planned, lastStepIterations),
										(Date.now() - startedAt) / 1000,
									);
									progressRef.current?.(progress);
								};
								const runPass: SweepPassRunner = async (iterations, seedOffset) => {
									if (sim.runs.isAborting(SimRunKind.StatSweep)) return null;
									const result = await sim.simulateStatSweepStep({
										iterations,
										seed: seed + seedOffset,
										bonusStats,
										onProgress: metrics => report(seedOffset + metrics.completedIterations),
									});
									// An aborted pass comes back as its error outcome.
									return 'dps' in result ? result : null;
								};
								const sample = await measureSweepStep(runPass, precision, metric, defaultIterations, total => {
									planned = total;
								});
								if (sample) lastStepIterations = sample.metrics[metric].n;
								return sample;
							};
							const statRun = await runSweepStat(
								batch[statIndex],
								measure,
								({ points, cappedStep }) => {
									const live = { ...record, points, ...(cappedStep && { cappedAt: cappedStep.offset }) };
									finished = live;
									setView(current => ({ ...current, live }));
								},
								skipped => {
									progress = skipSweepSteps(progress, statIndex, skipped);
									progressRef.current?.(progress);
								},
							);
							stopped = statRun.stopped;
							stepsBefore += statRun.points.length;
						} finally {
							await keep(finished);
						}
						if (stopped) break;
					}
				} catch (error: any) {
					console.error(error);
					toastManager.add({ variant: 'error', body: error?.message || 'Something went wrong running the stat sweep.' });
				}
			});
		},
		[sim, player],
	);

	const cancel = useCallback(() => {
		abort().catch(console.error);
	}, [abort]);

	const remove = useCallback((id: string) => {
		setView(current => ({
			...current,
			history: current.history.filter(run => run.id !== id),
			overlaid: current.overlaid.filter(other => other !== id),
		}));
		deleteRun(id).catch(console.error);
	}, []);

	const setOverlaid = useCallback(
		(id: string, overlaid: boolean) =>
			setView(current => ({ ...current, overlaid: overlaid ? [...current.overlaid, id] : current.overlaid.filter(other => other !== id) })),
		[],
	);

	const series = useMemo<StatSweepSeries[]>(() => {
		const shown = shownRuns(view);
		const colorSlots = shownColorSlots(shown);
		return shown.map(run => ({
			id: run.id,
			label: run.label,
			colorSlot: colorSlots[run.id],
			statLabel: UnitStat.fromJson(run.unitStat).getFullName(player.getClass()),
			baseStatValue: run.baseStatValue,
			points: run.points,
			cappedAt: run.cappedAt,
		}));
	}, [view, player]);
	const historyEntries = useMemo<StatSweepHistoryEntry[]>(() => {
		const colorSlots = Object.fromEntries(series.map(run => [run.id, run.colorSlot]));
		return view.history.map(run => ({
			id: run.id,
			label: runLabel(run),
			createdAt: run.createdAt,
			overlaid: view.overlaid.includes(run.id),
			colorSlot: colorSlots[run.id],
			metric: run.metric,
		}));
	}, [view, series]);

	return { isRunning, series, historyEntries, run, cancel, remove, setOverlaid };
};

export type StatSweep = ReturnType<typeof useStatSweep>;
