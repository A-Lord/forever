import { ResultsPanelStage } from '@features/results/components/SimResultsPanel/results_panel_store';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import type { TourStep } from '../../model/tour_steps';
import { OnboardingTour, type TourEnd } from './OnboardingTour';

const page = () => {
	document.body.insertAdjacentHTML('beforeend', '<div id="page"><button id="gear">Gear</button><button id="simulate">Simulate</button></div>');
};

const STEPS: ReadonlyArray<TourStep> = [
	{ id: 'welcome' },
	{ id: 'gear', target: '#gear', tab: 'gear-tab' },
	{ id: 'missing', target: '#not-on-this-page' },
	{ id: 'simulate', target: '#simulate', waitFor: 'simulation' },
	{ id: 'finish' },
];

const text = (step: TourStep) => ({ title: `title ${step.id}`, body: `body ${step.id}` });

const tour = (props: { stage?: ResultsPanelStage; prepare?: (step: TourStep) => void } = {}) => {
	const onEnd = vi.fn<(end: TourEnd<TourStep>) => void>();
	const view = render(<OnboardingTour steps={STEPS} text={text} onEnd={onEnd} {...props} />);
	return { onEnd, view };
};

const stepId = () => screen.getByTestId('tour').getAttribute('data-step');
const settle = (ms = 200) => act(() => vi.advanceTimersByTime(ms));

describe('OnboardingTour', () => {
	beforeEach(() => {
		vi.useFakeTimers();
		page();
	});
	afterEach(() => {
		vi.useRealTimers();
		document.getElementById('page')?.remove();
	});

	it('rings the part of the page each step explains, opening what it needs first', () => {
		const prepare = vi.fn();
		tour({ prepare });
		expect(stepId()).toBe('welcome');
		expect(screen.queryByTestId('tour-ring')).toBeNull();
		fireEvent.click(screen.getByTestId('tour-card-next'));
		settle();
		expect(stepId()).toBe('gear');
		expect(prepare).toHaveBeenLastCalledWith(STEPS[1]);
		expect(screen.getByTestId('tour-ring')).toBeTruthy();
		expect(screen.getByTestId('tour-card-title').textContent).toBe('title gear');
	});

	it('passes over a step whose part the page does not show, in the direction of travel', () => {
		tour();
		fireEvent.keyDown(window, { key: 'ArrowRight' });
		settle();
		fireEvent.keyDown(window, { key: 'ArrowRight' });
		settle(3000);
		expect(stepId()).toBe('simulate');
		fireEvent.keyDown(window, { key: 'ArrowLeft' });
		settle(3000);
		expect(stepId()).toBe('gear');
	});

	it('waits on the Simulate step until a run started there finishes, letting the page through meanwhile', () => {
		const { view, onEnd } = tour({ stage: ResultsPanelStage.Idle });
		const rerender = (stage: ResultsPanelStage) => view.rerender(<OnboardingTour steps={STEPS} text={text} onEnd={onEnd} stage={stage} />);
		for (let i = 0; i < 2; i++) {
			fireEvent.keyDown(window, { key: 'ArrowRight' });
			settle(3000);
		}
		expect(stepId()).toBe('simulate');
		expect(screen.getByTestId('tour').hasAttribute('data-waiting')).toBe(true);
		// The arrow key does not pass a waiting step; a result that was there before the run does not either.
		fireEvent.keyDown(window, { key: 'ArrowRight' });
		rerender(ResultsPanelStage.Result);
		expect(stepId()).toBe('simulate');
		rerender(ResultsPanelStage.Running);
		rerender(ResultsPanelStage.Result);
		expect(stepId()).toBe('finish');
	});

	it('reports how and where it ended', () => {
		const { onEnd } = tour();
		fireEvent.keyDown(window, { key: 'ArrowRight' });
		settle();
		fireEvent.keyDown(window, { key: 'Escape' });
		expect(onEnd).toHaveBeenCalledWith({ outcome: 'skipped', step: STEPS[1], index: 1, total: STEPS.length });
	});

	it('finishes as done from the last card', () => {
		const { onEnd } = tour({ stage: ResultsPanelStage.Idle });
		for (let i = 0; i < 2; i++) {
			fireEvent.keyDown(window, { key: 'ArrowRight' });
			settle(3000);
		}
		fireEvent.click(screen.getByTestId('tour-card-next'));
		settle();
		expect(stepId()).toBe('finish');
		fireEvent.click(screen.getByTestId('tour-card-next'));
		expect(onEnd).toHaveBeenCalledWith(expect.objectContaining({ outcome: 'done' }));
	});
});
