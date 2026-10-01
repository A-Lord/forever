import { ItemQuality, ItemSlot, type ItemSpec } from '@generated/proto/common';
import type { UIItem as Item } from '@generated/proto/ui';
import { renderWithHost } from '@testing/react';
import { fireEvent } from '@testing-library/react';
import type { ReactNode } from 'react';
import { describe, expect, it, vi } from 'vitest';

import type { ItemData, ItemListType } from '../../types';
import { SelectorModalTabs } from '../../types';

// The batch seam is mocked rather than driven: the row only asks whether the item is in the batch
// and calls one of two writes, and both are plain functions over the player now.
const batch = vi.hoisted(() => ({
	hasItem: vi.fn((_spec?: unknown) => false),
	addItem: vi.fn((_spec?: unknown) => undefined),
	removeItem: vi.fn((_spec?: unknown) => undefined),
}));

const store = vi.hoisted(() => {
	const listeners = new Set<() => void>();
	const state = {};
	return {
		subscribe: (callback: () => void) => {
			listeners.add(callback);
			return () => listeners.delete(callback);
		},
		getState: () => state,
		getInitialState: () => state,
	};
});

vi.mock('@sim/hooks/useActionId', () => ({
	useActionId: () => ({ iconUrl: 'icon-url.png', name: 'item-name', href: 'https://example.com/item', ready: true }),
}));
vi.mock('@sim/state/subscriptions', async () => (await import('@sim/testing')).mockSubscriptions(store.subscribe));
vi.mock('@features/bulk/model/items', () => ({
	hasBulkItem: (_player: unknown, spec: ItemSpec) => batch.hasItem(spec),
	addBulkItem: (_player: unknown, spec: ItemSpec) => batch.addItem(spec),
	removeBulkItem: (_player: unknown, spec: ItemSpec) => batch.removeItem(spec),
}));
vi.mock('../../../../tracking/utils', () => ({ trackEvent: vi.fn() }));
vi.mock('./ItemSource', () => ({ ItemSource: () => <div data-testid="item-source-marker" /> }));
vi.mock('../GearPicker/ItemNoticeIcon', () => ({
	ItemNoticeIcon: ({ itemId, enchantId }: { itemId?: number; enchantId?: number }) => (
		<div data-testid="item-notice-marker" data-item-id={itemId} data-enchant-id={enchantId} />
	),
}));

const { ItemListRow } = await import('./ItemListRow');
type ItemListRowProps = Parameters<typeof ItemListRow>[0];

const makeItemData = (overrides: Partial<ItemData<ItemListType, ReactNode>> = {}): ItemData<ItemListType, ReactNode> => ({
	item: { id: 1, ilvl: 0 } as Item,
	searchText: '',
	id: 1,
	actionId: {} as ItemData<ItemListType, ReactNode>['actionId'],
	quality: ItemQuality.ItemQualityEpic,
	phase: 1,
	ignoreEPFilter: false,
	nameDescription: '',
	onEquip: vi.fn(),
	name: 'Test Item',
	...overrides,
});

const renderRow = (
	overrides: Partial<ItemListRowProps> = {},
	batchOverrides: Partial<{ hasItem: ReturnType<typeof vi.fn>; addItem: ReturnType<typeof vi.fn>; removeItem: ReturnType<typeof vi.fn> }> = {},
) => {
	Object.assign(batch, { hasItem: vi.fn(() => false), addItem: vi.fn(), removeItem: vi.fn() }, batchOverrides);

	const props: ItemListRowProps = {
		itemData: makeItemData(),
		label: SelectorModalTabs.Items,
		slot: ItemSlot.ItemSlotHead,
		active: false,
		equippedEP: null,
		itemEP: 0,
		favourited: false,
		favouriteTooltipId: 'fav-tooltip',
		compareTooltipId: 'compare-tooltip',
		onEquip: vi.fn(),
		onToggleFavourite: vi.fn(),
		...overrides,
	};

	const { container, rerender } = renderWithHost(<ItemListRow {...props} />, { sim: { store } });

	return {
		container,
		batch,
		props,
		rerender: (next: Partial<ItemListRowProps>) => rerender(<ItemListRow {...props} {...next} />),
	};
};

describe('ItemListRow', () => {
	it('omits the ilvl cell outside the items and upgrades tabs', () => {
		const { container } = renderRow({ label: SelectorModalTabs.Enchants });
		expect(container.querySelector('[data-testid="selector-modal-list-item-ilvl-container"]')).toBeNull();
	});

	it.each([
		['prefers the itemData ilvl over the item ilvl', 500, 400, '500'],
		['falls back to the item ilvl when the itemData ilvl is unset', 0, 450, '450'],
	])('%s', (_name, ilvl, itemIlvl, shown) => {
		const { container } = renderRow({ itemData: makeItemData({ ilvl, item: { id: 1, ilvl: itemIlvl } as Item }) });
		expect(container.querySelector('[data-testid="selector-modal-list-item-ilvl-container"]')?.textContent).toBe(shown);
	});

	it('renders the item source cell only on the items tab', () => {
		const items = renderRow({ label: SelectorModalTabs.Items });
		expect(items.container.querySelector('[data-testid="selector-modal-list-item-source-container"]')).not.toBeNull();

		const enchants = renderRow({ label: SelectorModalTabs.Enchants });
		expect(enchants.container.querySelector('[data-testid="selector-modal-list-item-source-container"]')).toBeNull();
	});

	it('looks the notice up by enchant id on the enchants tab and by item id elsewhere', () => {
		const enchants = renderRow({ label: SelectorModalTabs.Enchants, itemData: makeItemData({ id: 1894 }) });
		const enchantNotice = enchants.container.querySelector('[data-testid="item-notice-marker"]')!;
		expect(enchantNotice.getAttribute('data-enchant-id')).toBe('1894');
		expect(enchantNotice.hasAttribute('data-item-id')).toBe(false);

		const items = renderRow({ label: SelectorModalTabs.Items, itemData: makeItemData({ id: 17182 }) });
		const itemNotice = items.container.querySelector('[data-testid="item-notice-marker"]')!;
		expect(itemNotice.getAttribute('data-item-id')).toBe('17182');
		expect(itemNotice.hasAttribute('data-enchant-id')).toBe(false);
	});

	it('hands the ep and equipped ep to the ep cell', () => {
		const { container } = renderRow({ itemEP: 80, equippedEP: 50 });
		expect(container.querySelector('[data-testid="selector-modal-list-item-ep-value"]')?.textContent).toBe('80');
		expect(container.querySelector('[data-testid="selector-modal-list-item-ep-delta"]')?.getAttribute('data-sign')).toBe('positive');
	});

	it('hides the ep cell for trinket slots', () => {
		const trinket1 = renderRow({ slot: ItemSlot.ItemSlotTrinket1 });
		expect(trinket1.container.querySelector('[data-testid="selector-modal-list-item-ep"]')).toBeNull();

		const trinket2 = renderRow({ slot: ItemSlot.ItemSlotTrinket2 });
		expect(trinket2.container.querySelector('[data-testid="selector-modal-list-item-ep"]')).toBeNull();

		const head = renderRow({ slot: ItemSlot.ItemSlotHead });
		expect(head.container.querySelector('[data-testid="selector-modal-list-item-ep"]')).not.toBeNull();
	});

	it('wires the icon and anchor to useActionId', () => {
		const { container } = renderRow();
		expect(container.querySelector('[data-testid="selector-modal-list-item-icon"]')?.getAttribute('src')).toBe('icon-url.png');
		expect(container.querySelector('[data-testid="selector-modal-list-item-link"]')?.getAttribute('href')).toBe('https://example.com/item');
	});

	it('reflects whether the item is already in the bulk batch', () => {
		const inBatch = renderRow({}, { hasItem: vi.fn(() => true) });
		expect(inBatch.container.querySelector('[data-testid="selector-modal-list-item-compare"]')?.getAttribute('data-in-batch')).toBe('true');

		const notInBatch = renderRow({}, { hasItem: vi.fn(() => false) });
		expect(notInBatch.container.querySelector('[data-testid="selector-modal-list-item-compare"]')?.getAttribute('data-in-batch')).toBe('false');
	});

	it('adds the item to the batch when absent, and removes it when present', () => {
		const notInBatch = renderRow({ itemData: makeItemData({ id: 7 }) }, { hasItem: vi.fn(() => false) });
		fireEvent.click(notInBatch.container.querySelector('[data-testid="selector-modal-list-item-compare"]')!);
		expect(notInBatch.batch.addItem).toHaveBeenCalledTimes(1);
		expect(notInBatch.batch.removeItem).not.toHaveBeenCalled();

		const inBatch = renderRow({ itemData: makeItemData({ id: 7 }) }, { hasItem: vi.fn(() => true) });
		fireEvent.click(inBatch.container.querySelector('[data-testid="selector-modal-list-item-compare"]')!);
		expect(inBatch.batch.removeItem).toHaveBeenCalledTimes(1);
		expect(inBatch.batch.addItem).not.toHaveBeenCalled();
	});

	it('re-reads the batch flag when a recycled row is handed a different item', () => {
		const { container, rerender } = renderRow({ itemData: makeItemData({ id: 1 }) }, { hasItem: vi.fn((spec: any) => spec.id === 2) });
		const compare = () => container.querySelector('[data-testid="selector-modal-list-item-compare"]')!;
		expect(compare().getAttribute('data-in-batch')).toBe('false');

		rerender({ itemData: makeItemData({ id: 2 }) });

		expect(compare().getAttribute('data-in-batch')).toBe('true');
	});

	it('acts on the item the recycled row now shows, not the one it showed before', () => {
		const { container, batch, rerender } = renderRow({ itemData: makeItemData({ id: 1 }) }, { hasItem: vi.fn((spec: any) => spec.id === 1) });

		rerender({ itemData: makeItemData({ id: 2 }) });
		fireEvent.click(container.querySelector('[data-testid="selector-modal-list-item-compare"]')!);

		expect(batch.addItem).toHaveBeenCalledTimes(1);
		expect(batch.removeItem).not.toHaveBeenCalled();
	});

	it('renders the compare container on the items tab alone', () => {
		const items = renderRow({ label: SelectorModalTabs.Items });
		expect(items.container.querySelector('[data-testid="selector-modal-list-item-compare"]')).not.toBeNull();

		const enchants = renderRow({ label: SelectorModalTabs.Enchants });
		expect(enchants.container.querySelector('[data-testid="selector-modal-list-item-compare"]')).toBeNull();
	});
});
