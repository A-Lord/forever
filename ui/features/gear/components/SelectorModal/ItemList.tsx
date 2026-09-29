import { Class, ItemSlot } from '@generated/proto/common';
import { UIEnchant as Enchant, UIGem as Gem, UIItem as Item } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { SortDirection } from '@sim/constants/other';
import { useSimHost } from '@sim/context/SimHostContext';
import { useSimStore } from '@sim/hooks/useSimStore';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { useUiStore } from '@sim/hooks/useUiStore';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { subscribeSimField } from '@sim/state/subscriptions';
import { Tooltip } from '@ui-kit/Tooltip';
import { VirtualList } from '@ui-kit/VirtualList';
import clsx from 'clsx';
import { useCallback, useId, useMemo, useRef, useState } from 'react';

import { applyFavourite, isItemFavourited } from '../../model/favourites';
import { getItemIdByItemType } from '../../model/item_ids';
import { matchesSearch } from '../../model/item_search';
import { defaultSortBy, ItemListSortBy, sortItemIdxs } from '../../model/item_sort';
import { type ItemDataFields, type ItemListType, SelectorModalTabs } from '../../types';
import { FiltersMenu } from '../FiltersMenu';
import { ItemListHeaderRow } from './ItemListHeaderRow';
import { ItemListRow } from './ItemListRow';
import { type ItemListToggle, ItemListToolbarView } from './ItemListToolbarView';
import { removeButtonLabel, type SelectorTab } from './utils';

const ROW_HEIGHT = 56;
const ROW_CLASS = clsx(
	'relative flex items-center gap-4 bg-gray-900 p-2',
	'hover:bg-gray-800 data-[stripe=even]:not-hover:bg-table-odd data-[stripe=odd]:not-hover:bg-table-even',
	'[&_[data-active]_.ui-selector-modal-list-item-icon]:outline-2 [&_[data-active]_.ui-selector-modal-list-item-icon]:outline-success',
);

type FilterFlag = 'oneHandedWeapons' | 'twoHandedWeapons' | 'matchingGemsOnly';

const toggle = (name: string, label: string, value: boolean, onChange: (value: boolean) => void): ItemListToggle => ({
	id: `show-${name}-selector`,
	testId: `selector-modal-show-${name}`,
	label,
	value,
	onChange,
});

export interface ItemListProps {
	tab: SelectorTab;
	slot: ItemSlot;
	equippedItem: EquippedItem | null;
}

export const ItemList = ({ tab, slot, equippedItem }: ItemListProps) => {
	const host = useSimHost();
	const player = host.player;
	const sim = player.sim;
	const { label, itemData, computeEP, equippedToItem, onRemove, socketColor } = tab;

	const [search, setSearch] = useState('');
	const [sortBy, setSortBy] = useState(() => defaultSortBy(slot));
	const [sortDirection, setSortDirection] = useState(SortDirection.DESC);
	const listRef = useRef<HTMLDivElement>(null);
	const [filtersOpen, setFiltersOpen] = useState(false);
	const tooltipId = useId();

	const filters = useStoreSubscribe(subscribeSimField(sim, 'filters'), () => sim.getFilters());
	const phase = useSimStore('phase');
	const showEPValues = useUiStore('showEPValues');

	const isFavourited = useCallback((row: ItemDataFields<ItemListType>) => isItemFavourited(filters, label, row), [filters, label]);

	const itemsToDisplay = useMemo(() => {
		let idxs = itemData.map((_, index) => index);
		const currentEquippedItem = player.getEquippedItem(slot);

		if (label === SelectorModalTabs.Items) {
			idxs = player.filterItemData(idxs, index => itemData[index].item as unknown as Item, slot);
		} else if (label === SelectorModalTabs.Enchants) {
			idxs = player.filterEnchantData(idxs, index => itemData[index].item as unknown as Enchant, slot, currentEquippedItem);
		} else if (label === SelectorModalTabs.Gem1 || label === SelectorModalTabs.Gem2 || label === SelectorModalTabs.Gem3) {
			idxs = player.filterGemData(idxs, index => itemData[index].item as unknown as Gem, slot, socketColor);
		}

		idxs = idxs.filter(index => itemData[index].phase <= phase && matchesSearch(search, itemData[index], npcId => sim.db.getNpc(npcId)?.name));
		return sortItemIdxs(idxs, itemData, { sortBy, sortDirection, computeEP, isFavourited });
		// `filters` is named here even though the body never reads it: `player.filterItemData` and
		// `player.filterGemData` reach `sim.getFilters()` internally, so a filter change has to
		// rebuild this. The lint calls it unnecessary; deleting it leaves correctness resting on
		// `isFavourited` happening to close over `filters` too.
		// oxlint-disable-next-line react-hooks/exhaustive-deps
	}, [itemData, player, sim, slot, label, socketColor, phase, search, sortBy, sortDirection, computeEP, isFavourited, filters]);

	const equipped = equippedToItem(equippedItem);
	const equippedId = getItemIdByItemType(label, equipped);
	const equippedEP = equipped !== null && equipped !== undefined ? computeEP(equipped) : null;

	const sort = (next: ItemListSortBy) => {
		setSortDirection(current => (sortBy === next ? 1 - current : SortDirection.DESC));
		setSortBy(next);
	};

	const toggleFavourite = (row: ItemDataFields<ItemListType>) => {
		const next = sim.getFilters();
		if (!applyFavourite(next, label, row, !isFavourited(row))) return;
		sim.setFilters(next);
	};

	const setFilterFlag = (field: FilterFlag) => (value: boolean) => {
		const next = sim.getFilters();
		next[field] = value;
		sim.setFilters(next);
	};

	const showEPOptions = ![ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2].includes(slot);
	const showWeaponOptions =
		label === SelectorModalTabs.Items &&
		(slot === ItemSlot.ItemSlotMainHand || (slot === ItemSlot.ItemSlotOffHand && player.getClass() === Class.ClassWarrior));

	const toggles = [
		showWeaponOptions &&
			toggle('1h-weapons', i18n.t('settings_tab.other.show_1h_weapons.label'), filters.oneHandedWeapons, setFilterFlag('oneHandedWeapons')),
		showWeaponOptions &&
			toggle('2h-weapons', i18n.t('settings_tab.other.show_2h_weapons.label'), filters.twoHandedWeapons, setFilterFlag('twoHandedWeapons')),
		label.startsWith('Gem') &&
			toggle('matching-gems', i18n.t('settings_tab.other.show_matching_gems.label'), filters.matchingGemsOnly, setFilterFlag('matchingGemsOnly')),
		showEPOptions && toggle('ep-values', i18n.t('settings_tab.other.show_ep_values.label'), showEPValues, value => sim.setShowEPValues(value)),
	].filter((entry): entry is ItemListToggle => !!entry);

	return (
		<>
			<ItemListToolbarView
				search={search}
				onSearchChange={setSearch}
				filtersMenu={label === SelectorModalTabs.Items ? <FiltersMenu slot={slot} open={filtersOpen} onOpenChange={setFiltersOpen} /> : undefined}
				onOpenFilters={() => setFiltersOpen(true)}
				phase={phase}
				onPhaseChange={next => sim.setPhase(next)}
				toggles={toggles}
				removeLabel={removeButtonLabel(label, key => i18n.t(key))}
				onRemove={onRemove}
			/>
			<ItemListHeaderRow
				label={label}
				showEPValues={showEPValues}
				epTooltipId={`${tooltipId}-ep`}
				sortBy={sortBy}
				sortDirection={sortDirection}
				onSort={sort}
			/>
			<div
				ref={listRef}
				className="mb-0 max-h-[60vh] w-full overflow-x-hidden overflow-y-scroll p-0"
				data-testid="selector-modal-list"
				data-hide-ep={!showEPValues ? '' : undefined}
				tabIndex={0}>
				<VirtualList
					count={itemsToDisplay.length}
					rowHeight={ROW_HEIGHT}
					getScrollElement={() => listRef.current}
					rowClassName={() => ROW_CLASS}
					renderRow={index => {
						const row = itemData[itemsToDisplay[index]];
						return (
							<ItemListRow
								itemData={row}
								label={label}
								slot={slot}
								active={row.id === equippedId}
								equippedEP={equippedEP}
								itemEP={computeEP(row.item)}
								favourited={isFavourited(row)}
								favouriteTooltipId={`${tooltipId}-favourite`}
								compareTooltipId={`${tooltipId}-compare`}
								onEquip={() => row.onEquip(row.item)}
								onToggleFavourite={() => toggleFavourite(row)}
							/>
						);
					}}
				/>
			</div>
			{showEPOptions && <Tooltip id={`${tooltipId}-ep`} content={i18n.t('gear_tab.gear_picker.ep_tooltip')} />}
			<Tooltip
				id={`${tooltipId}-favourite`}
				render={({ activeAnchor }) =>
					(activeAnchor as HTMLElement | null)?.dataset.favourited === 'true' ? 'Remove from favorites' : 'Add to favorites'
				}
			/>
			<Tooltip
				id={`${tooltipId}-compare`}
				render={({ activeAnchor }) => ((activeAnchor as HTMLElement | null)?.dataset.inBatch === 'true' ? 'Remove from Batch Sim' : 'Add to Batch Sim')}
			/>
		</>
	);
};
