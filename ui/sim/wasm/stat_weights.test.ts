// The backend omits RequestLow for the school-split hit/crit stats: sim/core/statweight.go
// zeroes their low-side stat modifier and only populates lowSimRequest when it is non-zero.
// Nothing in the type system catches a missing guard here — a non-null assertion on the low
// request type-checks and throws at runtime — so the absent-low path is covered explicitly.
import {
	ProgressMetrics,
	RaidSimRequest,
	RaidSimResult,
	SimOptions,
	StatWeightsRequest,
	StatWeightsResult,
	StatWeightsStatData,
	StatWeightsTarget,
	StatWeightsTargetMetric,
	StatWeightsTopUpPlan,
} from '@generated/proto/api';
import { Stat } from '@generated/proto/common';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { SimSignals } from '../sim_signal_manager';
import { noop } from '../utils/misc';
import type { WorkerPool } from '../workers/worker_pool';
import { runConcurrentStatWeights } from './stat_weights';

const runConcurrentSim = vi.hoisted(() => vi.fn());
vi.mock('./sim', () => ({ runConcurrentSim }));

const requestWithIterations = (iterations: number) => RaidSimRequest.create({ simOptions: SimOptions.create({ iterations }) });

const signals = { abort: { isTriggered: () => false } } as unknown as SimSignals;

const makePool = (statSimRequests: Array<{ requestLow?: RaidSimRequest; requestHigh: RaidSimRequest }>) =>
	({
		statWeightRequests: vi.fn(async () => ({
			baseRequest: requestWithIterations(100),
			epReferenceStat: 0,
			statSimRequests,
		})),
		statWeightCompute: vi.fn(async () => StatWeightsResult.create()),
	}) as unknown as WorkerPool;

describe('runConcurrentStatWeights', () => {
	beforeEach(() => {
		runConcurrentSim.mockReset();
		// Report one iteration per sim so the caller's progress handler runs and exposes the totals.
		runConcurrentSim.mockImplementation(async (request: RaidSimRequest, _pool: WorkerPool, onProgress: (pm: ProgressMetrics) => void) => {
			onProgress(ProgressMetrics.create({ completedIterations: request.simOptions!.iterations }));
			return RaidSimResult.create();
		});
	});

	it('skips the low sim, and its iterations, when requestLow is unset', async () => {
		const pool = makePool([{ requestHigh: requestWithIterations(30) }]);
		const progress: ProgressMetrics[] = [];

		await runConcurrentStatWeights({} as never, pool, pm => progress.push(pm), signals);

		// Baseline + high only; the absent low sim is neither run nor counted.
		expect(runConcurrentSim).toHaveBeenCalledTimes(2);
		expect(progress[0].totalSims).toBe(2);
		expect(progress[0].totalIterations).toBe(130);

		const calcRequest = vi.mocked(pool.statWeightCompute).mock.calls[0][0];
		expect(calcRequest.statSimResults[0].resultLow).toBeUndefined();
		expect(calcRequest.statSimResults[0].resultHigh).toBeDefined();
	});

	it('runs the low sim when requestLow is present', async () => {
		const pool = makePool([{ requestLow: requestWithIterations(20), requestHigh: requestWithIterations(30) }]);
		const progress: ProgressMetrics[] = [];

		await runConcurrentStatWeights({} as never, pool, pm => progress.push(pm), signals);

		expect(runConcurrentSim).toHaveBeenCalledTimes(3);
		expect(progress[0].totalSims).toBe(3);
		expect(progress[0].totalIterations).toBe(150);

		const calcRequest = vi.mocked(pool.statWeightCompute).mock.calls[0][0];
		expect(calcRequest.statSimResults[0].resultLow).toBeDefined();
	});

	describe('with a target ±', () => {
		const probe = 100;
		const seed = 7n;
		const unitStat = 1;
		const target = StatWeightsTarget.create({
			epCi: 0.01,
			maxIterationMultiplier: 3,
			metrics: [StatWeightsTargetMetric.StatWeightsTargetMetricDps, StatWeightsTargetMetric.StatWeightsTargetMetricHps],
		});
		const request = StatWeightsRequest.create({ target, hpsReferenceStat: Stat.StatIntellect });

		const makeTargetPool = (plan: (call: number) => StatWeightsTopUpPlan) => {
			let calls = 0;
			return {
				statWeightRequests: vi.fn(async () => ({
					baseRequest: RaidSimRequest.create({ simOptions: SimOptions.create({ iterations: probe, randomSeed: seed }) }),
					epReferenceStat: 0,
					hpsReferenceStat: Stat.StatIntellect,
					statSimRequests: [
						{
							statData: StatWeightsStatData.create({ unitStat }),
							requestLow: requestWithIterations(probe),
							requestHigh: requestWithIterations(probe),
						},
					],
				})),
				statWeightTopUpPlan: vi.fn(async () => plan(calls++)),
				statWeightCompute: vi.fn(async () => StatWeightsResult.create()),
			} as unknown as WorkerPool;
		};

		it('continues each planned run at the seed after its last iteration and adds the pass as a segment', async () => {
			const pool = makeTargetPool(call =>
				call === 0
					? StatWeightsTopUpPlan.create({ baselineIterations: 3 * probe, stats: [{ unitStat, iterations: 2 * probe }] })
					: StatWeightsTopUpPlan.create(),
			);
			const progress: ProgressMetrics[] = [];

			await runConcurrentStatWeights(request, pool, pm => progress.push(pm), signals);

			const topUpRequest = vi.mocked(pool.statWeightTopUpPlan).mock.calls[0][0];
			expect(topUpRequest.target).toEqual(target);
			expect(topUpRequest.calc!.hpsReferenceStat).toBe(Stat.StatIntellect);
			const topUps = runConcurrentSim.mock.calls.slice(3).map(([simRequest]) => (simRequest as RaidSimRequest).simOptions!);
			expect(topUps.map(options => options.iterations)).toEqual([2 * probe, probe, probe]);
			expect(topUps.map(options => options.randomSeed)).toEqual([seed + BigInt(probe), seed + BigInt(probe), seed + BigInt(probe)]);

			const calcRequest = vi.mocked(pool.statWeightCompute).mock.calls[0][0];
			expect(calcRequest.extraSegments).toHaveLength(1);
			expect(calcRequest.extraSegments[0].baseResult).toBeDefined();
			expect(calcRequest.extraSegments[0].statSimResults.map(result => result.statData!.unitStat)).toEqual([unitStat]);

			const last = progress.filter(pm => !pm.finalWeightResult).at(-1)!;
			expect(last.totalSims).toBe(6);
			expect(last.totalIterations).toBe(3 * probe + 2 * (2 * probe));
		});

		it('asks for a plan after every pass and stops at the first empty one', async () => {
			const passes = 2;
			const pool = makeTargetPool(call =>
				call < passes
					? StatWeightsTopUpPlan.create({ baselineIterations: (call + 2) * probe, stats: [{ unitStat, iterations: (call + 2) * probe }] })
					: StatWeightsTopUpPlan.create(),
			);

			await runConcurrentStatWeights(request, pool, noop, signals);

			expect(vi.mocked(pool.statWeightTopUpPlan).mock.calls).toHaveLength(passes + 1);
			expect(vi.mocked(pool.statWeightCompute).mock.calls[0][0].extraSegments).toHaveLength(passes);
		});
	});
});
