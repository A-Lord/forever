import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { Table } from './Table';
import { TableCell } from './TableCell';
import { TableHeaderCell } from './TableHeaderCell';
import { TableRow } from './TableRow';

const mount = (props: Parameters<typeof Table>[0] = {}) =>
	render(
		<Table {...props}>
			<thead>
				<TableRow testId="header-row">
					<TableHeaderCell testId="header-cell">Name</TableHeaderCell>
					<TableHeaderCell>DPS</TableHeaderCell>
				</TableRow>
			</thead>
			<tbody>
				<TableRow className="extra-row" testId="body-row">
					<TableCell testId="body-cell">Fireball</TableCell>
					<TableCell>1200</TableCell>
				</TableRow>
				<TableRow>
					<TableCell colSpan={2}>Total</TableCell>
				</TableRow>
			</tbody>
		</Table>,
	);

describe('Table', () => {
	it('renders a table of kit rows and cells, each with its class', () => {
		mount();

		const table = screen.getByTestId('table');
		expect(table.tagName).toBe('TABLE');
		expect(table.classList.contains('ui-table')).toBe(true);
		expect(screen.getByTestId('header-row').tagName).toBe('TR');
		expect(screen.getByTestId('header-row').classList.contains('ui-table-row')).toBe(true);
		expect(screen.getByTestId('header-cell').tagName).toBe('TH');
		expect(screen.getByTestId('header-cell').classList.contains('ui-table-header-cell')).toBe(true);
		expect(screen.getByTestId('body-cell').tagName).toBe('TD');
		expect(screen.getByTestId('body-cell').classList.contains('ui-table-cell')).toBe(true);
	});

	it('marks no variant by default', () => {
		mount();

		const table = screen.getByTestId('table');
		expect(table.hasAttribute('data-striped')).toBe(false);
		expect(table.hasAttribute('data-hover')).toBe(false);
		expect(table.hasAttribute('data-dense')).toBe(false);
		expect(table.hasAttribute('data-flush')).toBe(false);
	});

	it('marks each variant it is given as a data attribute', () => {
		mount({ striped: true, hover: true, dense: true, flush: true, testId: 'dps-table' });

		const table = screen.getByTestId('dps-table');
		expect(table.getAttribute('data-striped')).toBe('');
		expect(table.getAttribute('data-hover')).toBe('');
		expect(table.getAttribute('data-dense')).toBe('');
		expect(table.getAttribute('data-flush')).toBe('');
	});

	it('adds the caller class to its own and passes other attributes through', () => {
		mount({ className: 'extra-table', 'aria-label': 'Damage' });

		const table = screen.getByTestId('table');
		expect(table.className.split(' ')).toEqual(['ui-table', 'extra-table']);
		expect(table.getAttribute('aria-label')).toBe('Damage');
		expect(screen.getByTestId('body-row').className.split(' ')).toEqual(['ui-table-row', 'extra-row']);
		expect(screen.getByText('Total').getAttribute('colspan')).toBe('2');
	});

	it('marks the plain variant, and nothing for the default one', () => {
		const { unmount } = mount({ variant: 'plain' });
		expect(screen.getByTestId('table').getAttribute('data-variant')).toBe('plain');
		unmount();

		mount({ variant: 'default' });
		expect(screen.getByTestId('table').hasAttribute('data-variant')).toBe(false);
	});

	it('keeps a header cell on one line unless it is told to wrap', () => {
		render(
			<Table>
				<thead>
					<TableRow>
						<TableHeaderCell testId="fixed">Damage per second</TableHeaderCell>
						<TableHeaderCell testId="wrapping" wrap>
							Casts per minute
						</TableHeaderCell>
					</TableRow>
				</thead>
			</Table>,
		);

		expect(screen.getByTestId('fixed').hasAttribute('data-wrap')).toBe(false);
		expect(screen.getByTestId('wrapping').getAttribute('data-wrap')).toBe('');
	});
});
