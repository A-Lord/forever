import { act, fireEvent, render, screen } from '@testing-library/react';
import { NOTICE_COLOR } from '@ui-kit/Notice';
import { describe, expect, it } from 'vitest';

import { NoticeIcon } from './NoticeIcon';

describe('NoticeIcon', () => {
	it.each([
		['info', 'fa-circle-info'],
		['warning', 'fa-triangle-exclamation'],
		['error', 'fa-triangle-exclamation'],
	] as const)('draws the %s icon in its colour, with the level on the button', (level, glyph) => {
		render(
			<NoticeIcon level={level} label="Notices">
				Some notice
			</NoticeIcon>,
		);
		const button = screen.getByRole('button', { name: 'Notices' });
		expect(button.getAttribute('data-level')).toBe(level);
		const icon = button.querySelector('svg')!;
		expect(icon.classList.contains(glyph)).toBe(true);
		for (const token of NOTICE_COLOR[level].split(' ')) expect(icon.classList.contains(token)).toBe(true);
	});

	it('opens the tooltip on click and shows the children', async () => {
		render(
			<NoticeIcon level="warning" label="Notices">
				<p>First notice</p>
			</NoticeIcon>,
		);
		expect(screen.queryByText('First notice')).toBeNull();

		const button = screen.getByRole('button', { name: 'Notices' });
		// A real press focuses the button it lands on; `fireEvent.click` does not.
		await act(async () => {
			button.focus();
			fireEvent.click(button);
		});
		expect(screen.getByText('First notice')).toBeTruthy();
	});

	it('stays mounted but hidden when hidden is set', () => {
		render(
			<NoticeIcon level="warning" label="Notices" hidden>
				Some notice
			</NoticeIcon>,
		);
		const button = document.querySelector('button[aria-label="Notices"]') as HTMLElement;
		expect(button).toBeTruthy();
		expect(button.style.display).toBe('none');
	});
});
