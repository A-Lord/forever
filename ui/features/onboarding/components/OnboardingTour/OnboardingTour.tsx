import { type KeyboardEvent, type ReactNode, useEffect, useEffectEvent, useState } from 'react';

import { useTrackedBox } from '../../hooks/useTrackedBox';
import { dimPath, ringAround } from '../../model/placement';
import { cardKindAt, type SimulationState, stepCount, type TourOutcome, type TourStep } from '../../model/tour_steps';
import { TourCard } from '../TourCard';

/** How long a step looks for its part of the page (a tab's pane, a result) before it is passed over. */
const TARGET_WAIT_MS = 1500;
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
	/** For a step that waits for a simulation: it moves on once one it saw running is done. */
	simulation?: SimulationState;
	onEnd: (end: TourEnd<S>) => void;
}

// A hidden element still matches its selector, with an empty box.
const findTarget = (selector: string): Element | null => {
	const element = document.querySelector(selector);
	const rect = element?.getBoundingClientRect();
	return rect && rect.width > 0 && rect.height > 0 ? element : null;
};

const viewport = () => ({ width: window.innerWidth, height: window.innerHeight });

/**
 * A guided tour over the page: the page dims except for a ring around the part a card explains, and
 * the card sits beside it. The page cannot be clicked while a card explains it, except on a step that
 * waits for the player to run a simulation. A step whose part the page does not show is passed over in
 * the direction of travel. In the card, the right and left arrows move on and back; Escape leaves.
 */
export const OnboardingTour = <S extends TourStep<never>>({ steps, text, prepare, simulation, onEnd }: OnboardingTourProps<S>) => {
	// By id: the steps change under the tour (a result appears once the player has run the sim). When the
	// current one drops out, the tour stays where it was in the list.
	const [current, setCurrent] = useState({ id: steps[0].id, index: 0, direction: 1 as 1 | -1 });
	const found = steps.findIndex(entry => entry.id === current.id);
	const index = found >= 0 ? found : Math.min(current.index, steps.length - 1);
	const step = steps[index];
	// The element a step rings, kept with the step it was found for.
	const [located, setLocated] = useState<{ id: string; element: Element } | null>(null);
	const target = located?.id === step.id ? located.element : null;
	const box = useTrackedBox(target);
	const ring = box && ringAround(box, viewport());
	const kind = cardKindAt(steps, index);
	const waiting = step.waitFor === 'simulation';

	// The step a simulation was seen running on: a run counts only once the waiting step has seen it start.
	const [sawRun, setSawRun] = useState<string | null>(null);

	const finish = (outcome: TourOutcome) => onEnd({ outcome, step, index, total: steps.length });
	const leave = () => finish(kind === 'finish' ? 'done' : 'skipped');
	const go = (delta: 1 | -1) => {
		const next = steps[index + delta];
		if (!next) {
			if (delta === 1) finish('done');
			return;
		}
		setCurrent({ id: next.id, index: index + delta, direction: delta });
		setSawRun(null);
	};

	const open = useEffectEvent(() => prepare?.(step));
	const passOver = useEffectEvent(() => go(current.direction));
	const advance = useEffectEvent(() => go(1));

	// Opens what the step needs and finds the part of the page it rings.
	const { id: stepId, target: selector } = step;
	useEffect(() => {
		open();
		if (!selector) return;
		let tries = 0;
		let timer = 0;
		const look = () => {
			const element = findTarget(selector);
			if (element) {
				// At once and centred: a smooth scroll is cut short when a tab's pane re-renders.
				element.scrollIntoView({ block: 'center', inline: 'nearest' });
				setLocated({ id: stepId, element });
			} else if (++tries * TARGET_POLL_MS > TARGET_WAIT_MS) {
				passOver();
			} else {
				timer = window.setTimeout(look, TARGET_POLL_MS);
			}
		};
		// After the tab's pane has rendered.
		timer = window.setTimeout(look, TARGET_POLL_MS / 2);
		return () => clearTimeout(timer);
	}, [stepId, selector]);

	useEffect(() => {
		if (!waiting || simulation === undefined) return;
		if (simulation === 'running') setSawRun(stepId);
		else if (simulation === 'done' && sawRun === stepId) advance();
	}, [waiting, simulation, sawRun, stepId]);

	const onKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
		if (event.key === 'ArrowRight' && !waiting) go(1);
		else if (event.key === 'ArrowLeft' && index > 0) go(-1);
		else return;
		event.preventDefault();
	};

	const { title, body } = text(step);
	const { position, total } = stepCount(steps, index);

	return (
		<>
			<div className="ui-tour" data-waiting={waiting ? '' : undefined} data-testid="tour" data-step={step.id}>
				<div className="ui-tour-shield" aria-hidden="true" data-testid="tour-shield" />
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
			</div>
			{(!step.target || target) && (
				<TourCard
					// A card per step, so each one takes the focus as it opens.
					key={step.id}
					kind={kind}
					title={title}
					body={body}
					position={position}
					total={total}
					anchor={target}
					waiting={waiting}
					onNext={() => go(1)}
					onPrevious={index > 0 ? () => go(-1) : undefined}
					onSkip={leave}
					onKeyDown={onKeyDown}
				/>
			)}
		</>
	);
};
