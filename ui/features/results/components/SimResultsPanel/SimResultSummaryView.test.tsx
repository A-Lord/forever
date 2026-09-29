import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import type { ReferenceDiffs } from '../../model/reference_diffs';
import type { ResultMetric } from '../../model/topline_metrics';
import { SimResultSummaryView, type SimResultSummaryViewProps } from './SimResultSummaryView';

const METRICS: Array<ResultMetric> = [
	{ metric: 'dps', name: 'DPS', average: 1000, stdev: 10 },
	{ metric: 'tps', name: 'TPS', average: 800, stdev: 8 },
];

const DIFFS: ReferenceDiffs = {
	dps: { text: '+100.00 (10.00%)', tone: 'positive', significance: 'significant' },
	tps: { text: '+0.00 (0.00%)', tone: null, significance: 'not significant' },
};

const mount = (props: Partial<SimResultSummaryViewProps> = {}) => {
	const handlers = { onSetReference: vi.fn(), onSwapReference: vi.fn(), onClearReference: vi.fn() };
	const view = render(<SimResultSummaryView metrics={METRICS} pinClassName="text-class-warrior" {...handlers} {...props} />);
	return { ...handlers, ...view };
};

const strip = (container: HTMLElement) => container.querySelector('[data-testid="results-sim-reference"]')!;

describe('SimResultSummaryView', () => {
	it('lists every metric it is given and offers only the save button without a reference', () => {
		const { container } = mount();

		expect([...container.querySelectorAll('[data-testid="results-metric"]')].map(row => row.getAttribute('data-metric'))).toEqual(['dps', 'tps']);
		expect(strip(container).hasAttribute('data-has-reference')).toBe(false);
		expect(screen.getByTestId('results-sim-set-reference').hidden).toBe(false);
		expect(screen.getByTestId('results-sim-reference-swap').parentElement!.hidden).toBe(true);
		expect(container.querySelectorAll('[data-testid="results-reference-diff"]')).toHaveLength(0);
	});

	it('flags the strip, swaps the buttons and fills the deltas in once a reference is set', () => {
		const { container } = mount({ referenceDiffs: DIFFS });

		expect(strip(container).hasAttribute('data-has-reference')).toBe(true);
		expect(screen.getByTestId('results-sim-set-reference').hidden).toBe(true);
		expect(screen.getByTestId('results-sim-reference-swap').parentElement!.hidden).toBe(false);
		expect([...container.querySelectorAll('[data-testid="results-reference-diff"]')].map(diff => diff.textContent)).toEqual([
			'+100.00 (10.00%)',
			'+0.00 (0.00%)',
		]);
	});

	it('reports each of the three reference actions', () => {
		const { onSetReference, onSwapReference, onClearReference } = mount({ referenceDiffs: DIFFS });

		fireEvent.click(screen.getByTestId('results-sim-set-reference'));
		fireEvent.click(screen.getByTestId('results-sim-reference-swap'));
		fireEvent.click(screen.getByTestId('results-sim-reference-delete'));

		expect(onSetReference).toHaveBeenCalledOnce();
		expect(onSwapReference).toHaveBeenCalledOnce();
		expect(onClearReference).toHaveBeenCalledOnce();
	});

	it('draws the pin in the class accent and every glyph through Icon', () => {
		mount();

		const pin = screen.getByTestId('results-sim-set-reference').querySelector('svg')!;
		expect(pin.classList.contains('fa-map-pin')).toBe(true);
		expect(pin.classList.contains('text-class-warrior')).toBe(true);
		expect(screen.getByTestId('results-sim-reference-swap').querySelector('svg')!.classList.contains('fa-arrows-rotate')).toBe(true);
		expect(screen.getByTestId('results-sim-reference-delete').querySelector('svg')!.classList.contains('fa-xmark')).toBe(true);
		expect([...strip(document.body).querySelectorAll('svg')].every(icon => icon.getAttribute('aria-hidden') === 'true')).toBe(true);
	});
});
