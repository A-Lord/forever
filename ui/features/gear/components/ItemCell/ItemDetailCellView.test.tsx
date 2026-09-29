import { ItemSlot } from '@generated/proto/common';
import { translateSlotName } from '@i18n/localization';
import { makeEquippedItem, SOCKETED_ITEM } from '@testing/storybook/fixtures/gear';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { getEmptySlotIconUrl } from '../../model/empty_slot_icons';
import { SelectorModalTabs } from '../../types';
import { ItemDetailCellView, type ItemDetailCellViewProps } from './ItemDetailCellView';

const renderCell = (props: Partial<ItemDetailCellViewProps> = {}) =>
	render(<ItemDetailCellView slot={ItemSlot.ItemSlotMainHand} item={makeEquippedItem()} iconUrl="sword.jpg" href="https://example.com/item" {...props} />);

const icon = () => screen.getByTestId('item-picker-icon');
const name = () => screen.getByTestId('item-picker-name-container');

describe('ItemDetailCellView', () => {
	it('shows the item level, name, icon and link of the equipped item', () => {
		renderCell();
		expect(screen.getByTestId('item-picker-ilvl').textContent).toBe('80');
		expect(screen.getByTestId('item-picker-name').textContent).toBe('Thunderfury, Blessed Blade of the Windseeker');
		expect(icon().style.backgroundImage).toContain('sword.jpg');
		expect([icon(), name()].map(anchor => anchor.getAttribute('href'))).toEqual(['https://example.com/item', 'https://example.com/item']);
	});

	it('shows the slot name and its empty icon when nothing is equipped', () => {
		renderCell({ item: null, iconUrl: '' });
		expect(name().textContent).toBe(translateSlotName(ItemSlot.ItemSlotMainHand));
		expect(icon().style.backgroundImage).toContain(getEmptySlotIconUrl(ItemSlot.ItemSlotMainHand));
		expect(screen.queryByTestId('gem-socket-container')).toBeNull();
	});

	it('puts the wowhead tooltip on both the icon and the name', () => {
		renderCell({ wowhead: { 'data-wowhead': 'item=19019' } });
		expect([icon(), name()].map(anchor => anchor.dataset.wowhead)).toEqual(['item=19019', 'item=19019']);
	});

	it('renders the notice beside the name and the extra labels after the enchant', () => {
		renderCell({ notice: <span data-testid="notice-marker" />, extraLabels: <span data-testid="extra-marker" /> });
		expect(screen.getByTestId('item-picker-name-row').querySelector('[data-testid="notice-marker"]')).not.toBeNull();
		expect(screen.getByTestId('item-picker-labels-container').lastElementChild!.getAttribute('data-testid')).toBe('extra-marker');
	});

	it('opens the items tab from the icon and the name, and each gem tab from its socket', () => {
		const onOpen = vi.fn();
		renderCell({ item: makeEquippedItem({ item: SOCKETED_ITEM }), onOpen, href: '' });

		fireEvent.click(icon());
		fireEvent.click(name());
		screen.getAllByTestId('gem-socket-container').forEach(socket => fireEvent.click(socket));
		expect(onOpen.mock.calls).toEqual([[SelectorModalTabs.Items], [SelectorModalTabs.Items], [SelectorModalTabs.Gem1], [SelectorModalTabs.Gem2]]);
	});

	it('anchors each socket at its own tooltip', () => {
		renderCell({ item: makeEquippedItem({ item: SOCKETED_ITEM }), socketTooltipId: gemIdx => `socket-${gemIdx}` });
		expect(screen.getAllByTestId('gem-socket-container').map(socket => socket.getAttribute('data-tooltip-id'))).toEqual(['socket-0', 'socket-1']);
	});

	it('sits the name and labels at the start by default', () => {
		renderCell();
		const root = screen.getByTestId('item-picker-root');
		expect(root.dataset.align).toBe('start');
		expect(root.className).not.toContain('justify-end');
	});

	it('sits the name and labels against the icon at the end for the right column', () => {
		renderCell({ align: 'end' });
		const root = screen.getByTestId('item-picker-root');
		expect(root.dataset.align).toBe('end');
		expect(root.classList).toContain('md:[&_.ui-item-picker-name-row]:justify-end');
		expect(root.classList).toContain('md:[&_.ui-item-picker-labels-container]:items-end');
	});
});
