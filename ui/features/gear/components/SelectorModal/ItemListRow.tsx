import { addBulkItem, hasBulkItem, removeBulkItem } from '@features/bulk/model/items';
import { ItemSlot, ItemSpec } from '@generated/proto/common';
import { UIItem as Item } from '@generated/proto/ui';
import { useSimHost } from '@sim/context/SimHostContext';
import { useActionId } from '@sim/hooks/useActionId';
import { isIndividualSimHost } from '@sim/sim_host';
import { type ReactNode, useMemo } from 'react';
import { useStore } from 'zustand';

import { trackEvent } from '../../../../tracking/utils';
import type { ItemData, ItemListType } from '../../types';
import { SelectorModalTabs } from '../../types';
import { ItemNoticeIcon } from '../GearPicker/ItemNoticeIcon';
import { ItemListRowView } from './ItemListRowView';
import { ItemSource } from './ItemSource';

export interface ItemListRowProps {
	itemData: ItemData<ItemListType, ReactNode>;
	label: SelectorModalTabs;
	slot: ItemSlot;
	active: boolean;
	equippedEP: number | null;
	itemEP: number;
	favourited: boolean;
	favouriteTooltipId: string;
	compareTooltipId: string;
	onEquip: () => void;
	onToggleFavourite: () => void;
}

export const ItemListRow = ({
	itemData,
	label,
	slot,
	active,
	equippedEP,
	itemEP,
	favourited,
	favouriteTooltipId,
	compareTooltipId,
	onEquip,
	onToggleFavourite,
}: ItemListRowProps) => {
	const host = useSimHost();
	const { iconUrl, href } = useActionId(itemData.actionId);
	const showEp = ![ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2].includes(slot);
	const isItemsTab = label === SelectorModalTabs.Items;
	const isEnchantsTab = label === SelectorModalTabs.Enchants;

	const batchPlayer = isIndividualSimHost(host) ? host.player : null;
	const batchSpec = useMemo(() => ItemSpec.create({ id: itemData.id }), [itemData.id]);
	// Read through zustand rather than `useStoreSubscribe`: the answer is keyed by a prop, and that
	// hook re-reads on a notification only, so a recycled row would keep the previous item's flag.
	const inBatch = useStore(host.sim.store, () => !!batchPlayer && hasBulkItem(batchPlayer, batchSpec));

	const toggleCompare = () => {
		const present = !!batchPlayer && hasBulkItem(batchPlayer, batchSpec);
		if (batchPlayer) (present ? removeBulkItem : addBulkItem)(batchPlayer, batchSpec);
		trackEvent({ action: 'click', category: 'batch', label: present ? 'remove-item' : 'add-item' });
	};

	return (
		<ItemListRowView
			name={itemData.name}
			quality={itemData.quality}
			nameDescription={itemData.nameDescription}
			iconUrl={iconUrl}
			href={href}
			ilvl={isItemsTab ? itemData.ilvl || (itemData.item as unknown as Item).ilvl : undefined}
			active={active}
			onEquip={onEquip}
			notice={<ItemNoticeIcon itemId={isEnchantsTab ? undefined : itemData.id} enchantId={isEnchantsTab ? itemData.id : undefined} />}
			source={isItemsTab ? <ItemSource item={itemData.item as unknown as Item} sim={host.sim} /> : undefined}
			ep={showEp ? { ep: itemEP, equippedEP } : undefined}
			favourited={favourited}
			favouriteTooltipId={favouriteTooltipId}
			onToggleFavourite={onToggleFavourite}
			compare={isItemsTab ? { inBatch, tooltipId: compareTooltipId, onToggle: toggleCompare } : undefined}
		/>
	);
};
