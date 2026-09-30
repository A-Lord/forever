import i18n from '@i18n/config';
import { SortDirection } from '@sim/constants/other';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { ItemListSortBy } from '../../model/item_sort';
import { SelectorModalTabs } from '../../types';
import { ItemListHeaderRow, type ItemListHeaderRowProps } from './ItemListHeaderRow';

const renderRow = (props: Partial<ItemListHeaderRowProps> = {}) => {
	const onSort = vi.fn();
	const { container } = render(
		<ItemListHeaderRow
			label={SelectorModalTabs.Items}
			showEPValues
			epTooltipId="ep-tooltip"
			sortBy={ItemListSortBy.EP}
			sortDirection={SortDirection.DESC}
			onSort={onSort}
			{...props}
		/>,
	);
	const headers = Array.from(container.querySelectorAll('h6')).map(node => node.getAttribute('data-testid'));
	return { onSort, headers };
};

describe('ItemListHeaderRow', () => {
	it('gives the ilvl, source and compare columns to the items tab alone', () => {
		expect(renderRow().headers).toEqual(['ilvl-label', 'item-label', 'source-label', 'ep-label', 'favorite-label', 'compare-label']);
	});

	it('keeps the item, EP and favourite columns on another tab', () => {
		expect(renderRow({ label: SelectorModalTabs.Enchants }).headers).toEqual(['item-label', 'ep-label', 'favorite-label']);
	});

	it('sorts by item level and by EP from their headers', () => {
		const { onSort } = renderRow();
		fireEvent.click(within(screen.getByTestId('ilvl-label')).getByRole('button'));
		fireEvent.click(within(screen.getByTestId('ep-label')).getByRole('button', { name: /^EP/ }));
		expect(onSort.mock.calls).toEqual([[ItemListSortBy.ILVL], [ItemListSortBy.EP]]);
	});

	it.each([
		[ItemListSortBy.ILVL, SortDirection.ASC, 'ilvl-label', 'ep-label', 'gear_tab.gear_picker.sorted_ascending'],
		[ItemListSortBy.EP, SortDirection.DESC, 'ep-label', 'ilvl-label', 'gear_tab.gear_picker.sorted_descending'],
	] as const)('names the sort direction on the active column’s button only (sort %s, direction %s)', (sortBy, sortDirection, active, other, directionKey) => {
		renderRow({ sortBy, sortDirection });
		const headerName = (id: 'ilvl-label' | 'ep-label') => (id === 'ep-label' ? 'EP' : i18n.t('gear_tab.gear_picker.table_headers.ilvl'));
		expect(
			within(screen.getByTestId(active)).getByRole('button', { name: new RegExp(`^${headerName(active)}\\s*,\\s*${i18n.t(directionKey)}$`) }),
		).toBeTruthy();
		expect(within(screen.getByTestId(other)).getByRole('button', { name: headerName(other) })).toBeTruthy();
	});

	it('hides the EP header when EP values are off', () => {
		renderRow({ showEPValues: false });
		expect(screen.getByTestId('ep-label').style.display).toBe('none');
	});

	it('anchors the EP help at the tooltip it is given', () => {
		renderRow({ epTooltipId: 'the-ep-tooltip' });
		expect(screen.getByRole('button', { name: i18n.t('gear_tab.gear_picker.ep_tooltip') }).getAttribute('data-tooltip-id')).toBe('the-ep-tooltip');
	});
});
