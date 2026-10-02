import { SimRun, SimRunData } from '@generated/proto/ui';
import { useSimHost } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { SimResult } from '@sim/proto/sim_result';
import { subscribeSimSettingsChange } from '@sim/state/subscriptions';
import { isDevMode } from '@sim/utils/env';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

import { trackEvent } from '../../../../tracking/utils';
import { useSimResult } from '../../hooks/useSimResult';
import type { SimResultsManager } from '../../model/results_manager';
import { AuraMetricsTable } from '../AuraMetricsTable';
import { CastMetricsTable } from '../CastMetricsTable';
import { DamageMetricsTable } from '../DamageMetricsTable';
import { DtpsMetricsTable } from '../DtpsMetricsTable';
import { HealingMetricsTable } from '../HealingMetricsTable';
import { LogRunner } from '../LogRunner';
import { ResourceMetricsTable } from '../ResourceMetricsTable';
import { ALL_UNITS, hasTarget, ResultsFilter, simResultFilter } from '../ResultsFilter';
import { ThreatMetricsTable } from '../ThreatMetricsTable';
import { Timeline } from '../Timeline';
import { DetailedResultsPane } from './DetailedResultsPane';
import { DetailedResultsView } from './DetailedResultsView';
import { DEFAULT_DETAILED_RESULTS_TAB, DETAILED_RESULTS_TABS } from './utils';

export interface DetailedResultsProps {
	resultsManager: SimResultsManager;
}

export const DetailedResults = ({ resultsManager }: DetailedResultsProps) => {
	const host = useSimHost();
	const sim = host.sim;
	const resultsEmitter = host.resultChannel;

	const settingsSubscribe = subscribeSimSettingsChange(sim);
	const { damage: showDamage, threat: showThreat, healing: showHealing } = useDisplayMetrics(sim);

	const visibleTabs = useMemo(
		() =>
			DETAILED_RESULTS_TABS.filter(tab => {
				if (tab.id === 'damageTab') return showDamage;
				if (tab.id === 'healingTab') return showHealing;
				if (tab.id === 'threatTab' || tab.id === 'damageTakenTab') return showThreat;
				return true;
			}),
		[showDamage, showThreat, showHealing],
	);

	const [activeId, setActiveId] = useState<string>(DEFAULT_DETAILED_RESULTS_TAB);
	const [deathDisabled, setDeathDisabled] = useState(true);
	const [target, setTarget] = useState(ALL_UNITS);
	const hasResults = useSimResult() !== null;

	// `updateResults` is bound to the emitter, not to the filter, so the selection it reads is a ref.
	const targetRef = useRef(target);
	// What the last emit already carried, so the reset below does not queue a second one.
	const emittedTarget = useRef(target);

	const latestRun = useRef<SimRunData | null>(null);
	const currentSimResult = useRef<SimResult | null>(null);
	const latestDeathSeeds = useRef<Array<bigint>>([]);
	const recentlyEditedSeed = useRef(false);

	const updateResults = useCallback(
		async (simRunData: SimRunData | null) => {
			if (simRunData?.run?.request?.requestId !== latestRun.current?.run?.request?.requestId) {
				latestRun.current = simRunData;
				currentSimResult.current = await SimResult.fromProto(simRunData?.run || SimRun.create());
			}

			const playerMetrics = latestRun.current?.run?.result?.raidMetrics?.parties.map(party => party.players).flat();

			if (isDevMode() && playerMetrics) {
				console.log('Found player metrics:');
				console.log(playerMetrics);
			}

			if (playerMetrics?.length) {
				const deathSeeds = playerMetrics[0].deathSeeds;

				if (isDevMode() && !!deathSeeds.length) {
					console.log('Found death seeds:');
					console.log(deathSeeds);
				}

				if (deathSeeds.length > 1 || latestDeathSeeds.current.length == 0) {
					latestDeathSeeds.current = deathSeeds;
				}
			}

			if (currentSimResult.current == null) {
				resultsEmitter.emit(null);
			} else {
				// A run with fewer targets than the last one drops the selection, and it happens here
				// rather than in the filter: the emit below is what every metrics table reads, and a
				// component-level effect would let one render through pointing at a target this
				// result does not have.
				if (!hasTarget(currentSimResult.current, targetRef.current)) {
					targetRef.current = ALL_UNITS;
					emittedTarget.current = ALL_UNITS;
					setTarget(ALL_UNITS);
				}
				resultsEmitter.emit({ result: currentSimResult.current, filter: simResultFilter(targetRef.current) });
			}
		},
		[resultsEmitter],
	);

	// The seed the "Simulate a Death" button pinned is released by the next settings change or run.
	const releaseEditedSeed = useCallback(() => {
		if (!recentlyEditedSeed.current) return;
		sim.setFixedRngSeed(0);
		recentlyEditedSeed.current = false;
	}, [sim]);

	useEffect(() => settingsSubscribe(releaseEditedSeed), [settingsSubscribe, releaseEditedSeed]);

	useEffect(
		() =>
			resultsManager.currentChangeEmitter.on(async () => {
				const runData = resultsManager.getRunData();
				if (runData) {
					releaseEditedSeed();
					await updateResults(runData);
				}
				setDeathDisabled(latestDeathSeeds.current.length < 2);
			}),
		[resultsManager, releaseEditedSeed, updateResults],
	);

	// Picking a target re-emits the last run under the new filter. Guarded on the value rather than on a mount flag, so a change
	// of `updateResults` cannot fire it on its own.
	useEffect(() => {
		if (emittedTarget.current === target) return;
		emittedTarget.current = target;
		void updateResults(latestRun.current);
	}, [target, updateResults]);

	useEffect(() => {
		if (visibleTabs.some(tab => tab.id === activeId)) return;
		setActiveId(visibleTabs[0]?.id ?? DEFAULT_DETAILED_RESULTS_TAB);
	}, [visibleTabs, activeId]);

	const onSimulateDeath = () => {
		trackEvent({ action: 'sim', category: 'simulate', label: 'death' });

		if (latestDeathSeeds.current.length > 1) {
			sim.setFixedRngSeed(Number(latestDeathSeeds.current.pop()));
			recentlyEditedSeed.current = true;

			if (isDevMode()) {
				console.log('Setting fixed seed:');
				console.log(sim.getFixedRngSeed());
			}
		}

		void host.runSingleIteration();
	};

	return (
		<DetailedResultsView
			tabs={visibleTabs}
			activeId={activeId}
			onActiveIdChange={setActiveId}
			hasResults={hasResults}
			iterationDisabled={host.disabled}
			deathDisabled={deathDisabled}
			onSimulateIteration={() => {
				trackEvent({ action: 'sim', category: 'simulate', label: 'once' });
				void host.runSingleIteration();
			}}
			onSimulateDeath={onSimulateDeath}
			targetPicker={
				<ResultsFilter
					target={target}
					onTargetChange={next => {
						targetRef.current = next;
						setTarget(next);
					}}
				/>
			}>
			{showDamage && (
				<DetailedResultsPane id="damageTab" className="[&_.ui-metrics-table]:text-xs" topline histogram>
					<DamageMetricsTable />
				</DetailedResultsPane>
			)}
			{showThreat && (
				<DetailedResultsPane id="threatTab" className="[&_.ui-metrics-table]:text-xs" topline>
					<ThreatMetricsTable />
				</DetailedResultsPane>
			)}
			{showHealing && (
				<DetailedResultsPane id="healingTab" className="[&_.ui-metrics-table]:text-xs" topline>
					<HealingMetricsTable />
				</DetailedResultsPane>
			)}
			{showThreat && (
				<DetailedResultsPane id="damageTakenTab" topline>
					<DtpsMetricsTable />
				</DetailedResultsPane>
			)}
			<DetailedResultsPane id="buffsTab">
				<AuraMetricsTable useDebuffs={false} />
			</DetailedResultsPane>
			<DetailedResultsPane id="debuffsTab">
				<AuraMetricsTable useDebuffs={true} />
			</DetailedResultsPane>
			<DetailedResultsPane id="castsTab">
				<CastMetricsTable />
			</DetailedResultsPane>
			<DetailedResultsPane id="resourcesTab">
				<ResourceMetricsTable />
			</DetailedResultsPane>
			<DetailedResultsPane id="timelineTab" filling>
				<Timeline active={activeId === 'timelineTab'} />
			</DetailedResultsPane>
			<DetailedResultsPane id="logTab">
				<LogRunner active={activeId === 'logTab'} />
			</DetailedResultsPane>
		</DetailedResultsView>
	);
};
