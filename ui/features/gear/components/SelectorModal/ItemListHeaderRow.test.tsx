import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { ItemListSortBy } from '../../model/item_sort';
import { SelectorModalTabs } from '../../types';
import { ItemListHeaderRow, type ItemListHeaderRowProps } from './ItemListHeaderRow';

const renderRow = (props: Partial<ItemListHeaderRowProps> = {}) => {
	const onSort = vi.fn();
	const { container } = render(<ItemListHeaderRow label={SelectorModalTabs.Items} showEPValues epTooltipId="ep-tooltip" onSort={onSort} {...props} />);
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
		fireEvent.click(screen.getByTestId('ilvl-label'));
		fireEvent.click(screen.getByTestId('ep-label'));
		expect(onSort.mock.calls).toEqual([[ItemListSortBy.ILVL], [ItemListSortBy.EP]]);
	});

	it('hides the EP header when EP values are off', () => {
		renderRow({ showEPValues: false });
		expect(screen.getByTestId('ep-label').style.display).toBe('none');
	});

	it('anchors the EP help at the tooltip it is given', () => {
		renderRow({ epTooltipId: 'the-ep-tooltip' });
		expect(screen.getByRole('button').getAttribute('data-tooltip-id')).toBe('the-ep-tooltip');
	});
});
