import type { ItemSlot } from '@generated/proto/common';
import { useSimHost } from '@sim/context/SimHostContext';
import { useActionId } from '@sim/hooks/useActionId';
import { ActionId } from '@sim/proto/action_id';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { useMemo } from 'react';

import { useEquippedItemWowheadDataset } from '../../hooks/useEquippedItemWowheadDataset';
import { GearChangeIconView } from './GearChangeIconView';
import { gearChangeEnchant, gearChangeSockets } from './utils';

export interface GearChangeIconProps {
	slot: ItemSlot;
	/** The slot after the run. Undefined renders the empty frame, which is what an unfilled slot looks like. */
	item?: EquippedItem;
	previousItem?: EquippedItem;
}

export const GearChangeIcon = ({ slot, item, previousItem }: GearChangeIconProps) => {
	const player = useSimHost().player;

	const actionId = useMemo(() => item?.asActionId(), [item]);
	const { iconUrl, href } = useActionId(actionId);
	const wowhead = useEquippedItemWowheadDataset(player, item);
	const sockets = useMemo(() => gearChangeSockets(item, previousItem), [item, previousItem]);

	const changedEnchant = useMemo(() => gearChangeEnchant(item, previousItem), [item, previousItem]);
	const enchantActionId = useMemo(
		() =>
			changedEnchant ? (changedEnchant.spellId ? ActionId.fromSpellId(changedEnchant.spellId) : ActionId.fromItemId(changedEnchant.itemId)) : undefined,
		[changedEnchant],
	);
	const { iconUrl: enchantIconUrl } = useActionId(enchantActionId);
	const enchant = changedEnchant ? { name: changedEnchant.name, iconUrl: enchantIconUrl } : undefined;

	return <GearChangeIconView slot={slot} equipped={!!item} iconUrl={iconUrl} href={href} wowhead={wowhead} sockets={sockets} enchant={enchant} />;
};
