import { render, screen } from '@testing-library/react';
import { createRef } from 'react';
import { describe, expect, it } from 'vitest';

import { FieldGrid } from './FieldGrid';

describe('FieldGrid', () => {
	// A layout primitive has no role or state to assert on; its class is the only observable
	// contract (STYLING §5's exception for layout primitives).
	it('renders the grid class, with the caller classes after its own', () => {
		render(<FieldGrid testId="group" className={['justify-end', { hidden: false }]} />);
		const group = screen.getByTestId('group');
		expect(group.getAttribute('class')).toBe('ui-field-grid justify-end');
		expect(group.getAttribute('role')).toBeNull();
	});

	it('renders its children and forwards native attributes and the ref', () => {
		const ref = createRef<HTMLDivElement>();
		render(
			<FieldGrid ref={ref} testId="group" role="group" aria-label="Consumes">
				<span>Flask</span>
			</FieldGrid>,
		);
		const group = screen.getByRole('group', { name: 'Consumes' });
		expect(ref.current).toBe(group);
		expect(group.textContent).toBe('Flask');
	});

	it('lets an explicit data-testid stand in for testId', () => {
		render(<FieldGrid data-testid="consumes-food" />);
		expect(screen.getByTestId('consumes-food').classList.contains('ui-field-grid')).toBe(true);
	});
});
