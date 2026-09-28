import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { TabBadge } from './TabBadge';

describe('TabBadge', () => {
	it('renders nothing without a label', () => {
		const { container } = render(<TabBadge />);

		expect(container.childNodes).toHaveLength(0);
	});

	it('renders nothing for an empty label', () => {
		const { container } = render(<TabBadge label="" />);

		expect(container.childNodes).toHaveLength(0);
	});

	it('puts the label in parentheses as text, with the label in its own span', () => {
		const { container } = render(<TabBadge label="Beta" />);

		expect(container.textContent).toBe(' (Beta)');
		expect(screen.getByTestId('tab-badge').textContent).toBe('Beta');
		expect(screen.getByTestId('tab-badge').className).toBe('text-success');
	});

	it('takes its testId', () => {
		render(<TabBadge label="New" testId="bulk-badge" />);

		expect(screen.getByTestId('bulk-badge').textContent).toBe('New');
	});
});
