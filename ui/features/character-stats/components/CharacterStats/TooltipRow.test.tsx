import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { TooltipRow } from './TooltipRow';

describe('TooltipRow', () => {
	it('sets the label against the value, carrying the tone class', () => {
		const { getByTestId } = render(<TooltipRow label="Gear" value="+180" className="text-success" />);
		const row = getByTestId('character-stats-tooltip-row');
		expect([...row.children].map(cell => cell.textContent)).toEqual(['Gear', '+180']);
		expect(row.className).toContain('text-success');
	});
});
