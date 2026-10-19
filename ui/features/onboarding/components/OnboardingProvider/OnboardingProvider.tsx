import type { Env } from '@sim/state/env';
import { createContext, type ReactNode, use, useCallback, useEffect, useMemo, useState } from 'react';

import { readTourOutcome, type TourOutcome, writeTourOutcome } from '../../model/tour_steps';

/** How long the page settles, once ready, before the tour opens by itself. */
export const FIRST_OPEN_DELAY_MS = 900;

/** How the tour opened: by itself on a first visit, or because the player asked for it. */
export type TourTrigger = 'auto' | 'restart';

export interface TourRun {
	trigger: TourTrigger;
	/** Ends the tour and records how, so it does not open by itself again. */
	end: (outcome: TourOutcome) => void;
}

const OnboardingContext = createContext<{ restart: () => void } | null>(null);

/** Opens the tour again from its first card; undefined outside an `OnboardingProvider`. */
export const useOnboardingRestart = (): (() => void) | undefined => use(OnboardingContext)?.restart;

export interface OnboardingProviderProps {
	/** The page is ready to be toured; the tour waits for it. */
	ready: boolean;
	storage: Env['storage'];
	/** Where the outcome is kept: once a tour has ended it does not open by itself again. */
	storageKey: string;
	/** Draws the running tour. A new run (a restart) remounts it at its first card. */
	renderTour: (run: TourRun) => ReactNode;
	children: ReactNode;
}

/** Runs a tour once by itself, when the page is first ready and none has ended before, and again whenever `useOnboardingRestart` asks. */
export const OnboardingProvider = ({ ready, storage, storageKey, renderTour, children }: OnboardingProviderProps) => {
	// The run's number: 0 for the one that opened by itself, each restart one more; null while closed.
	const [run, setRun] = useState<number | null>(null);

	useEffect(() => {
		if (!ready || readTourOutcome(storage, storageKey)) return;
		const timer = setTimeout(() => setRun(current => current ?? 0), FIRST_OPEN_DELAY_MS);
		return () => clearTimeout(timer);
	}, [ready, storage, storageKey]);

	const restart = useCallback(() => setRun(current => (current ?? 0) + 1), []);
	const end = useCallback(
		(outcome: TourOutcome) => {
			writeTourOutcome(storage, storageKey, outcome);
			setRun(null);
		},
		[storage, storageKey],
	);
	const value = useMemo(() => ({ restart }), [restart]);

	return (
		<OnboardingContext value={value}>
			{children}
			{run !== null && ready && <TourMount key={run} trigger={run === 0 ? 'auto' : 'restart'} end={end} renderTour={renderTour} />}
		</OnboardingContext>
	);
};

const TourMount = ({ trigger, end, renderTour }: TourRun & { renderTour: (run: TourRun) => ReactNode }) => renderTour({ trigger, end });
