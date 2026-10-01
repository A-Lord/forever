import { enchantActionId } from '@features/gear/model/item_data';
import { BulkEnchantPreset, BulkEnchantSlotOptions } from '@generated/proto/api';
import { ItemSlot } from '@generated/proto/common';
import type { UIEnchant as Enchant } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { translateBulkSlotName, translateSlotName } from '@i18n/localization';
import { BulkSimItemSlot } from '@sim/bulk/utils';
import { usePlayer } from '@sim/context/SimHostContext';
import { useActionId } from '@sim/hooks/useActionId';
import { getEnchantDescription } from '@sim/proto/enchants';
import { useEffect, useMemo, useState } from 'react';

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
	enchant: Enchant;
	checked: boolean;
	onToggle: () => void;
}

const EnchantTile = ({ enchant, checked, onToggle }: EnchantTileProps) => {
	const actionId = useMemo(() => enchantActionId(enchant), [enchant]);
	const { iconUrl, href } = useActionId(actionId);
	const [effect, setEffect] = useState<string>();

	// The description falls back to the enchant's name, which the tile already shows.
	useEffect(() => {
		let live = true;
		getEnchantDescription(enchant)
			.then(text => {
				if (live) setEffect(text === enchant.name ? undefined : text);
			})
			.catch(() => {});
		return () => {
			live = false;
		};
	}, [enchant]);

	return <BulkEnchantTile name={enchant.name} effect={effect} iconUrl={iconUrl} href={href} checked={checked} onToggle={onToggle} />;
};

const slotLabel = (slot: ItemSlot) => (slot === ItemSlot.ItemSlotFinger1 ? translateBulkSlotName(BulkSimItemSlot.ItemSlotFinger) : translateSlotName(slot));

const uniqueByEffectId = (enchants: Enchant[]) =>
	enchants.filter((enchant, index) => enchants.findIndex(other => other.effectId === enchant.effectId) === index);

export const BulkEnchantPresetDialog = ({ open, preset, onClose }: BulkEnchantPresetDialogProps) => {
	const player = usePlayer();
	const presets = useBulkState(slice => slice.enchantPresets);
	const [name, setName] = useState(preset?.name ?? '');
	const [selected, setSelected] = useState<ReadonlyMap<ItemSlot, readonly number[]>>(
		() => new Map(preset?.slots.map(options => [options.slot, options.effectIds]) ?? []),
	);

	const slots = useMemo(
		() => BULK_ENCHANT_SLOTS.map(slot => ({ slot, enchants: uniqueByEffectId(player.getEnchants(slot)) })).filter(({ enchants }) => enchants.length),
		[player],
	);

	const problem = validateBulkEnchantPresetName(presets, name, preset?.name);

	const toggle = (slot: ItemSlot, effectId: number) => {
		const current = selected.get(slot) ?? [];
		const next = current.includes(effectId) ? current.filter(id => id !== effectId) : [...current, effectId];
		setSelected(new Map(selected).set(slot, next));
	};

	const save = () => {
		const presetSlots = BULK_ENCHANT_SLOTS.flatMap(slot => {
			const effectIds = selected.get(slot);
			return effectIds?.length ? [BulkEnchantSlotOptions.create({ slot, effectIds: [...effectIds] })] : [];
		});
		saveBulkEnchantPreset(player, BulkEnchantPreset.create({ name, slots: presetSlots }), preset?.name);
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
				label: slotLabel(slot),
				tiles: enchants.map(enchant => (
					<EnchantTile
						key={enchant.effectId}
						enchant={enchant}
						checked={!!selected.get(slot)?.includes(enchant.effectId)}
						onToggle={() => toggle(slot, enchant.effectId)}
					/>
				)),
			}))}
		/>
	);
};
