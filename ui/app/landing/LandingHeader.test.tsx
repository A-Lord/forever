import i18n from '@i18n/config';
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

	it('opens every external nav link in a new tab with no opener', () => {
		render(<LandingHeader />);
		const names = [i18n.t('info.discord'), i18n.t('info.github'), i18n.t('landing.header.supportDevs')];

		names.forEach(name => {
			const link = screen.getByRole('link', { name });
			expect(link.getAttribute('target')).toBe('_blank');
			expect(link.getAttribute('rel')).toBe('noopener noreferrer');
		});
	});

	it('does not turn the logo into a link', () => {
		render(<LandingHeader />);
		expect(screen.getByTestId('wowsims-title').closest('a')).toBeNull();
	});
});
