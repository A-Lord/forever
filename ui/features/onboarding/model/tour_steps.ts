import type { Env } from '@sim/state/env';

/** What a step points at: a CSS selector, or a lookup for what a selector cannot say. */
export type TourTarget = string | ((root: ParentNode) => Element | null);

/**
 * One card of a tour. A step with no target sits in the middle of the page (the welcome and the last
 * card); `Context` is what the page tells the tour about itself, for `applies`.
 */
export interface TourStep<Context = unknown> {
	id: string;
	target?: TourTarget;
	/** The tab to open before the target is looked for. */
	tab?: string;
	/** The card waits for a simulation to finish, then moves on by itself. */
	waitFor?: 'simulation';
	/** False leaves the step out of the tour and its count (a gear planner has no Simulate button). */
	applies?: (context: Context) => boolean;
}

export const applicableSteps = <S extends TourStep<Context>, Context>(steps: ReadonlyArray<S>, context: Context): ReadonlyArray<S> =>
	steps.filter(step => !step.applies || step.applies(context));

/** The card's place in the tour: the first and last centred cards open and close it and are not counted. */
export type TourCardKind = 'welcome' | 'step' | 'finish';

export const cardKindAt = (steps: ReadonlyArray<TourStep<never>>, index: number): TourCardKind => {
	const step = steps[index];
	if (step.target) return 'step';
	if (index === 0) return 'welcome';
	if (index === steps.length - 1) return 'finish';
	return 'step';
};

/** "3 of 12": the counted steps, those between the welcome and the last card. */
export const stepCount = (steps: ReadonlyArray<TourStep<never>>, index: number): { position: number; total: number } => {
	const counted = steps.map((_, i) => i).filter(i => cardKindAt(steps, i) === 'step');
	return { position: counted.indexOf(index) + 1, total: counted.length };
};

/** How a tour ended: every card seen, or left early. Either way it does not open by itself again. */
export type TourOutcome = 'done' | 'skipped';

export const readTourOutcome = (storage: Env['storage'], key: string): TourOutcome | null => {
	try {
		const value = storage.getItem(key);
		return value === 'done' || value === 'skipped' ? value : null;
	} catch {
		// Storage blocked: count the tour as seen, so it does not open on every page.
		return 'skipped';
	}
};

export const writeTourOutcome = (storage: Env['storage'], key: string, outcome: TourOutcome): void => {
	try {
		storage.setItem(key, outcome);
	} catch {
		// Storage blocked: readTourOutcome already treats the tour as seen.
	}
};
