import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { SummaryTableRow } from './SummaryTableRow';

describe('SummaryTableRow', () => {
	it('lays its children out as one flex line, in order', () => {
		render(
			<SummaryTableRow>
				<span>Runed Crimson Spinel</span>
				<span>4</span>
			</SummaryTableRow>,
		);

		const row = screen.getByTestId('summary-table-row');
		expect(row.tagName).toBe('DIV');
		expect(row.classList.contains('flex')).toBe(true);
		expect(row.classList.contains('items-center')).toBe(true);
		expect([...row.children].map(child => child.textContent)).toEqual(['Runed Crimson Spinel', '4']);
	});

	it('adds the caller class and takes its testId', () => {
		render(<SummaryTableRow className="py-1" testId="gem-row" />);

		expect(screen.getByTestId('gem-row').classList.contains('py-1')).toBe(true);
	});
});
