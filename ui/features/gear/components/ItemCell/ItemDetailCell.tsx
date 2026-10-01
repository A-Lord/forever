import { usePlayer } from '@sim/context/SimHostContext';
import { useActionId } from '@sim/hooks/useActionId';
import { useMemo } from 'react';

import { useEquippedItemWowheadDataset } from '../../hooks/useEquippedItemWowheadDataset';
import { MISSING_RANDOM_SUFFIX_WARNING } from '../../item_notices';
import { ItemNoticeIcon } from '../GearPicker/ItemNoticeIcon';
import { ItemDetailCellView, type ItemDetailCellViewProps } from './ItemDetailCellView';

export type ItemDetailCellProps = Omit<ItemDetailCellViewProps, 'iconUrl' | 'href' | 'wowhead' | 'notice'>;

/** Resolves the equipped item's icon, link, wowhead tooltip and notice for `ItemDetailCellView`. */
export const ItemDetailCell = (props: ItemDetailCellProps) => {
	const { item } = props;
	const player = usePlayer();

	const actionId = useMemo(() => item?.asActionId(), [item]);
	const { iconUrl, href } = useActionId(actionId);
	const wowhead = useEquippedItemWowheadDataset(player, item);

	return (
		<ItemDetailCellView
			{...props}
			iconUrl={iconUrl}
			href={href}
			wowhead={wowhead}
			notice={
				!!item && (
					<ItemNoticeIcon
						itemId={item.item.id}
						enchantId={item.enchant?.effectId}
						additionalNotice={item.hasRandomSuffixOptions() && !item.randomSuffix ? MISSING_RANDOM_SUFFIX_WARNING : undefined}
					/>
				)
			}
		/>
	);
};
