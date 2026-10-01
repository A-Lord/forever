import { BulkEnchantPreset, BulkEnchantSlotOptions } from '@generated/proto/api';
import { ItemSlot } from '@generated/proto/common';
import { UIEnchant } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import { bulkState, seedBulkSettings } from '@sim/settings/bulk_settings';
import { createSimStore } from '@sim/state/sim_store';
import { beforeEach, describe, expect, it } from 'vitest';

import { bulkEnchantPool, deleteBulkEnchantPreset, saveBulkEnchantPreset, setBulkEnchantPresetChecked, validateBulkEnchantPresetName } from './enchants';

const AGILITY = 10;
const STAMINA = 11;
const STRENGTH = 12;
const CRUSADER = 20;

const USABLE = new Map<ItemSlot, number[]>([
	[ItemSlot.ItemSlotBack, [AGILITY, STAMINA, STRENGTH]],
	[ItemSlot.ItemSlotMainHand, [CRUSADER]],
]);

const preset = (name: string, slots: Array<[ItemSlot, number[]]>) =>
	BulkEnchantPreset.create({ name, slots: slots.map(([slot, effectIds]) => BulkEnchantSlotOptions.create({ slot, effectIds })) });

let player: Player<any>;

beforeEach(() => {
	player = {
		storeKey: 0,
		sim: { store: createSimStore() },
		getEnchants: (slot: ItemSlot) => (USABLE.get(slot) ?? []).map(effectId => UIEnchant.create({ effectId })),
	} as unknown as Player<any>;
	seedBulkSettings(player);
});

const saveChecked = (...presets: BulkEnchantPreset[]) =>
	presets.forEach(saved => {
		saveBulkEnchantPreset(player, saved);
		setBulkEnchantPresetChecked(player, saved.name, true);
	});

describe('bulkEnchantPool', () => {
	it('unions the checked presets per slot, each enchant once', () => {
		saveChecked(preset('Undead', [[ItemSlot.ItemSlotBack, [STAMINA, AGILITY]]]), preset('Beasts', [[ItemSlot.ItemSlotBack, [AGILITY, STRENGTH]]]));

		expect(bulkEnchantPool(player)).toEqual([BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY, STAMINA, STRENGTH] })]);
	});

	it('ignores a preset that is not checked', () => {
		saveChecked(preset('Undead', [[ItemSlot.ItemSlotBack, [AGILITY]]]));
		saveBulkEnchantPreset(player, preset('Beasts', [[ItemSlot.ItemSlotMainHand, [CRUSADER]]]));

		expect(bulkEnchantPool(player)).toEqual([BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY] })]);
	});

	it('drops an enchant the player cannot use in that slot, and a slot left empty by it', () => {
		saveChecked(
			preset('Outdoor', [
				[ItemSlot.ItemSlotBack, [AGILITY, 99]],
				[ItemSlot.ItemSlotHead, [AGILITY]],
			]),
		);

		expect(bulkEnchantPool(player)).toEqual([BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY] })]);
	});
});

describe('enchant preset writers', () => {
	it('keeps a renamed preset checked, under its new name', () => {
		saveChecked(preset('Undead', [[ItemSlot.ItemSlotBack, [AGILITY]]]));

		saveBulkEnchantPreset(player, preset('Scourge', [[ItemSlot.ItemSlotBack, [STAMINA]]]), 'Undead');

		expect(bulkState(player).enchantPresets.map(saved => saved.name)).toEqual(['Scourge']);
		expect(bulkState(player).checkedEnchantPresets).toEqual(['Scourge']);
		expect(bulkEnchantPool(player)).toEqual([BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [STAMINA] })]);
	});

	it('unchecks a deleted preset', () => {
		saveChecked(preset('Undead', [[ItemSlot.ItemSlotBack, [AGILITY]]]));

		deleteBulkEnchantPreset(player, 'Undead');

		expect(bulkState(player).enchantPresets).toEqual([]);
		expect(bulkState(player).checkedEnchantPresets).toEqual([]);
		expect(bulkEnchantPool(player)).toEqual([]);
	});
});

describe('validateBulkEnchantPresetName', () => {
	const presets = [preset('Undead', []), preset('Beasts', [])];

	it.each([
		['an empty name', '', undefined, 'empty'],
		['a blank name', '   ', undefined, 'empty'],
		['a taken name', ' Beasts ', undefined, 'duplicate'],
		['another preset name while editing', 'Beasts', 'Undead', 'duplicate'],
		['the edited preset its own name', 'Undead', 'Undead', null],
		['a new name', 'Outdoor', undefined, null],
	])('answers %s', (_, name, previousName, expected) => {
		expect(validateBulkEnchantPresetName(presets, name, previousName)).toBe(expected);
	});
});
