import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { LandingHeader } from './LandingHeader';

describe('LandingHeader', () => {
	it('opens and closes the collapsed navigation from its togglers', () => {
		render(<LandingHeader />);
		const [open, close] = screen.getAllByTestId('navbar-toggler');
		const collapse = document.getElementById(open.getAttribute('aria-controls')!)!;
		expect(collapse.classList.contains('hidden')).toBe(true);

		fireEvent.click(open);
		expect(collapse.classList.contains('hidden')).toBe(false);
		expect(close.getAttribute('aria-expanded')).toBe('true');

		fireEvent.click(close);
		expect(collapse.classList.contains('hidden')).toBe(true);
	});
});
