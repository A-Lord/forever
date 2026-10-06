import { ItemQuality, ScalingItemProperties } from '@generated/proto/common';
import type { UIEnchant, UIItem } from '@generated/proto/ui';
import { ActionId } from '@sim/proto/action_id';
import { EquippedItem } from '@sim/proto/equipped_item';

import type { ItemData, ItemListType } from '../../../features/gear/types';
import { makeUIItem } from './protos';

export const wowheadIconUrl = (file: string) => `https://wow.zamimg.com/images/wow/icons/large/${file}.jpg`;

export interface EquippedItemOptions {
	item?: Partial<UIItem>;
	enchant?: UIEnchant;
}

/**
 * An `EquippedItem` around `makeUIItem`, with the base scaling entry that `EquippedItem.ilvl` reads
 * filled from the item's own `ilvl`. Its action id already carries the fixture's name and icon.
 */
export const makeEquippedItem = ({ item = {}, enchant }: EquippedItemOptions = {}): EquippedItem => {
	const base = makeUIItem(item);
	base.scalingOptions = { 0: ScalingItemProperties.create({ ilvl: base.ilvl }), ...base.scalingOptions };
	const actionId = Object.assign(ActionId.fromItemId(base.id), {
		baseName: base.name,
		name: base.name,
		iconUrl: wowheadIconUrl(base.icon),
	});
	return Object.assign(new EquippedItem({ item: base, enchant }), { asActionId: () => actionId });
};

/** One selector-modal row for Thunderfury. A view takes the icon resolved, so nothing fills `actionId`. */
export const makeItemData = (overrides: Partial<ItemData<ItemListType, string>> = {}): ItemData<ItemListType, string> => {
	const item = makeUIItem();
	return {
		item,
		id: item.id,
		actionId: ActionId.fromItemId(item.id),
		name: item.name,
		searchText: item.name,
		quality: ItemQuality.ItemQualityLegendary,
		phase: 1,
		ilvl: item.ilvl,
		ignoreEPFilter: false,
		nameDescription: '',
		onEquip: () => undefined,
		...overrides,
	};
};
