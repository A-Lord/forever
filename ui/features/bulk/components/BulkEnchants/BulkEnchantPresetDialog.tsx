import { useEnchantDescription } from '@features/gear/hooks/useEnchantDescription';
import { enchantActionId } from '@features/gear/model/item_data';
import type { ItemSlot } from '@generated/proto/common';
import type { BulkEnchantPreset, UIEnchant as Enchant } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { translateBulkSlotName } from '@i18n/localization';
import { ITEM_SLOT_TO_BULK_SIM_ITEM_SLOT } from '@sim/bulk/utils';
import { usePlayer } from '@sim/context/SimHostContext';
import { useActionId } from '@sim/hooks/useActionId';
import { useActionIdWowheadDataset } from '@sim/hooks/useActionIdWowheadDataset';
import { memo, useCallback, useMemo, useState } from 'react';

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

interface EnchantTileProps {
	slot: ItemSlot;
	enchant: Enchant;
	checked: boolean;
	onToggle: (slot: ItemSlot, effectId: number) => void;
}

const EnchantTile = memo(({ slot, enchant, checked, onToggle }: EnchantTileProps) => {
	const actionId = useMemo(() => enchantActionId(enchant), [enchant]);
	const { iconUrl, href } = useActionId(actionId);
	const wowhead = useActionIdWowheadDataset(actionId);
	const description = useEnchantDescription(enchant);
	const toggle = useCallback(() => onToggle(slot, enchant.effectId), [onToggle, slot, enchant]);

	return (
		<BulkEnchantTile
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

export const BulkEnchantPresetDialog = ({ open, preset, onClose }: BulkEnchantPresetDialogProps) => {
	const player = usePlayer();
	const presets = useBulkState(slice => slice.enchantPresets);
	const [name, setName] = useState(preset?.name ?? '');
	const [selected, setSelected] = useState<ReadonlyMap<ItemSlot, readonly number[]>>(
		() => new Map(preset?.slots.map(options => [options.slot, options.effectIds]) ?? []),
	);

	// One tile per enchant row; rows that share an effect id are one enchant to the sim, so they tick together.
	const slots = useMemo(
		() => BULK_ENCHANT_SLOTS.map(slot => ({ slot, enchants: player.getEnchants(slot) })).filter(({ enchants }) => enchants.length),
		[player],
	);

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
			groups={slots.map(({ slot, enchants }) => ({
				slot,
				label: translateBulkSlotName(ITEM_SLOT_TO_BULK_SIM_ITEM_SLOT.get(slot)!),
				tiles: enchants.map(enchant => (
					<EnchantTile
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
