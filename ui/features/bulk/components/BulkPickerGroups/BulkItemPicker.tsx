import { useOpenSelectorModal } from '@features/gear/hooks/useSelectorModal';
import type { SelectorModalTabs } from '@features/gear/types';
import type { ItemSlot } from '@generated/proto/common';
import { BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS, BulkSimItemSlot } from '@sim/bulk/utils';
import { usePlayer } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { getEligibleItemSlots } from '@sim/proto/items';

import { useBulkState } from '../../hooks/useBulkState';
import { createBulkGearData } from '../../model/gear_data';
import { removeBulkItemByIndex } from '../../model/items';
import { frozenItemSlot } from '../../model/picker_groups';
import { BulkItemPickerView } from './BulkItemPickerView';

export interface BulkItemPickerProps {
	bulkSlot: BulkSimItemSlot;
	/** Below zero for the two equipped slots the group covers; those report what is worn rather than offering a choice. */
	index: number;
	item: EquippedItem;
}

/** The physical slot an equipped entry stands for. Null for a user-added one. */
const equippedSlotOf = (bulkSlot: BulkSimItemSlot, index: number): ItemSlot | null => {
	if (index >= 0) return null;
	const slots = BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS.get(bulkSlot);
	if (!slots) return null;
	return index === -1 ? slots[0] : slots[1];
};

export const BulkItemPicker = ({ bulkSlot, index, item }: BulkItemPickerProps) => {
	const openSelectorModal = useOpenSelectorModal();
	const player = usePlayer();

	const gear = usePlayerStore('gear');
	const frozenItems = useBulkState(slice => slice.frozenItems);
	const frozenWeaponSlot = useBulkState(slice => slice.frozenWeaponSlot);

	const ownSlot = equippedSlotOf(bulkSlot, index);
	const frozenBulkSlot = frozenItemSlot(gear, BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS.get(bulkSlot), frozenItems.get(bulkSlot));
	const isCurrentlyEquipped = bulkSlot !== BulkSimItemSlot.ItemSlotHandWeapon && player.getEquippedItems().some(equipped => equipped?.id === item.id);
	const isEditable = index >= 0 && !isCurrentlyEquipped;
	const isFrozen = !ownSlot ? false : ownSlot === frozenWeaponSlot ? (gear.getEquippedItem(ownSlot)?.equals(item) ?? false) : ownSlot === frozenBulkSlot;

	const slot = getEligibleItemSlots(item.item)[0];

	return (
		<BulkItemPickerView
			slot={slot}
			item={item}
			state={isFrozen ? 'frozen' : isEditable ? 'editable' : 'equipped'}
			onOpen={(tab: SelectorModalTabs) => {
				if (!isEditable) return;
				openSelectorModal(slot, tab, createBulkGearData(player, bulkSlot, index));
			}}
			onRemove={isEditable ? () => removeBulkItemByIndex(player, index) : undefined}
		/>
	);
};
