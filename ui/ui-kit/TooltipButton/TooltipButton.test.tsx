import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { TooltipButton } from './TooltipButton';

describe('TooltipButton', () => {
	it('renders a chromeless link-coloured icon button, with the extra classes the callers pass', () => {
		render(<TooltipButton tooltip="What this does" className="ml-2" />);
		const button = screen.getByRole('button');
		expect(Array.from(button.classList).filter(name => name.startsWith('ui-'))).toEqual(['ui-button-icon', 'ui-button-link']);
		expect(button.classList.contains('ml-2')).toBe(true);
		for (const token of ['flex', 'items-center', 'justify-center']) expect(button.classList.contains(token)).toBe(true);
		expect(button.getAttribute('data-testid')).toBe('tooltip-button');
		expect(button.getAttribute('type')).toBe('button');
		// `Icon` normalises the FA5 spelling `question-circle` into its FA6 name.
		const icon = button.querySelector('svg')!;
		expect(icon.classList).toContain('fa-circle-question');
		expect(icon.getAttribute('data-prefix')).toBe('far');
	});

	it('carries the testId it is given in place of the tooltip-button default', () => {
		render(<TooltipButton tooltip="What this does" testId="consumes-help" />);
		expect(screen.getByRole('button').getAttribute('data-testid')).toBe('consumes-help');
	});

	it('takes another glyph, which the vanilla component hardcodes', () => {
		render(<TooltipButton tooltip="Warning" icon="triangle-exclamation" iconStyle="solid" />);
		const icon = screen.getByRole('button').querySelector('svg')!;
		expect(icon.classList).toContain('fa-triangle-exclamation');
		expect(icon.getAttribute('data-prefix')).toBe('fas');
	});

	it('shows its tooltip on hover and nothing before', () => {
		render(<TooltipButton tooltip="What this does" />);
		expect(screen.queryByText('What this does')).toBeNull();

		fireEvent.mouseEnter(screen.getByRole('button'));
		expect(screen.getByText('What this does')).toBeTruthy();
	});

	it('gives each button its own tooltip', () => {
		render(
			<>
				<TooltipButton tooltip="First" />
				<TooltipButton tooltip="Second" />
			</>,
		);
		const [first, second] = screen.getAllByRole('button');
		expect(first.dataset.tooltipId).not.toBe(second.dataset.tooltipId);

		fireEvent.mouseEnter(second);
		expect(screen.getByText('Second')).toBeTruthy();
		expect(screen.queryByText('First')).toBeNull();
	});
});
