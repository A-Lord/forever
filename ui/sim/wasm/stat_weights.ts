import {
	ErrorOutcome,
	ErrorOutcomeType,
	ProgressMetrics,
	RaidSimRequest,
	RaidSimResult,
	StatWeightsCalcRequest,
	StatWeightsCalcSegment,
	StatWeightsRequest,
	StatWeightsResult,
	StatWeightsStatResultData,
	StatWeightsTopUpRequest,
} from '@generated/proto/api';
import { SimRequest } from '@worker/types';

import { SimSignals } from '../sim_signal_manager';
import { isDevMode } from '../utils/env';
import { generateRequestId, WorkerPool, WorkerProgressCallback } from '../workers/worker_pool';
import { runConcurrentSim } from './sim';

// The defaults runStatWeights in sim/core/statweight.go uses: the cap as a multiple of the first
// pass's iterations, and the first top-up plus two adaptive passes.
const DEFAULT_MAX_ITERATION_MULTIPLIER = 8;
const MAX_TOP_UP_PASSES = 3;

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

	const manualResponse = await workerPool.statWeightRequests(request);
	manualResponse.baseRequest!.requestId = newRaidSimRequestId();

	if (signals.abort.isTriggered()) {
		return makeAndSendWeightsError(ErrorOutcome.create({ type: ErrorOutcomeType.ErrorOutcomeAborted }), onProgress);
	}

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

	const baseLine = await runConcurrentSim(manualResponse.baseRequest!, workerPool, progressHandler, signals);
	if (baseLine.error) return makeAndSendWeightsError(baseLine.error, onProgress);

	const calcRequest = StatWeightsCalcRequest.create({
		baseResult: baseLine,
		epReferenceStat: manualResponse.epReferenceStat,
		statSimResults: [],
	});

	for (const statReqData of manualResponse.statSimRequests) {
		if (signals.abort.isTriggered()) return makeAndSendWeightsError(ErrorOutcome.create({ type: ErrorOutcomeType.ErrorOutcomeAborted }), onProgress);

		lastIterations = 0;
		let lowRes: RaidSimResult | undefined;
		if (statReqData.requestLow) {
			statReqData.requestLow.requestId = newRaidSimRequestId();
			lowRes = await runConcurrentSim(statReqData.requestLow, workerPool, progressHandler, signals);
			if (lowRes.error) return makeAndSendWeightsError(lowRes.error, onProgress);
		}

		lastIterations = 0;
		statReqData.requestHigh!.requestId = newRaidSimRequestId();
		const highRes = await runConcurrentSim(statReqData.requestHigh!, workerPool, progressHandler, signals);
		if (highRes.error) return makeAndSendWeightsError(highRes.error, onProgress);

		calcRequest.statSimResults.push(
			StatWeightsStatResultData.create({
				statData: statReqData.statData,
				resultLow: lowRes,
				resultHigh: highRes,
			}),
		);
	}

	if (request.targetEpCi > 0) {
		const baseRequest = manualResponse.baseRequest!;
		const probeIterations = baseRequest.simOptions!.iterations;
		const seed = baseRequest.simOptions!.randomSeed;
		const maxIterationsPerSide = probeIterations * (request.maxIterationMultiplier || DEFAULT_MAX_ITERATION_MULTIPLIER);
		const statRequests = new Map(manualResponse.statSimRequests.map(statReqData => [statReqData.statData!.unitStat, statReqData]));
		const statIterations = new Map(manualResponse.statSimRequests.map(statReqData => [statReqData.statData!.unitStat, probeIterations]));
		let baselineIterations = probeIterations;

		// Iteration i runs at seed randomSeed + i, so a run continued at the seed after its last iteration
		// repeats what one longer run would have simulated.
		const continueRun = (simRequest: RaidSimRequest, done: number, iterations: number) => {
			const continued = RaidSimRequest.clone(simRequest);
			continued.simOptions!.iterations = iterations;
			continued.simOptions!.randomSeed = seed + BigInt(done);
			continued.requestId = newRaidSimRequestId();
			lastIterations = 0;
			return runConcurrentSim(continued, workerPool, progressHandler, signals);
		};

		for (let pass = 0; pass < MAX_TOP_UP_PASSES; pass++) {
			const plan = await workerPool.statWeightTopUpPlan(
				StatWeightsTopUpRequest.create({ calc: calcRequest, targetEpCi: request.targetEpCi, maxIterationsPerSide }),
			);
			if (!plan.stats.length) break;

			const baselineDelta = plan.baselineIterations - baselineIterations;
			if (baselineDelta > 0) {
				iterationsTotal += baselineDelta;
				simsTotal += 1;
			}
			for (const planned of plan.stats) {
				const runs = statRequests.get(planned.unitStat)!.requestLow ? 2 : 1;
				iterationsTotal += runs * (planned.iterations - statIterations.get(planned.unitStat)!);
				simsTotal += runs;
			}

			const segment = StatWeightsCalcSegment.create();
			if (baselineDelta > 0) {
				if (signals.abort.isTriggered())
					return makeAndSendWeightsError(ErrorOutcome.create({ type: ErrorOutcomeType.ErrorOutcomeAborted }), onProgress);
				const baseRes = await continueRun(baseRequest, baselineIterations, baselineDelta);
				if (baseRes.error) return makeAndSendWeightsError(baseRes.error, onProgress);
				segment.baseResult = baseRes;
				baselineIterations = plan.baselineIterations;
			}

			for (const planned of plan.stats) {
				if (signals.abort.isTriggered())
					return makeAndSendWeightsError(ErrorOutcome.create({ type: ErrorOutcomeType.ErrorOutcomeAborted }), onProgress);

				const statReqData = statRequests.get(planned.unitStat)!;
				const done = statIterations.get(planned.unitStat)!;
				const delta = planned.iterations - done;

				let lowRes: RaidSimResult | undefined;
				if (statReqData.requestLow) {
					lowRes = await continueRun(statReqData.requestLow, done, delta);
					if (lowRes.error) return makeAndSendWeightsError(lowRes.error, onProgress);
				}
				const highRes = await continueRun(statReqData.requestHigh!, done, delta);
				if (highRes.error) return makeAndSendWeightsError(highRes.error, onProgress);

				segment.statSimResults.push(
					StatWeightsStatResultData.create({
						statData: statReqData.statData,
						resultLow: lowRes,
						resultHigh: highRes,
					}),
				);
				statIterations.set(planned.unitStat, planned.iterations);
			}
			calcRequest.extraSegments.push(segment);
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
