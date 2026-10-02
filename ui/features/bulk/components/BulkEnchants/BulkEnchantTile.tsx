import { useEnchantDescription } from '@features/gear/hooks/useEnchantDescription';
import { enchantActionId } from '@features/gear/model/item_data';
import type { ItemSlot } from '@generated/proto/common';
import type { UIEnchant as Enchant } from '@generated/proto/ui';
import { useActionId } from '@sim/hooks/useActionId';
import { useActionIdWowheadDataset } from '@sim/hooks/useActionIdWowheadDataset';
import { memo, useCallback, useMemo } from 'react';

import { BulkEnchantTileView } from './BulkEnchantTileView';

export interface BulkEnchantTileProps {
	slot: ItemSlot;
	enchant: Enchant;
	checked: boolean;
	onToggle: (slot: ItemSlot, effectId: number) => void;
}

export const BulkEnchantTile = memo(({ slot, enchant, checked, onToggle }: BulkEnchantTileProps) => {
	const actionId = useMemo(() => enchantActionId(enchant), [enchant]);
	const { iconUrl, href } = useActionId(actionId);
	const wowhead = useActionIdWowheadDataset(actionId);
	const description = useEnchantDescription(enchant);
	const toggle = useCallback(() => onToggle(slot, enchant.effectId), [onToggle, slot, enchant]);

	return (
		<BulkEnchantTileView
			name={enchant.name}
			effect={description === enchant.name ? undefined : description}
			iconUrl={iconUrl}
			href={href}
			wowhead={wowhead}
			checked={checked}
			onToggle={toggle}
		/>
	);
});
