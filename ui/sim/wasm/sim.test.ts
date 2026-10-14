import {
	ErrorOutcome,
	ErrorOutcomeType,
	ProgressMetrics,
	RaidSimRequest,
	RaidSimRequestSplitRequest,
	RaidSimRequestSplitResult,
	RaidSimResult,
	RaidSimResultCombinationRequest,
	SimOptions,
} from '@generated/proto/api';
import { describe, expect, it, vi } from 'vitest';

import type { SimSignals } from '../sim_signal_manager';
import { noop } from '../utils/misc';
import type { WorkerPool } from '../workers/worker_pool';
import { type QueuedSim, runConcurrentSim, runConcurrentSimQueue } from './sim';

const makeSignals = () => {
	let triggered = false;
	return {
		abort: {
			isTriggered: () => triggered,
			trigger: async () => {
				triggered = true;
			},
		},
	} as unknown as SimSignals;
};

// Splits a request in two halves under its own id, as the sim does.
const splitInTwo = async ({ request }: RaidSimRequestSplitRequest) =>
	RaidSimRequestSplitResult.create({
		splitsDone: 2,
		requests: [0, 1].map(() => RaidSimRequest.create({ ...request, simOptions: SimOptions.create({ iterations: request!.simOptions!.iterations / 2 }) })),
	});

// A pool of two workers whose splits end when the test ends them, in the order it chooses.
const makeHeldPool = () => {
	const held: Array<{ request: RaidSimRequest; end: (result: RaidSimResult) => void }> = [];
	const pool = {
		getNumWorkers: () => 2,
		raidSimRequestSplit: vi.fn(splitInTwo),
		raidSimAsync: vi.fn((request: RaidSimRequest) => new Promise<RaidSimResult>(end => held.push({ request, end }))),
		raidSimResultCombination: vi.fn(async ({ results }: RaidSimResultCombinationRequest) =>
			RaidSimResult.create({
				iterationsDone: results.reduce((total, result) => total + result.iterationsDone, 0),
				avgIterationDuration: results[0].iterationsDone,
			}),
		),
	} as unknown as WorkerPool;
	return { pool, held };
};

const queuedSim = (iterations: number, requestId: string) => ({
	request: RaidSimRequest.create({ requestId, simOptions: SimOptions.create({ iterations }) }),
	onProgress: vi.fn(),
	onResult: vi.fn(),
});

describe('runConcurrentSim', () => {
	it('sends each split under its own id, as one worker can be handed two splits of a sim', async () => {
		const split = RaidSimRequest.create({ requestId: 'raidSimAsync-1', simOptions: SimOptions.create({ iterations: 50 }) });
		const pool = {
			getNumWorkers: () => 2,
			raidSimRequestSplit: vi.fn(async () => RaidSimRequestSplitResult.create({ splitsDone: 2, requests: [split, RaidSimRequest.clone(split)] })),
			raidSimAsync: vi.fn(async (_request: RaidSimRequest, onProgress: (pm: ProgressMetrics) => void) => {
				const result = RaidSimResult.create();
				onProgress(ProgressMetrics.create({ completedIterations: 50, finalRaidResult: result }));
				return result;
			}),
			raidSimResultCombination: vi.fn(async () => RaidSimResult.create()),
		} as unknown as WorkerPool;

		await runConcurrentSim(RaidSimRequest.create({ simOptions: SimOptions.create({ iterations: 100 }) }), pool, noop, makeSignals());

		const ids = vi.mocked(pool.raidSimAsync).mock.calls.map(([request]) => request.requestId);
		expect(ids).toHaveLength(2);
		expect(new Set(ids).size).toBe(2);
		expect(ids.every(id => id.startsWith('raidSimAsync-1'))).toBe(true);
	});
});

describe('runConcurrentSimQueue', () => {
	it('keeps one split on each worker and sends the next, of the same sim or the next, as one comes back', async () => {
		const { pool, held } = makeHeldPool();
		const sims: QueuedSim[] = [queuedSim(100, 'a'), queuedSim(60, 'b')];

		const queue = runConcurrentSimQueue(sims, pool, makeSignals());
		await vi.waitFor(() => expect(held).toHaveLength(2));
		expect(held.map(({ request }) => request.requestId)).toEqual(['a-0', 'a-1']);

		held[1].end(RaidSimResult.create({ iterationsDone: 50 }));
		await vi.waitFor(() => expect(held).toHaveLength(3));
		expect(held[2].request.requestId).toBe('b-0');

		held[0].end(RaidSimResult.create({ iterationsDone: 50 }));
		await vi.waitFor(() => expect(held).toHaveLength(4));
		held[3].end(RaidSimResult.create({ iterationsDone: 30 }));
		held[2].end(RaidSimResult.create({ iterationsDone: 30 }));

		expect(await queue).toBeUndefined();
		expect(vi.mocked(pool.raidSimAsync).mock.calls).toHaveLength(4);
	});

	it('combines each sim from its splits in split order and hands the result over with a final progress', async () => {
		const { pool, held } = makeHeldPool();
		const sim = queuedSim(100, 'a');

		const queue = runConcurrentSimQueue([sim], pool, makeSignals());
		await vi.waitFor(() => expect(held).toHaveLength(2));
		held[1].end(RaidSimResult.create({ iterationsDone: 49 }));
		held[0].end(RaidSimResult.create({ iterationsDone: 51 }));
		await queue;

		const combined = vi.mocked(pool.raidSimResultCombination).mock.calls[0][0].results;
		expect(combined.map(result => result.iterationsDone)).toEqual([51, 49]);
		expect(sim.onResult).toHaveBeenCalledWith(RaidSimResult.create({ iterationsDone: 100, avgIterationDuration: 51 }));
		expect(sim.onProgress.mock.calls.at(-1)![0].finalRaidResult).toBeDefined();
	});

	it('stops at a failure and reports it over the aborts it sets off', async () => {
		const { pool, held } = makeHeldPool();
		const signals = makeSignals();
		const failure = ErrorOutcome.create({ type: ErrorOutcomeType.ErrorOutcomeError, message: 'failed' });

		const queue = runConcurrentSimQueue([queuedSim(100, 'a'), queuedSim(100, 'b')], pool, signals);
		await vi.waitFor(() => expect(held).toHaveLength(2));
		held[1].end(RaidSimResult.create({ error: failure }));
		held[0].end(RaidSimResult.create({ error: ErrorOutcome.create({ type: ErrorOutcomeType.ErrorOutcomeAborted }) }));

		expect(await queue).toEqual(failure);
		expect(signals.abort.isTriggered()).toBe(true);
		expect(vi.mocked(pool.raidSimAsync).mock.calls).toHaveLength(2);
		expect(pool.raidSimResultCombination).not.toHaveBeenCalled();
	});
});
