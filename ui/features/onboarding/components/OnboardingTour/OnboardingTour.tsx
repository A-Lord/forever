import { ResultsPanelStage } from '@features/results/components/SimResultsPanel/results_panel_store';
import { type ReactNode, useCallback, useEffect, useLayoutEffect, useRef, useState } from 'react';

import { useTrackedBox } from '../../hooks/useTrackedBox';
import { type CardPlacement, dimPath, placeCard, ringAround } from '../../model/placement';
import { cardKindAt, stepCount, type TourOutcome, type TourStep, type TourTarget } from '../../model/tour_steps';
import { TourCard, type TourCardProps } from '../TourCard';

/** How long a step waits for its part of the page (a tab's pane, a result) before it is passed over. */
const TARGET_WAIT_MS = 2500;
const TARGET_POLL_MS = 100;

export interface TourEnd<S> {
	outcome: TourOutcome;
	/** The card the tour ended on, and where it was: for analytics. */
	step: S;
	index: number;
	total: number;
}

export interface OnboardingTourProps<S extends TourStep<never>> {
	/** The steps this page has, in order (`applicableSteps`). */
	steps: ReadonlyArray<S>;
	/** A step's title and body. */
	text: (step: S) => { title: ReactNode; body: ReactNode };
	/** Opens what a step needs before its target is looked for, such as its tab. */
	prepare?: (step: S) => void;
	/** The sim run's stage, for a step that waits for a simulation. */
	stage?: ResultsPanelStage;
	onEnd: (end: TourEnd<S>) => void;
	/** Draws the card; `TourCard` by default. */
	renderCard?: (props: TourCardProps) => ReactNode;
	className?: string;
}

/** The element a target names, or null when the page does not have it right now. */
export const resolveTarget = (target: TourTarget, root: ParentNode = document): Element | null =>
	typeof target === 'string' ? root.querySelector(target) : target(root);

const viewport = () => ({ width: window.innerWidth, height: window.innerHeight });

/**
 * A guided tour over the page: the page dims except for a ring around the part a card explains, and
 * the card sits beside it. The page cannot be clicked while a card explains it, except on a step that
 * waits for the player to run a simulation. A step whose part the page does not show is passed over in
 * the direction of travel. Right and left arrows move on and back; Escape leaves.
 */
export const OnboardingTour = <S extends TourStep<never>>({ steps, text, prepare, stage, onEnd, renderCard, className }: OnboardingTourProps<S>) => {
	// By id: the steps can change under the tour (a result appears once the player has run the sim).
	const [stepId, setStepId] = useState(steps[0].id);
	const direction = useRef<1 | -1>(1);
	const index = Math.max(
		0,
		steps.findIndex(entry => entry.id === stepId),
	);
	const step = steps[index];
	const [target, setTarget] = useState<Element | null>(null);
	const [searching, setSearching] = useState(false);
	const box = useTrackedBox(target);
	const ring = box && ringAround(box, viewport());

	const latestPrepare = useRef(prepare);
	latestPrepare.current = prepare;

	const finish = useCallback((outcome: TourOutcome) => onEnd({ outcome, step, index, total: steps.length }), [onEnd, step, index, steps.length]);

	const go = useCallback(
		(delta: 1 | -1) => {
			direction.current = delta;
			const next = steps[index + delta];
			if (next) setStepId(next.id);
			else if (delta === 1) finish('done');
		},
		[steps, index, finish],
	);

	// Opens what the step needs and finds the part of the page it rings.
	useEffect(() => {
		setTarget(null);
		latestPrepare.current?.(step);
		if (!step.target) {
			setSearching(false);
			return;
		}
		setSearching(true);
		let tries = 0;
		let timer = 0;
		const look = () => {
			const found = resolveTarget(step.target!);
			if (found) {
				setSearching(false);
				// At once and centred: a smooth scroll is cut short when a tab's pane re-renders.
				found.scrollIntoView({ block: 'center', inline: 'nearest' });
				setTarget(found);
				return;
			}
			if (++tries * TARGET_POLL_MS > TARGET_WAIT_MS) {
				setSearching(false);
				go(direction.current);
				return;
			}
			timer = window.setTimeout(look, TARGET_POLL_MS);
		};
		// After the tab's pane has rendered.
		timer = window.setTimeout(look, TARGET_POLL_MS / 2);
		return () => clearTimeout(timer);
		// Once per step: `go` changes as the steps do, and must not restart the search.
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [step.id]);

	// A step that waits for a simulation moves on once a run started on it has finished.
	const sawRun = useRef(false);
	useEffect(() => {
		sawRun.current = false;
	}, [step.id]);
	useEffect(() => {
		if (step.waitFor !== 'simulation' || stage === undefined) return;
		if (stage === ResultsPanelStage.Pending || stage === ResultsPanelStage.Running) sawRun.current = true;
		else if (stage === ResultsPanelStage.Result && sawRun.current) go(1);
	}, [stage, step.waitFor, go]);

	const waiting = step.waitFor === 'simulation';
	const kind = cardKindAt(steps, index);
	const { position, total } = stepCount(steps, index);

	// The card's place from its own size, hidden until measured so it does not jump.
	const card = useRef<HTMLDivElement>(null);
	const [placement, setPlacement] = useState<CardPlacement | null>(null);
	useLayoutEffect(() => {
		const element = card.current;
		if (!element || (step.target && !ring)) return setPlacement(null);
		setPlacement(placeCard(ring, { width: element.offsetWidth, height: element.offsetHeight }, viewport()));
		// `ring` is rebuilt on every render; its numbers are what matter.
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [step.id, ring?.left, ring?.top, ring?.width, ring?.height]);

	// The primary button takes the focus on every card, so Enter and Space move on.
	useEffect(() => {
		if (!searching) card.current?.querySelector<HTMLElement>('[data-tour-primary]')?.focus({ preventScroll: true });
	}, [step.id, searching]);

	useEffect(() => {
		const onKey = (event: KeyboardEvent) => {
			if (event.target instanceof HTMLElement && event.target.closest('input, textarea, select, [contenteditable="true"]')) return;
			if (event.key === 'Escape') finish(kind === 'finish' ? 'done' : 'skipped');
			else if (event.key === 'ArrowRight' && !waiting) go(1);
			else if (event.key === 'ArrowLeft' && index > 0) go(-1);
			else return;
			event.preventDefault();
		};
		window.addEventListener('keydown', onKey);
		return () => window.removeEventListener('keydown', onKey);
	}, [finish, go, index, waiting, kind]);

	const { title, body } = text(step);
	const cardProps: TourCardProps = {
		kind,
		title,
		body,
		position,
		total,
		waiting,
		side: placement?.side,
		style: placement ? { transform: `translate(${Math.round(placement.left)}px, ${Math.round(placement.top)}px)` } : { visibility: 'hidden' },
		cardRef: card,
		onNext: () => go(1),
		onPrevious: index > 0 ? () => go(-1) : undefined,
		onSkip: () => finish(kind === 'finish' ? 'done' : 'skipped'),
	};

	return (
		<div className={className ?? 'ui-tour'} data-waiting={waiting ? '' : undefined} data-testid="tour" data-step={step.id}>
			{/* Blocks the page while a card explains it; lets it through while the card waits for the player. */}
			<div className="ui-tour-shield" aria-hidden="true" data-testid="tour-shield" />
			{/* The dim with a hole where the ring is: one path, cheap to paint where a page-sized shadow is not. */}
			<svg className="ui-tour-dim" aria-hidden="true" width="100%" height="100%">
				<path fillRule="evenodd" d={dimPath(ring, viewport())} />
			</svg>
			{ring && (
				<div
					className="ui-tour-ring"
					aria-hidden="true"
					style={{ transform: `translate(${ring.left}px, ${ring.top}px)`, width: ring.width, height: ring.height }}
					data-testid="tour-ring"
				/>
			)}
			{renderCard ? renderCard(cardProps) : <TourCard key={step.id} {...cardProps} />}
		</div>
	);
};
