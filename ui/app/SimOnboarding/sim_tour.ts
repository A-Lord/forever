import { BULK_TAB_ID, RESULTS_TAB_ID } from '@features/hooks/useActivateTab';
import type { TourStep } from '@features/onboarding';

/** What the spec page offers, so a step with nothing to show is left out of the tour and its count. */
export interface SimTourContext {
	/** A gear planner has no Simulate button, rotation, results or batch sim. */
	canSimulate: boolean;
	/** A finished run to point at. */
	hasResult: boolean;
}

const tabButton = (id: string) => `[data-testid="sim-tabs"] [data-testid="${id}"]`;
const canSimulate = (context: SimTourContext) => context.canSimulate;

/** The spec page's tour, in order. Each step's words are `onboarding.steps.<id>.title` and `.body`. */
export const SIM_TOUR_STEPS: ReadonlyArray<TourStep<SimTourContext>> = [
	{ id: 'welcome' },
	{ id: 'spec', target: '[data-testid="sim-title-dropdown-root"]' },
	{ id: 'gear_tab', tab: 'gear-tab', target: tabButton('gear-tab') },
	{ id: 'gear', tab: 'gear-tab', target: '[data-testid="gear-picker-root"]' },
	{ id: 'presets', tab: 'gear-tab', target: '[data-testid="gear-presets"]' },
	{ id: 'import_export', target: '[data-testid="import-export"]' },
	{ id: 'settings_tab', tab: 'settings-tab', target: tabButton('settings-tab') },
	{ id: 'talents_tab', tab: 'talents-tab', target: tabButton('talents-tab') },
	{ id: 'rotation_tab', tab: 'rotation-tab', target: tabButton('rotation-tab'), applies: canSimulate },
	{ id: 'simulate', target: '[data-testid="dps-action"]', waitFor: 'simulation', applies: canSimulate },
	{ id: 'result', target: '[data-testid="results-sim"]', applies: context => context.canSimulate && context.hasResult },
	{ id: 'results_tab', tab: RESULTS_TAB_ID, target: tabButton(RESULTS_TAB_ID), applies: canSimulate },
	{ id: 'bulk_tab', tab: BULK_TAB_ID, target: tabButton(BULK_TAB_ID), applies: canSimulate },
	{ id: 'stat_weights', target: '[data-testid="ep-weights-action"]' },
	{ id: 'options', target: '[data-testid="sim-toolbar"]' },
	{ id: 'finish' },
];
