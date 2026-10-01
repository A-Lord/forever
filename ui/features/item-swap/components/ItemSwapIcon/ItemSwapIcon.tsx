import { useEquippedItemWowheadDataset } from '@features/gear/hooks/useEquippedItemWowheadDataset';
import { useOpenSelectorModal } from '@features/gear/hooks/useSelectorModal';
import { SelectorModalTabs } from '@features/gear/types';
import type { ItemSlot } from '@generated/proto/common';
import { useSimHost } from '@sim/context/SimHostContext';
import { useActionId } from '@sim/hooks/useActionId';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { useMemo } from 'react';

import { createItemSwapGearData } from '../../model/gear_data';
import { ItemSwapIconView } from './ItemSwapIconView';

export interface ItemSwapIconProps {
	slot: ItemSlot;
}

export const ItemSwapIcon = ({ slot }: ItemSwapIconProps) => {
	const host = useSimHost();
	const openSelectorModal = useOpenSelectorModal();
	const player = host.player;

	const item = usePlayerStore('itemSwapGear').getEquippedItem(slot);

	const actionId = useMemo(() => item?.asActionId(), [item]);
	const { iconUrl, name, href } = useActionId(actionId);

	const wowheadProps = useEquippedItemWowheadDataset(player, item);

	return (
		<ItemSwapIconView
			slot={slot}
			item={item}
			iconUrl={iconUrl}
			name={name}
			href={href}
			wowhead={wowheadProps}
			onOpen={() => openSelectorModal(slot, SelectorModalTabs.Items, createItemSwapGearData(player, slot))}
		/>
	);
};
