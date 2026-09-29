import { ConnectedBooleanPicker } from '@features/inputs/components/ConnectedBooleanPicker';
import type { BooleanPickerConfig } from '@features/inputs/configs';
import type { ItemSlot, Spec } from '@generated/proto/common';
import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import type { Player } from '@sim/player/player';
import { LocaleHtml } from '@ui-kit/Tooltip';
import { useMemo } from 'react';

import { swapWithGear } from '../../model/swap_with_gear';
import { ItemSwapIcon } from '../ItemSwapIcon';
import { ItemSwapPickerView } from './ItemSwapPickerView';

export interface ItemSwapPickerProps {
	itemSlots: ReadonlyArray<ItemSlot>;
	note?: string;
}

export const ItemSwapPicker = <SpecType extends Spec>({ itemSlots, note }: ItemSwapPickerProps) => {
	const host = useSimHost();
	const player = host.player as Player<SpecType>;
	const enabled = usePlayerStore('itemSwapEnabled');

	const enableConfig = useMemo(
		(): BooleanPickerConfig<Player<SpecType>> => ({
			id: 'enable-item-swap',
			reverse: true,
			label: i18n.t('settings_tab.other.enable_item_swap.label'),
			labelTooltip: <LocaleHtml html={i18n.t('settings_tab.other.enable_item_swap.tooltip')} />,
			layout: 'split',
			storeField: 'itemSwap',
			getValue: (subject: Player<SpecType>) => subject.itemSwapSettings.getEnableItemSwap(),
			setValue: (subject: Player<SpecType>, newValue: boolean) => subject.itemSwapSettings.setEnableItemSwap(newValue),
		}),
		[],
	);

	return (
		<div className="grid gap-2" data-testid="item-swap-picker-root">
			<ConnectedBooleanPicker modObject={player} config={enableConfig} />
			{enabled && (
				<ItemSwapPickerView note={note} onSwap={() => swapWithGear(player, itemSlots)}>
					{itemSlots.map(itemSlot => (
						<ItemSwapIcon key={itemSlot} slot={itemSlot} />
					))}
				</ItemSwapPickerView>
			)}
		</div>
	);
};
