import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { CombinationsCountView, type CombinationsCountViewProps } from './CombinationsCountView';

const renderCount = (props: Partial<CombinationsCountViewProps> = {}) =>
	render(<CombinationsCountView pending={false} combinations={12} iterations={24000} iterationsLimit={50000} {...props} />);

describe('CombinationsCountView', () => {
	it('shows a spinner alone while the count is pending', () => {
		renderCount({ pending: true });
		const count = screen.getByTestId('bulk-combinations-count');

		expect(count.querySelector('[data-testid="loader"]')).not.toBeNull();
		expect(count.textContent).toBe('');
	});

	it('counts the combinations and the iterations they cost', () => {
		renderCount();
		const count = screen.getByTestId('bulk-combinations-count');

		expect(count.textContent).toContain('bulk_tab.settings.combinations_count');
		expect(count.textContent).toContain('24,000 bulk_tab.settings.iterations');
		expect(screen.queryByTestId('warning')).toBeNull();
	});

	it('names a single combination in the singular', () => {
		renderCount({ combinations: 1 });

		expect(screen.getByTestId('bulk-combinations-count').textContent).toContain('bulk_tab.settings.combination_singular');
	});

	it('warns once the iterations pass the limit', () => {
		renderCount({ iterations: 60000 });
		const warning = screen.getByTestId('warning');

		expect(warning.tagName).toBe('BUTTON');
		expect(warning.getAttribute('aria-label')).toBe('bulk_tab.warning.iterations_limit');
		expect(warning.hasAttribute('data-tooltip-id')).toBe(true);
	});
});
