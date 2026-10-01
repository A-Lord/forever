import { BulkEnchantPreset, BulkEnchantSlotOptions } from '@generated/proto/api';
import { ItemSlot } from '@generated/proto/common';
import type { Player } from '@sim/player/player';
import { bulkState, patchBulkState } from '@sim/settings/bulk_settings';

export type BulkEnchantPresetNameProblem = 'empty' | 'duplicate';

// Finger1 stands for both fingers.
export const BULK_ENCHANT_SLOTS: readonly ItemSlot[] = [
	ItemSlot.ItemSlotHead,
	ItemSlot.ItemSlotShoulder,
	ItemSlot.ItemSlotBack,
	ItemSlot.ItemSlotChest,
	ItemSlot.ItemSlotWrist,
	ItemSlot.ItemSlotHands,
	ItemSlot.ItemSlotLegs,
	ItemSlot.ItemSlotFeet,
	ItemSlot.ItemSlotFinger1,
	ItemSlot.ItemSlotMainHand,
	ItemSlot.ItemSlotOffHand,
	ItemSlot.ItemSlotRanged,
];

// Sorted, so the request that seeds the batch's RNG does not depend on preset order or the order enchants were picked in.
export const bulkEnchantPool = (player: Player<any>): BulkEnchantSlotOptions[] => {
	const { enchantPresets, checkedEnchantPresets } = bulkState(player);
	const checked = enchantPresets.filter(preset => checkedEnchantPresets.includes(preset.name));
	if (!checked.length) return [];

	return BULK_ENCHANT_SLOTS.flatMap(slot => {
		const wanted = new Set(checked.flatMap(preset => preset.slots.filter(options => options.slot === slot).flatMap(options => options.effectIds)));
		if (!wanted.size) return [];
		const usable = new Set(player.getEnchants(slot).map(enchant => enchant.effectId));
		const effectIds = [...wanted].filter(id => usable.has(id)).sort((a, b) => a - b);
		return effectIds.length ? [BulkEnchantSlotOptions.create({ slot, effectIds })] : [];
	});
};

export const validateBulkEnchantPresetName = (
	presets: readonly BulkEnchantPreset[],
	name: string,
	previousName?: string,
): BulkEnchantPresetNameProblem | null => {
	const trimmed = name.trim();
	if (!trimmed) return 'empty';
	if (presets.some(preset => preset.name === trimmed && preset.name !== previousName)) return 'duplicate';
	return null;
};

export const saveBulkEnchantPreset = (player: Player<any>, preset: BulkEnchantPreset, previousName?: string) => {
	const { enchantPresets, checkedEnchantPresets } = bulkState(player);
	const saved = BulkEnchantPreset.create({
		name: preset.name.trim(),
		slots: preset.slots.filter(options => options.effectIds.length).map(options => BulkEnchantSlotOptions.clone(options)),
	});
	const replaced = previousName !== undefined && enchantPresets.some(existing => existing.name === previousName);

	patchBulkState(
		player,
		{
			enchantPresets: replaced ? enchantPresets.map(existing => (existing.name === previousName ? saved : existing)) : [...enchantPresets, saved],
			checkedEnchantPresets: replaced ? checkedEnchantPresets.map(name => (name === previousName ? saved.name : name)) : checkedEnchantPresets,
		},
		['settings'],
	);
};

export const deleteBulkEnchantPreset = (player: Player<any>, name: string) => {
	const { enchantPresets, checkedEnchantPresets } = bulkState(player);
	patchBulkState(
		player,
		{
			enchantPresets: enchantPresets.filter(preset => preset.name !== name),
			checkedEnchantPresets: checkedEnchantPresets.filter(checkedName => checkedName !== name),
		},
		['settings'],
	);
};

export const setBulkEnchantPresetChecked = (player: Player<any>, name: string, checked: boolean) => {
	const { checkedEnchantPresets } = bulkState(player);
	if (checkedEnchantPresets.includes(name) === checked) return;

	patchBulkState(
		player,
		{ checkedEnchantPresets: checked ? [...checkedEnchantPresets, name] : checkedEnchantPresets.filter(checkedName => checkedName !== name) },
		['settings'],
	);
};
