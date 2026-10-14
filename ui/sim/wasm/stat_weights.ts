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
import { runConcurrentSimQueue } from './sim';

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

	// One per sim, as the sims of a pass run at once.
	const progressHandler = () => {
		let lastIterations = 0;
		return (pm: ProgressMetrics) => {
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
	};

	// Runs the sims as one queue of splits across the workers. Each hands its result to `use` as it ends, so no whole result
	// outlives its sim, and as each `use` writes its own slot the values do not depend on which sim ends first.
	const runSims = (sims: Array<{ request: RaidSimRequest; use: (result: RaidSimResult) => void }>) =>
		runConcurrentSimQueue(
			sims.map(({ request, use }) => ({ request, onProgress: progressHandler(), onResult: use })),
			workerPool,
			signals,
		);

	let baseLine = { values: [] as number[][], chanceOfDeath: 0 };
	const baseError = await runSims([
		{
			request: manualResponse.baseRequest!,
			use: result => (baseLine = { values: runValues(result), chanceOfDeath: firstPlayer(result).chanceOfDeath }),
		},
	]);
	if (baseError) return makeAndSendWeightsError(baseError, onProgress);

	// Only the baseline keeps its per-iteration values, for the stat runs of later passes to pair with.
	const baselineValues = baseLine.values;
	const calcRequest = StatWeightsCalcRequest.create({
		epReferenceStat: manualResponse.epReferenceStat,
		hpsReferenceStat: manualResponse.hpsReferenceStat,
		probeIterations: baselineValues[0].length,
		baselineIterations: baselineValues[0].length,
		baseChanceOfDeath: baseLine.chanceOfDeath,
	});
	const statPairs = new Map<number, StatWeightsStatPairs>();

	const paired = (request: RaidSimRequest, offset: number, use: (pairs: StatWeightsSidePairs) => void) => ({
		request,
		use: (result: RaidSimResult) => use(pairRun(baselineValues, result, offset)),
	});

	const statSims = manualResponse.statSimRequests.flatMap(statReqData => {
		const pairs = StatWeightsStatPairs.create({ statData: statReqData.statData });
		calcRequest.statPairs.push(pairs);
		statPairs.set(statReqData.statData!.unitStat, pairs);
		statReqData.requestHigh!.requestId = newRaidSimRequestId();
		const high = paired(statReqData.requestHigh!, 0, high => (pairs.high = high));
		if (!statReqData.requestLow) return [high];
		statReqData.requestLow.requestId = newRaidSimRequestId();
		return [paired(statReqData.requestLow, 0, low => (pairs.low = low)), high];
	});
	const statError = await runSims(statSims);
	if (statError) return makeAndSendWeightsError(statError, onProgress);

	if (request.target) {
		const baseRequest = manualResponse.baseRequest!;
		const seed = baseRequest.simOptions!.randomSeed;
		const statRequests = new Map(manualResponse.statSimRequests.map(statReqData => [statReqData.statData!.unitStat, statReqData]));

		// Iteration i runs at seed randomSeed + i, so a run continued at the seed after its last iteration
		// repeats what one longer run would have simulated.
		const continued = (simRequest: RaidSimRequest, done: number, iterations: number): RaidSimRequest => ({
			...simRequest,
			requestId: newRaidSimRequestId(),
			simOptions: { ...simRequest.simOptions!, iterations, randomSeed: seed + BigInt(done) },
		});

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
				const baseError = await runSims([
					{
						request: continued(baseRequest, baselineIterations, baselineDelta),
						use: result =>
							runValues(result).forEach((values, metric) => {
								for (const value of values) baselineValues[metric].push(value);
							}),
					},
				]);
				if (baseError) return makeAndSendWeightsError(baseError, onProgress);
				calcRequest.baselineIterations = baselineValues[0].length;
			}

			const topUps = plan.stats.flatMap(planned => {
				const statReqData = statRequests.get(planned.unitStat)!;
				const pairs = statPairs.get(planned.unitStat)!;
				const done = pairs.high!.dps!.n;
				const delta = planned.iterations - done;
				const high = paired(continued(statReqData.requestHigh!, done, delta), done, high => addPairs(pairs.high!, high));
				if (!statReqData.requestLow) return [high];
				return [paired(continued(statReqData.requestLow, done, delta), done, low => addPairs(pairs.low!, low)), high];
			});
			const topUpError = await runSims(topUps);
			if (topUpError) return makeAndSendWeightsError(topUpError, onProgress);
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
