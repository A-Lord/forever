import { OpenSelectorModalContext, type SelectorModalState, useSelectorModalState } from '@features/gear/hooks/useSelectorModal';
import { SelectorModalTabs } from '@features/gear/types';
import { ItemSlot, ItemSpec } from '@generated/proto/common';
import { BulkSimItemSlot } from '@sim/bulk/constants_auto_gen';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { bulkState, seedBulkSettings } from '@sim/settings/bulk_settings';
import { createSimStore, PLAYER_FIELDS, seedKeyed, zeroVersions } from '@sim/state/sim_store';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@features/gear/components/ItemCell', () => ({
	ItemDetailCell: ({ rootDataAttributes, testId }: { rootDataAttributes?: Record<string, string>; testId?: string }) => (
		<div data-testid={testId} {...rootDataAttributes} />
	),
}));

vi.mock('@sim/proto/items', async () => {
	const actual = await vi.importActual<typeof import('@sim/proto/items')>('@sim/proto/items');
	return { ...actual, getEligibleItemSlots: () => [ItemSlot.ItemSlotFinger1], canEquipItem: () => true };
});

vi.mock('@sim/bulk/utils', async () => {
	const actual = await vi.importActual<typeof import('@sim/bulk/utils')>('@sim/bulk/utils');
	return { ...actual, getBulkPlayerCanDualWield: () => true };
});

const { BulkPickerGroups } = await import('./BulkPickerGroups');
const { hasBulkItem, seedBulkPickerGroups } = await import('../../model/items');

const STORE_KEY = 7;

const fakeItem = (id: number, name: string): EquippedItem => {
	const item = {
		id,
		item: { id, name, unique: false },
		_item: { id, name, unique: false },
		asSpec: () => ItemSpec.create({ id, gems: [0, 0] }),
		equals: (other: EquippedItem | null) => !!other && other.id === id,
		withDynamicStats: () => item as unknown as EquippedItem,
	};
	return item as unknown as EquippedItem;
};

const FIRST = fakeItem(1, 'First Ring');
const SECOND = fakeItem(2, 'Second Ring');

const mount = () => {
	const store = createSimStore();
	const gear = { getEquippedItem: () => null };
	seedKeyed(store, 'players', STORE_KEY, { gear, v: zeroVersions(PLAYER_FIELDS) } as never);

	const db = { lookupItemSpec: (spec: ItemSpec) => [FIRST, SECOND].find(item => item.id === spec.id) ?? null };
	const player = {
		sim: { store, db, isNative: false, encounter: { getAreaTypes: () => [] } },
		storeKey: STORE_KEY,
		getGear: () => gear,
		getEquippedItems: () => [],
		getSpec: () => 0,
		getPlayerSpec: () => ({ friendlyName: 'test' }),
	} as unknown as Player<any>;

	seedBulkSettings(player);
	seedBulkPickerGroups(player);

	let modal: SelectorModalState | null = null;
	const Harness = () => {
		modal = useSelectorModalState();
		return (
			<OpenSelectorModalContext value={modal.openTab}>
				<BulkPickerGroups />
			</OpenSelectorModalContext>
		);
	};

	const host = { player, sim: player.sim } as never;
	const { container } = render(
		<SimHostProvider host={host}>
			<Harness />
		</SimHostProvider>,
	);

	const clickAdd = (group: string) =>
		fireEvent.click(
			within(container.querySelector<HTMLElement>(`[data-gear-group="${group}"]`)!).getByRole('button', { name: 'bulk_tab.picker.add_item' }),
		);

	return { player, clickAdd, modal: () => modal! };
};

describe('BulkItemPickerGroup add button', () => {
	it.each([
		['rings', ItemSlot.ItemSlotFinger1],
		['trinkets', ItemSlot.ItemSlotTrinket1],
		['weapons', ItemSlot.ItemSlotMainHand],
		['head', ItemSlot.ItemSlotHead],
	])('opens the item list for the %s group on its slot', (group, itemSlot) => {
		const h = mount();

		h.clickAdd(group);

		expect(h.modal().open).toBe(true);
		expect(h.modal().request).toMatchObject({ slot: itemSlot, tab: SelectorModalTabs.Items });
		expect(h.modal().request!.gearData.getEquippedItem()).toBeNull();
	});

	it('adds each picked item to the group and leaves the modal open', () => {
		const h = mount();
		h.clickAdd('rings');
		const { gearData } = h.modal().request!;

		act(() => gearData.equipItem(FIRST));
		act(() => gearData.equipItem(SECOND));

		const rings = bulkState(h.player).pickerGroups.get(BulkSimItemSlot.ItemSlotFinger)!;
		expect(rings.map(entry => entry.item)).toEqual([FIRST, SECOND]);
		expect(screen.getAllByTestId('bulk-item-picker')).toHaveLength(2);
		expect(h.modal().open).toBe(true);
		expect(gearData.getEquippedItem()).toBeNull();
	});

	it("adds a socketed pick in the shape the list's in-batch mark looks for", () => {
		const h = mount();
		h.clickAdd('rings');

		act(() => h.modal().request!.gearData.equipItem(FIRST));

		expect(hasBulkItem(h.player, ItemSpec.create({ id: FIRST.id }))).toBe(true);
	});
});
