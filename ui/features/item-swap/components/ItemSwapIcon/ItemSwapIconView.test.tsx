import { getEmptySlotIconUrl } from '@features/gear/model/empty_slot_icons';
import { ItemSlot } from '@generated/proto/common';
import { translateSlotName } from '@i18n/localization';
import { makeEquippedItem, SOCKETED_ITEM } from '@testing/storybook/fixtures/gear';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { ItemSwapIconView, type ItemSwapIconViewProps } from './ItemSwapIconView';

const renderIcon = (props: Partial<ItemSwapIconViewProps> = {}) => {
	const onOpen = vi.fn();
	render(<ItemSwapIconView slot={ItemSlot.ItemSlotMainHand} item={null} iconUrl="" name="" href="" onOpen={onOpen} {...props} />);
	return { onOpen };
};

const icon = () => screen.getByTestId('icon-picker-button');

describe('ItemSwapIconView', () => {
	it('draws an empty slot as its placeholder, unfilled and named for the slot', () => {
		renderIcon();

		expect(icon().hasAttribute('data-active')).toBe(false);
		expect(icon().style.backgroundImage).toContain(getEmptySlotIconUrl(ItemSlot.ItemSlotMainHand));
		expect(icon().getAttribute('aria-label')).toBe(translateSlotName(ItemSlot.ItemSlotMainHand));
		expect(icon().hasAttribute('href')).toBe(false);
		expect(screen.getByTestId('item-picker-sockets-container').children).toHaveLength(0);
	});

	it('draws a filled slot active, with its icon, name, link and wowhead data', () => {
		renderIcon({
			item: makeEquippedItem(),
			iconUrl: 'sword.jpg',
			name: 'Thunderfury',
			href: 'https://wowhead.test/item=19019',
			wowhead: { 'data-wowhead': 'item=19019' },
		});

		expect(icon().hasAttribute('data-active')).toBe(true);
		expect(icon().style.backgroundImage).toContain('sword.jpg');
		expect(icon().getAttribute('aria-label')).toBe('Thunderfury');
		expect(icon().getAttribute('href')).toBe('https://wowhead.test/item=19019');
		expect(icon().dataset.whtticon).toBe('false');
		expect(icon().dataset.wowhead).toBe('item=19019');
	});

	it('hangs one socket per gem socket off the root, outside the icon', () => {
		renderIcon({ item: makeEquippedItem({ item: SOCKETED_ITEM }), iconUrl: 'sword.jpg' });

		const sockets = screen.getAllByTestId('gem-socket-container');
		expect(sockets).toHaveLength(2);
		for (const socket of sockets) expect(icon().contains(socket)).toBe(false);
	});

	it('opens the slot when the icon is clicked', () => {
		const { onOpen } = renderIcon();

		fireEvent.click(icon());
		expect(onOpen).toHaveBeenCalledOnce();
	});
});
