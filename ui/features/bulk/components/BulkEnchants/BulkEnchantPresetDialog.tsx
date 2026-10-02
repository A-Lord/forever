import { getEmptySlotIconUrl } from '@features/gear/model/empty_slot_icons';
import { byLvlDescending } from '@features/gear/model/item_data';
import { formatSearchQuery } from '@features/gear/model/item_search';
import { ItemSlot } from '@generated/proto/common';
import type { BulkEnchantPreset, UIEnchant as Enchant } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { translateBulkSlotName } from '@i18n/localization';
import { ITEM_SLOT_TO_BULK_SIM_ITEM_SLOT } from '@sim/bulk/utils';
import { usePlayer } from '@sim/context/SimHostContext';
import { useSimReady } from '@sim/hooks/useSimReady';
import { useSimStore } from '@sim/hooks/useSimStore';
import { isOffHandOnlyEnchant } from '@sim/proto/items';
import { useCallback, useMemo, useState } from 'react';

import { useBulkState } from '../../hooks/useBulkState';
import { BULK_ENCHANT_SLOTS, saveBulkEnchantPreset, validateBulkEnchantPresetName } from '../../model/enchants';
import { BulkEnchantPresetDialogView } from './BulkEnchantPresetDialogView';
import { BulkEnchantTile } from './BulkEnchantTile';

export interface BulkEnchantPresetDialogProps {
	open: boolean;
	/** Null for a new preset. */
	preset: BulkEnchantPreset | null;
	onClose: () => void;
}

const slotEnchants = (slot: ItemSlot, enchants: Enchant[]): Enchant[] => {
	if (slot === ItemSlot.ItemSlotMainHand) return enchants.filter(enchant => !isOffHandOnlyEnchant(enchant));
	if (slot === ItemSlot.ItemSlotOffHand) return enchants.filter(isOffHandOnlyEnchant);
	return enchants;
};

const slotLabel = (slot: ItemSlot): string => {
	if (slot === ItemSlot.ItemSlotMainHand) return i18n.t('bulk_tab.settings.enchant_presets.slot_weapons');
	if (slot === ItemSlot.ItemSlotOffHand) return i18n.t('bulk_tab.settings.enchant_presets.slot_off_hand');
	return translateBulkSlotName(ITEM_SLOT_TO_BULK_SIM_ITEM_SLOT.get(slot)!);
};

export const BulkEnchantPresetDialog = ({ open, preset, onClose }: BulkEnchantPresetDialogProps) => {
	const player = usePlayer();
	const presets = useBulkState(slice => slice.enchantPresets);
	const [name, setName] = useState(preset?.name ?? '');
	const [selected, setSelected] = useState<ReadonlyMap<ItemSlot, readonly number[]>>(
		() => new Map(preset?.slots.map(options => [options.slot, options.effectIds]) ?? []),
	);

	const ready = useSimReady();
	const slots = useMemo(
		() =>
			ready
				? BULK_ENCHANT_SLOTS.map(slot => ({
						slot,
						enchants: slotEnchants(slot, [...player.getEnchants(slot)].sort(byLvlDescending)).map(enchant => ({
							enchant,
							relevant: player.isEnchantRelevant(enchant),
							searchName: formatSearchQuery(enchant.name),
						})),
					})).filter(({ enchants }) => enchants.length)
				: [],
		[player, ready],
	);

	const [openSlots, setOpenSlots] = useState(() => preset?.slots.filter(options => options.effectIds.length).map(options => String(options.slot)) ?? []);
	const [search, setSearch] = useState('');
	// A search opens every slot it matches in, less those closed by hand since the search text last changed. The slots
	// opened by hand keep their own state, so clearing the search brings them back.
	const [searchClosedSlots, setSearchClosedSlots] = useState<ReadonlySet<string>>(() => new Set());
	const showAll = useSimStore('filters').showAllEnchants;
	const changeShowAll = (next: boolean) => {
		const filters = player.sim.getFilters();
		filters.showAllEnchants = next;
		player.sim.setFilters(filters);
	};
	const terms = useMemo(() => formatSearchQuery(search).split(' ').filter(Boolean), [search]);
	const shown = useMemo(
		() =>
			slots
				.map(({ slot, enchants }) => ({
					slot,
					enchants: enchants
						.filter(
							({ enchant, relevant, searchName }) =>
								(showAll || relevant || !!selected.get(slot)?.includes(enchant.effectId)) && terms.every(term => searchName.includes(term)),
						)
						.map(({ enchant }) => enchant),
				}))
				.filter(({ enchants }) => enchants.length),
		[slots, terms, showAll, selected],
	);
	const shownSlots = shown.map(({ slot }) => String(slot));
	const changeSearch = (next: string) => {
		setSearch(next);
		setSearchClosedSlots(new Set());
	};
	const changeSearchOpenSlots = (next: string[]) =>
		setSearchClosedSlots(current => new Set([...current, ...shownSlots].filter(slot => !next.includes(slot))));

	const problem = validateBulkEnchantPresetName(presets, name, preset?.name);

	const toggle = useCallback(
		(slot: ItemSlot, effectId: number) =>
			setSelected(current => {
				const ids = current.get(slot) ?? [];
				return new Map(current).set(slot, ids.includes(effectId) ? ids.filter(id => id !== effectId) : [...ids, effectId]);
			}),
		[],
	);

	const save = () => {
		saveBulkEnchantPreset(player, { name, slots: [...selected].map(([slot, effectIds]) => ({ slot, effectIds })) }, preset?.name);
		onClose();
	};

	return (
		<BulkEnchantPresetDialogView
			open={open}
			onOpenChange={next => !next && onClose()}
			title={i18n.t(preset ? 'bulk_tab.settings.enchant_presets.edit_title' : 'bulk_tab.settings.enchant_presets.new_title')}
			name={name}
			onNameChange={setName}
			nameTaken={problem === 'duplicate'}
			canSave={problem === null}
			onSave={save}
			openSlots={terms.length ? shownSlots.filter(slot => !searchClosedSlots.has(slot)) : openSlots}
			onOpenSlotsChange={terms.length ? changeSearchOpenSlots : setOpenSlots}
			search={search}
			onSearchChange={changeSearch}
			showAll={showAll}
			onShowAllChange={changeShowAll}
			groups={shown.map(({ slot, enchants }) => ({
				slot,
				label: slotLabel(slot),
				iconUrl: getEmptySlotIconUrl(slot),
				checkedCount: selected.get(slot)?.length ?? 0,
				// One tile per enchant row; rows that share an effect id are one enchant to the sim, so they tick together.
				tiles: enchants.map(enchant => (
					<BulkEnchantTile
						key={`${enchant.effectId}-${enchant.name}`}
						slot={slot}
						enchant={enchant}
						checked={!!selected.get(slot)?.includes(enchant.effectId)}
						onToggle={toggle}
					/>
				)),
			}))}
		/>
	);
};
