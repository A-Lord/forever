import {
	ErrorOutcome,
	ErrorOutcomeType,
	ProgressMetrics,
	RaidSimRequest,
	RaidSimResult,
	StatWeightsCalcRequest,
	StatWeightsPairedSums,
	StatWeightsRequest,
	StatWeightsResult,
	StatWeightsSidePairs,
	StatWeightsStatPairs,
	StatWeightsTopUpRequest,
} from '@generated/proto/api';
import { SimRequest } from '@worker/types';

import { SimSignals } from '../sim_signal_manager';
import { isDevMode } from '../utils/env';
import { generateRequestId, WorkerPool, WorkerProgressCallback } from '../workers/worker_pool';
import { runConcurrentSim } from './sim';

const makeAndSendWeightsError = (err: string | ErrorOutcome, onProgress: WorkerProgressCallback): StatWeightsResult => {
	const errRes = StatWeightsResult.create();
	if (typeof err === 'string') {
		console.error(err);
		errRes.error = ErrorOutcome.create({ message: err });
	} else {
		if (err.message) console.error(err.message);
		errRes.error = err;
	}
	onProgress(ProgressMetrics.create({ finalWeightResult: errRes }));
	return errRes;
};

const firstPlayer = (result: RaidSimResult) => result.raidMetrics!.parties[0].players[0];

// The first player's per-iteration DPS, HPS, threat, DTPS and TMI of a run.
const runValues = (result: RaidSimResult): number[][] => {
	const player = firstPlayer(result);
	return [player.dps!, player.hps!, player.threat!, player.dtps!, player.tmi!].map(metric => metric.allValues);
};

// Pairs each iteration of a stat run with the baseline's iteration offset + i, which ran at the same
// seed, in the same order of operations as pairStatWeightRun in sim/core/statweight.go.
const pairRun = (baseline: number[][], result: RaidSimResult, offset: number): StatWeightsSidePairs => {
	const [dps, hps, tps, dtps, tmi] = runValues(result).map((values, metric) => {
		let sum = 0;
		let sumSq = 0;
		for (let i = 0; i < values.length; i++) {
			const diff = values[i] - baseline[metric][offset + i];
			sum += diff;
			sumSq += diff * diff;
		}
		return StatWeightsPairedSums.create({ n: values.length, sum, sumSq });
	});
	return StatWeightsSidePairs.create({ dps, hps, tps, dtps, tmi, chanceOfDeath: firstPlayer(result).chanceOfDeath });
};

// Adds a later pass's pairs to a side's. The chance of death stays the first run's.
const addPairs = (side: StatWeightsSidePairs, pass: StatWeightsSidePairs) => {
	for (const metric of ['dps', 'hps', 'tps', 'dtps', 'tmi'] as const) {
		side[metric]!.n += pass[metric]!.n;
		side[metric]!.sum += pass[metric]!.sum;
		side[metric]!.sumSq += pass[metric]!.sumSq;
	}
};

export const runConcurrentStatWeights = async (
	request: StatWeightsRequest,
	workerPool: WorkerPool,
	onProgress: WorkerProgressCallback,
	signals: SimSignals,
): Promise<StatWeightsResult> => {
	if (isDevMode()) {
		console.log('Getting stat weight sim requests.');
	}

	const newRaidSimRequestId = () => generateRequestId(SimRequest.raidSimAsync);

	const aborted = () => makeAndSendWeightsError(ErrorOutcome.create({ type: ErrorOutcomeType.ErrorOutcomeAborted }), onProgress);

	const manualResponse = await workerPool.statWeightRequests(request);
	manualResponse.baseRequest!.requestId = newRaidSimRequestId();

	if (signals.abort.isTriggered()) return aborted();

	let iterationsTotal = manualResponse.baseRequest!.simOptions!.iterations;
	let iterationsDone = 0;
	let simsTotal = 1;
	let simsDone = 0;

	// The backend leaves RequestLow unset for the school-split hit/crit stats, whose low-side
	// modifier is always zero (sim/core/statweight.go). Skip the low sim when it is absent rather
	// than counting iterations for a request that does not exist.
	for (const statReqData of manualResponse.statSimRequests) {
		if (statReqData.requestLow) {
			iterationsTotal += statReqData.requestLow.simOptions!.iterations;
			simsTotal += 1;
		}
		iterationsTotal += statReqData.requestHigh!.simOptions!.iterations;
		simsTotal += 1;
	}

	if (isDevMode()) {
		console.log(`Need to run a total of ${simsTotal} sims and ${iterationsTotal} iterations.`);
	}

	let lastIterations = 0;
	const progressHandler = (pm: ProgressMetrics) => {
		iterationsDone += pm.completedIterations - lastIterations;
		lastIterations = pm.completedIterations;

		onProgress(
			ProgressMetrics.create({
				totalIterations: iterationsTotal,
				completedIterations: iterationsDone,
				totalSims: simsTotal,
				completedSims: simsDone,
			}),
		);

		if (pm.finalRaidResult) simsDone++;
	};

	// Runs a sim and keeps only what reduce takes from its result, so no whole result outlives the call.
	const runReduced = async <T>(
		simRequest: RaidSimRequest,
		reduce: (result: RaidSimResult) => T,
	): Promise<{ error: ErrorOutcome } | { error?: undefined; value: T }> => {
		const result = await runConcurrentSim(simRequest, workerPool, progressHandler, signals);
		return result.error ? { error: result.error } : { value: reduce(result) };
	};

	const baseLine = await runReduced(manualResponse.baseRequest!, result => ({
		values: runValues(result),
		chanceOfDeath: firstPlayer(result).chanceOfDeath,
	}));
	if (baseLine.error) return makeAndSendWeightsError(baseLine.error, onProgress);

	// Only the baseline keeps its per-iteration values, for the stat runs of later passes to pair with.
	const baselineValues = baseLine.value.values;
	const calcRequest = StatWeightsCalcRequest.create({
		epReferenceStat: manualResponse.epReferenceStat,
		hpsReferenceStat: manualResponse.hpsReferenceStat,
		probeIterations: baselineValues[0].length,
		baselineIterations: baselineValues[0].length,
		baseChanceOfDeath: baseLine.value.chanceOfDeath,
	});
	const statPairs = new Map<number, StatWeightsStatPairs>();

	for (const statReqData of manualResponse.statSimRequests) {
		if (signals.abort.isTriggered()) return aborted();

		const pairs = StatWeightsStatPairs.create({ statData: statReqData.statData });
		lastIterations = 0;
		if (statReqData.requestLow) {
			statReqData.requestLow.requestId = newRaidSimRequestId();
			const low = await runReduced(statReqData.requestLow, result => pairRun(baselineValues, result, 0));
			if (low.error) return makeAndSendWeightsError(low.error, onProgress);
			pairs.low = low.value;
		}

		lastIterations = 0;
		statReqData.requestHigh!.requestId = newRaidSimRequestId();
		const high = await runReduced(statReqData.requestHigh!, result => pairRun(baselineValues, result, 0));
		if (high.error) return makeAndSendWeightsError(high.error, onProgress);
		pairs.high = high.value;

		calcRequest.statPairs.push(pairs);
		statPairs.set(statReqData.statData!.unitStat, pairs);
	}

	if (request.target) {
		const baseRequest = manualResponse.baseRequest!;
		const seed = baseRequest.simOptions!.randomSeed;
		const statRequests = new Map(manualResponse.statSimRequests.map(statReqData => [statReqData.statData!.unitStat, statReqData]));

		// Iteration i runs at seed randomSeed + i, so a run continued at the seed after its last iteration
		// repeats what one longer run would have simulated.
		const continueRun = <T>(simRequest: RaidSimRequest, done: number, iterations: number, reduce: (result: RaidSimResult) => T) => {
			lastIterations = 0;
			return runReduced(
				{
					...simRequest,
					requestId: newRaidSimRequestId(),
					simOptions: { ...simRequest.simOptions!, iterations, randomSeed: seed + BigInt(done) },
				},
				reduce,
			);
		};

		for (let passes = 0; ; passes++) {
			const plan = await workerPool.statWeightTopUpPlan(StatWeightsTopUpRequest.create({ calc: calcRequest, target: request.target, passes }));
			if (!plan.stats.length) break;

			const baselineIterations = calcRequest.baselineIterations;
			const baselineDelta = plan.baselineIterations - baselineIterations;
			if (baselineDelta > 0) {
				iterationsTotal += baselineDelta;
				simsTotal += 1;
			}
			for (const planned of plan.stats) {
				const runs = statRequests.get(planned.unitStat)!.requestLow ? 2 : 1;
				iterationsTotal += runs * (planned.iterations - statPairs.get(planned.unitStat)!.high!.dps!.n);
				simsTotal += runs;
			}

			if (baselineDelta > 0) {
				if (signals.abort.isTriggered()) return aborted();
				const base = await continueRun(baseRequest, baselineIterations, baselineDelta, result => {
					runValues(result).forEach((values, metric) => {
						for (const value of values) baselineValues[metric].push(value);
					});
				});
				if (base.error) return makeAndSendWeightsError(base.error, onProgress);
				calcRequest.baselineIterations = baselineValues[0].length;
			}

			for (const planned of plan.stats) {
				if (signals.abort.isTriggered()) return aborted();

				const statReqData = statRequests.get(planned.unitStat)!;
				const pairs = statPairs.get(planned.unitStat)!;
				const done = pairs.high!.dps!.n;
				const delta = planned.iterations - done;

				if (statReqData.requestLow) {
					const low = await continueRun(statReqData.requestLow, done, delta, result => pairRun(baselineValues, result, done));
					if (low.error) return makeAndSendWeightsError(low.error, onProgress);
					addPairs(pairs.low!, low.value);
				}
				const high = await continueRun(statReqData.requestHigh!, done, delta, result => pairRun(baselineValues, result, done));
				if (high.error) return makeAndSendWeightsError(high.error, onProgress);
				addPairs(pairs.high!, high.value);
			}
		}
	}

	if (isDevMode()) {
		console.log(`All ${simsTotal} sims finished successfully. Computing weights.`);
	}

	const weightResult = await workerPool.statWeightCompute(calcRequest);
	if (weightResult.error) return makeAndSendWeightsError(weightResult.error, onProgress);
	onProgress(ProgressMetrics.create({ finalWeightResult: weightResult }));
	return weightResult;
};
