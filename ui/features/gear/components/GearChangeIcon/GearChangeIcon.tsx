import type { ItemSlot } from '@generated/proto/common';
import { useSimHost } from '@sim/context/SimHostContext';
import { useActionId } from '@sim/hooks/useActionId';
import { useActionIdWowheadDataset } from '@sim/hooks/useActionIdWowheadDataset';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { useMemo } from 'react';

import { useEquippedItemWowheadDataset } from '../../hooks/useEquippedItemWowheadDataset';
import { enchantActionId } from '../../model/item_data';
import { GearChangeIconView } from './GearChangeIconView';
import { enchantChanged, gearChangeSockets } from './utils';

export interface GearChangeIconProps {
	slot: ItemSlot;
	/** The slot after the run. Undefined renders the empty frame, which is what an unfilled slot looks like. */
	item?: EquippedItem;
	previousItem?: EquippedItem;
	/** Draws the item's enchant marker, bordered when the enchant changed. */
	showEnchant?: boolean;
}

export const GearChangeIcon = ({ slot, item, previousItem, showEnchant = false }: GearChangeIconProps) => {
	const player = useSimHost().player;

	const actionId = useMemo(() => item?.asActionId(), [item]);
	const { iconUrl, href } = useActionId(actionId);
	const wowhead = useEquippedItemWowheadDataset(player, item);
	const sockets = useMemo(() => gearChangeSockets(item, previousItem), [item, previousItem]);

	const itemEnchant = showEnchant ? (item?.enchant ?? undefined) : undefined;
	const enchantActionIdOf = useMemo(() => (showEnchant && item?.enchant ? enchantActionId(item.enchant) : undefined), [showEnchant, item]);
	const { iconUrl: enchantIconUrl, href: enchantHref } = useActionId(enchantActionIdOf);
	const enchantWowhead = useActionIdWowheadDataset(enchantActionIdOf);
	const enchant = itemEnchant ? { name: itemEnchant.name, iconUrl: enchantIconUrl, href: enchantHref } : undefined;

	return (
		<GearChangeIconView
			slot={slot}
			equipped={!!item}
			iconUrl={iconUrl}
			href={href}
			wowhead={wowhead}
			sockets={sockets}
			enchant={enchant}
			enchantChanged={enchantChanged(item, previousItem)}
			enchantWowhead={enchantWowhead}
		/>
	);
};
