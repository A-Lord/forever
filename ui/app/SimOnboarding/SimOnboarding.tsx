import { useActivateTab } from '@features/hooks/useActivateTab';
import { applicableSteps, OnboardingProvider, OnboardingTour, type SimulationState, type TourEnd, type TourRun, type TourStep } from '@features/onboarding';
import { ResultsPanelStage } from '@features/results/components/SimResultsPanel/results_panel_store';
import i18n from '@i18n/config';
import type { Sim } from '@sim/sim';
import { type ReactNode, useCallback, useEffect, useMemo, useState, useSyncExternalStore } from 'react';

import { trackEvent } from '../../tracking/utils';
import type { SimHostObject } from '../individual_sim_ui';
import { ONBOARDING_STORAGE_KEY } from '../storage_keys';
import { SIM_TOUR_STEPS, type SimTourContext } from './sim_tour';

export interface SimOnboardingProps {
	/** Null until the shell has built it; the tour waits for it. */
	host: SimHostObject<any> | null;
	sim: Sim;
	/** The sim has loaded; the tour waits for it. */
	simReady: boolean;
	children: ReactNode;
}

/**
 * The spec page's guided tour: opens by itself the first time a player opens any spec page, once the
 * page is ready, and again from the toolbar (`useOnboardingRestart`). Inside `SimTabs`, which it opens
 * tabs through.
 */
export const SimOnboarding = ({ host, sim, simReady, children }: SimOnboardingProps) => {
	const ready = simReady && !!host;
	const renderTour = useCallback((run: TourRun) => host && <SimTour host={host} run={run} />, [host]);
	return (
		<OnboardingProvider ready={ready} storage={sim.env.storage} storageKey={ONBOARDING_STORAGE_KEY} renderTour={renderTour}>
			{children}
		</OnboardingProvider>
	);
};

const text = (step: TourStep<SimTourContext>) => ({
	title: i18n.t(`onboarding.steps.${step.id}.title`),
	body: i18n.t(`onboarding.steps.${step.id}.body`),
});

const simulationState = (stage: ResultsPanelStage): SimulationState =>
	stage === ResultsPanelStage.Pending || stage === ResultsPanelStage.Running ? 'running' : stage === ResultsPanelStage.Result ? 'done' : 'idle';

// The tab strip's open tab, whose test id is the tab's id.
const openTab = () => document.querySelector('[data-testid="sim-tabs"] [aria-selected="true"]')?.getAttribute('data-testid') ?? null;

const SimTour = ({ host, run }: { host: SimHostObject<any>; run: TourRun }) => {
	const activateTab = useActivateTab();
	const stage = useSyncExternalStore(host.resultsPanel.subscribe, host.resultsPanel.getStage);
	const context = useMemo((): SimTourContext => ({ canSimulate: !host.simDisabled, hasResult: stage === ResultsPanelStage.Result }), [host, stage]);
	const steps = useMemo(() => applicableSteps(SIM_TOUR_STEPS, context), [context]);
	// The tour opens tabs as it goes; the player gets back the one they were on.
	const [startTab] = useState(openTab);

	useEffect(() => {
		if (run.trigger === 'restart') trackEvent({ action: 'click', category: 'onboarding', label: 'start' });
	}, [run.trigger]);

	const prepare = useCallback(
		(step: TourStep<SimTourContext>) => {
			if (step.tab) activateTab(step.tab);
		},
		[activateTab],
	);

	const onEnd = useCallback(
		({ outcome, step, index, total }: TourEnd<TourStep<SimTourContext>>) => {
			trackEvent({
				action: 'click',
				category: 'onboarding',
				label: outcome,
				value: step.id,
				additionalData: { trigger: run.trigger, step: index + 1, steps: total },
			});
			if (startTab && startTab !== openTab()) activateTab(startTab);
			run.end(outcome);
		},
		[activateTab, run, startTab],
	);

	return <OnboardingTour steps={steps} text={text} prepare={prepare} simulation={simulationState(stage)} onEnd={onEnd} />;
};
