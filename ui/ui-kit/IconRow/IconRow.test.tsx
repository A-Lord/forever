import { render, screen } from '@testing-library/react';
import { createRef } from 'react';
import { describe, expect, it } from 'vitest';

import { IconRow } from './IconRow';

describe('IconRow', () => {
	// A layout primitive has no role or state to assert on; its class is the only observable
	// contract (STYLING §5's exception, as Task 0 made for `iconGroupClassName`).
	it('renders the icon row class, with the caller classes after its own', () => {
		render(<IconRow testId="group" className={['justify-end', { hidden: false }]} />);
		const group = screen.getByTestId('group');
		expect(group.getAttribute('class')).toBe('ui-icon-row justify-end');
		expect(group.getAttribute('role')).toBeNull();
	});

	it('renders its children and forwards native attributes and the ref', () => {
		const ref = createRef<HTMLDivElement>();
		render(
			<IconRow ref={ref} testId="group" role="group" aria-label="Consumes">
				<span>Flask</span>
			</IconRow>,
		);
		const group = screen.getByRole('group', { name: 'Consumes' });
		expect(ref.current).toBe(group);
		expect(group.textContent).toBe('Flask');
	});

	it('lets an explicit data-testid stand in for testId', () => {
		render(<IconRow data-testid="consumes-food" />);
		expect(screen.getByTestId('consumes-food').classList.contains('ui-icon-row')).toBe(true);
	});
});
