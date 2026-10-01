import { Spec } from '@generated/proto/common';
import { usePlayer } from '@sim/context/SimHostContext';
import type { ReactNode } from 'react';

import { ENCHANT_NOTICES, ITEM_INFO_NOTICES, ITEM_NOTICES } from '../../item_notices';
import { ItemNoticeIconView } from './ItemNoticeIconView';

export interface ItemNoticeIconProps {
	itemId?: number;
	enchantId?: number;
	additionalNotice?: ReactNode;
}

export const ItemNoticeIcon = ({ itemId, enchantId, additionalNotice }: ItemNoticeIconProps) => {
	const spec = usePlayer().getSpec();

	const itemNotice = itemId ? ITEM_NOTICES.get(itemId) : undefined;
	const ownNotice = itemNotice?.[spec] || itemNotice?.[Spec.SpecUnknown];
	const enchantNotice = enchantId ? ENCHANT_NOTICES.get(enchantId) : undefined;
	const ownEnchantNotice = enchantNotice?.[spec] || enchantNotice?.[Spec.SpecUnknown];
	const infoNotice = itemId ? ITEM_INFO_NOTICES.get(itemId) : undefined;

	if (!ownNotice && !ownEnchantNotice && !additionalNotice && !infoNotice) return null;

	const warns = !!ownNotice || !!ownEnchantNotice || !!additionalNotice;

	return (
		<ItemNoticeIconView level={warns ? 'warning' : 'info'}>
			{ownNotice}
			{ownEnchantNotice}
			{additionalNotice}
			{infoNotice}
		</ItemNoticeIconView>
	);
};
