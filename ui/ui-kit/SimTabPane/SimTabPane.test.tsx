import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { SimTabPane } from './SimTabPane';

describe('SimTabPane', () => {
	it('names the pane by its id, as the element id and a class, and makes it a container around one columns container', () => {
		render(
			<SimTabPane id="gear-tab" className="p-0">
				<span data-testid="body" />
			</SimTabPane>,
		);

		const pane = screen.getByTestId('sim-tab');
		expect(pane.id).toBe('gear-tab');
		expect(pane.className.split(' ')).toEqual(['gear-tab', '@container', 'p-0']);
		const content = screen.getByTestId('tab-pane-content-container');
		expect(content.parentElement).toBe(pane);
		expect(content.classList.contains('ui-columns')).toBe(true);
		expect(screen.getByTestId('body').parentElement).toBe(content);
	});

	it('puts contentClassName on the columns container, not on the pane', () => {
		render(
			<SimTabPane id="settings-tab" contentClassName="max-xxl:flex-col">
				<span />
			</SimTabPane>,
		);

		expect(screen.getByTestId('tab-pane-content-container').classList.contains('max-xxl:flex-col')).toBe(true);
		expect(screen.getByTestId('sim-tab').classList.contains('max-xxl:flex-col')).toBe(false);
	});

	it('takes its testId and passes other attributes to the pane', () => {
		render(
			<SimTabPane id="rotation-tab" testId="rotation-pane" data-rotation-type="apl">
				<span />
			</SimTabPane>,
		);

		expect(screen.getByTestId('rotation-pane').getAttribute('data-rotation-type')).toBe('apl');
	});
});
