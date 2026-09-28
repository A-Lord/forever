import { render, screen } from '@testing-library/react';
import { Table } from '@ui-kit/Table';
import { describe, expect, it } from 'vitest';

import { ReforgeStatCapRowView, type ReforgeStatCapRowViewProps } from './ReforgeStatCapRowView';

const renderRow = (props: Partial<ReforgeStatCapRowViewProps> = {}) =>
	render(
		<Table>
			<tbody>
				<ReforgeStatCapRowView
					statName="Haste"
					percentage={<span data-testid="percentage" />}
					undershoot={<span data-testid="undershoot" />}
					{...props}
				/>
			</tbody>
		</Table>,
	);

describe('ReforgeStatCapRowView', () => {
	it('puts the stat name, the percentage and the undershoot flag in one row', () => {
		renderRow();
		const [row] = screen.getAllByRole('row');

		expect(Array.from(row.children).map(cell => cell.textContent || cell.firstElementChild?.getAttribute('data-testid'))).toEqual([
			'Haste ',
			'percentage',
			'undershoot',
		]);
	});

	it('explains the stat through a tooltip anchor when it has one', () => {
		renderRow({ tooltip: 'Final percentage value including all buffs/gear.' });

		expect(screen.getByRole('button').hasAttribute('data-tooltip-id')).toBe(true);
	});

	it('leaves the tooltip anchor out when the stat has no tooltip', () => {
		renderRow();

		expect(screen.queryByRole('button')).toBeNull();
	});

	it('adds a second row for the spec’s preset caps', () => {
		renderRow({ presets: <span data-testid="presets" /> });
		const rows = screen.getAllByRole('row');

		expect(rows).toHaveLength(2);
		expect(rows[1].contains(screen.getByTestId('presets'))).toBe(true);
	});
});
