import { fireEvent, render, screen } from '@testing-library/react';
import { Table } from '@ui-kit/Table';
import { describe, expect, it } from 'vitest';

import { CRIT_CAP_INFO } from '../../../../testing/storybook/fixtures/character-stats';
import { CritCapRow } from './CritCapRow';

const renderRow = (info = CRIT_CAP_INFO) =>
	render(
		<Table>
			<tbody>
				<CritCapRow info={info} text="8.40% under" />
			</tbody>
		</Table>,
	);

const openRows = async (container: HTMLElement) => {
	fireEvent.mouseEnter(container.querySelector('[data-testid="stat-value-link"]')!);
	await screen.findByText('sidebar.character_stats.attack_table.can_raise_by');
	return screen.getAllByTestId('character-stats-tooltip-row').map(row => [...row.children].map(cell => cell.textContent));
};

describe('CritCapRow', () => {
	it('shows the text coloured by the cap delta, and a breakdown ending in how far hit and expertise can raise the cap', async () => {
		const { container } = renderRow();
		const link = container.querySelector('[data-testid="stat-value-link"]')!;
		expect(link.textContent).toBe('8.40% under ');
		expect(link.className).toContain('text-success');

		const rows = await openRows(container);
		expect(rows.map(([label]) => label)).not.toContain('sidebar.character_stats.attack_table.spec_offsets');
		expect(rows.at(-1)).toEqual(['sidebar.character_stats.attack_table.can_raise_by', '3.70%']);
	});

	it('adds the spec offset row when there is one', async () => {
		const { container } = renderRow({ ...CRIT_CAP_INFO, specSpecificOffset: 1.5 });
		expect(await openRows(container)).toContainEqual(['sidebar.character_stats.attack_table.spec_offsets', '1.50%']);
	});
});
