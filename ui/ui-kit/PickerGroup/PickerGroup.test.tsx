import { render, screen } from '@testing-library/react';
import { createRef } from 'react';
import { describe, expect, it } from 'vitest';

import { PickerGroup } from './PickerGroup';

describe('PickerGroup', () => {
	it('renders a grid group, with the caller classes after its own', () => {
		render(<PickerGroup testId="group" className={['justify-end', { hidden: false }]} />);
		const group = screen.getByTestId('group');
		expect(group.getAttribute('class')).toBe('ui-picker-group justify-end');
	});

	it('adds the icon row class for the icons variant', () => {
		render(<PickerGroup testId="group" variant="icons" />);
		expect(screen.getByTestId('group').getAttribute('class')).toBe('ui-picker-group ui-picker-group-icons');
	});

	it('renders its children and forwards native attributes and the ref', () => {
		const ref = createRef<HTMLDivElement>();
		render(
			<PickerGroup ref={ref} testId="group" role="group" aria-label="Consumes">
				<span>Flask</span>
			</PickerGroup>,
		);
		const group = screen.getByRole('group', { name: 'Consumes' });
		expect(ref.current).toBe(group);
		expect(group.textContent).toBe('Flask');
	});

	it('lets an explicit data-testid stand in for testId', () => {
		render(<PickerGroup data-testid="consumes-food" />);
		expect(screen.getByTestId('consumes-food').classList.contains('ui-picker-group')).toBe(true);
	});
});
