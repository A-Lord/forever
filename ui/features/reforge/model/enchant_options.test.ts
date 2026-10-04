import { BulkEnchantSlotOptions } from '@generated/proto/api';
import { EnchantProcKind, HandType, ItemSlot, ItemType, PseudoStat, RangedWeaponType, Stat, WeaponType } from '@generated/proto/common';
import type { UIEnchant, UIItem } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import { Gear } from '@sim/proto/gear';
import { Stats } from '@sim/proto/stats';
import { makeEquippedItem } from '@testing/storybook/fixtures/gear';
import { makeUIEnchant } from '@testing/storybook/fixtures/protos';
import { describe, expect, it } from 'vitest';

import { buildEnchantOptions } from './enchant_options';

const WEIGHTS = new Stats().withStat(Stat.StatStrength, 2).withStat(Stat.StatAgility, 1).withPseudoStat(PseudoStat.PseudoStatRangedDps, 3);

const stats = (stat: Stat, value: number) => new Stats().withStat(stat, value).asProtoArray();

const armor = (type: ItemType): Partial<UIItem> => ({ type, weaponType: WeaponType.WeaponTypeUnknown, handType: HandType.HandTypeUnknown, weaponSpeed: 0 });

const BOW: Partial<UIItem> = { ...armor(ItemType.ItemTypeRanged), rangedWeaponType: RangedWeaponType.RangedWeaponTypeBow, weaponSpeed: 3 };
const WAND: Partial<UIItem> = { ...BOW, rangedWeaponType: RangedWeaponType.RangedWeaponTypeWand, weaponSpeed: 1.5 };

const enchant = (effectId: number, name: string, type: ItemType, overrides: Partial<UIEnchant> = {}) => makeUIEnchant({ effectId, name, type, ...overrides });

const WRIST_STRENGTH = enchant(1885, 'Enchant Bracer - Superior Strength', ItemType.ItemTypeWrist, { stats: stats(Stat.StatStrength, 9) });
const WRIST_AGILITY = enchant(247, 'Enchant Bracer - Minor Agility', ItemType.ItemTypeWrist, { stats: stats(Stat.StatAgility, 1) });
const WRIST_SPIRIT = enchant(1883, 'Enchant Bracer - Superior Spirit', ItemType.ItemTypeWrist, { stats: stats(Stat.StatSpirit, 9) });
const CRUSADER = makeUIEnchant({ procKind: EnchantProcKind.EnchantProcKindCombat });
const WEAPON_AGILITY = enchant(2564, 'Enchant Weapon - Agility', ItemType.ItemTypeWeapon, { stats: stats(Stat.StatAgility, 15) });
const CHEST_ABSORPTION = enchant(63, 'Enchant Chest - Lesser Absorption', ItemType.ItemTypeChest, { procKind: EnchantProcKind.EnchantProcKindSurvival });
const CHEST_STATS = enchant(1891, 'Enchant Chest - Greater Stats', ItemType.ItemTypeChest, {
	stats: new Stats().withStat(Stat.StatStrength, 4).withStat(Stat.StatAgility, 4).asProtoArray(),
});
const BACK_THREAT = enchant(2613, 'Enchant Cloak - Subtlety', ItemType.ItemTypeBack);
const BACK_AGILITY = enchant(849, 'Enchant Cloak - Lesser Agility', ItemType.ItemTypeBack, { stats: stats(Stat.StatAgility, 3) });
const SCOPE = enchant(664, 'Sniper Scope', ItemType.ItemTypeRanged, { weaponDamage: 7 });

const ALL_ENCHANTS = [WRIST_STRENGTH, WRIST_AGILITY, WRIST_SPIRIT, CRUSADER, WEAPON_AGILITY, CHEST_ABSORPTION, CHEST_STATS, BACK_THREAT, BACK_AGILITY, SCOPE];

// The rule reads two members of the player: its enchant list and whether its spec tanks.
const makePlayer = ({ isTankSpec = false, enchants = ALL_ENCHANTS }: { isTankSpec?: boolean; enchants?: UIEnchant[] } = {}) =>
	({ playerSpec: { isTankSpec }, getEnchants: () => enchants }) as unknown as Player<any>;

const makeGear = (items: Partial<Record<ItemSlot, Partial<UIItem>>>) =>
	new Gear(Object.fromEntries(Object.entries(items).map(([slot, item]) => [slot, makeEquippedItem({ item })])));

describe('buildEnchantOptions', () => {
	it('lists the fitting static enchants with EP above 0, ids sorted and deduplicated', () => {
		const duplicate = makeUIEnchant({ ...WRIST_STRENGTH, name: 'Enchant Bracer - Superior Strength (second row)' });
		const player = makePlayer({ enchants: [WRIST_STRENGTH, WRIST_SPIRIT, WRIST_AGILITY, duplicate] });

		const result = buildEnchantOptions(player, makeGear({ [ItemSlot.ItemSlotWrist]: armor(ItemType.ItemTypeWrist) }), [], WEIGHTS);

		expect(result.options).toEqual([BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotWrist, effectIds: [247, 1885] })]);
		expect(result.skipped).toEqual([]);
	});

	it('skips a weapon that can take a combat proc and names the proc', () => {
		const result = buildEnchantOptions(makePlayer(), makeGear({ [ItemSlot.ItemSlotMainHand]: {} }), [], WEIGHTS);

		expect(result.options).toEqual([]);
		expect(result.skipped).toEqual([{ slot: ItemSlot.ItemSlotMainHand, blockingEnchant: CRUSADER }]);
	});

	it('skips a slot that can take a survival proc for a tank spec only', () => {
		const gear = makeGear({ [ItemSlot.ItemSlotChest]: armor(ItemType.ItemTypeChest) });

		const tank = buildEnchantOptions(makePlayer({ isTankSpec: true }), gear, [], WEIGHTS);
		const dps = buildEnchantOptions(makePlayer(), gear, [], WEIGHTS);

		expect(tank.options).toEqual([]);
		expect(tank.skipped).toEqual([{ slot: ItemSlot.ItemSlotChest, blockingEnchant: CHEST_ABSORPTION }]);
		expect(dps.options).toEqual([BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotChest, effectIds: [1891] })]);
		expect(dps.skipped).toEqual([]);
	});

	it('leaves frozen and empty slots out of both lists', () => {
		const gear = makeGear({ [ItemSlot.ItemSlotWrist]: armor(ItemType.ItemTypeWrist), [ItemSlot.ItemSlotMainHand]: {} });

		const result = buildEnchantOptions(makePlayer(), gear, [ItemSlot.ItemSlotWrist, ItemSlot.ItemSlotMainHand], WEIGHTS);

		expect(result).toEqual({ options: [], skipped: [] });
	});

	it('neither blocks on nor picks an enchant with a permanent effect and no stats', () => {
		const gear = makeGear({ [ItemSlot.ItemSlotBack]: armor(ItemType.ItemTypeBack) });

		const withAgility = buildEnchantOptions(makePlayer(), gear, [], WEIGHTS);
		const threatOnly = buildEnchantOptions(makePlayer({ enchants: [BACK_THREAT] }), gear, [], WEIGHTS);

		expect(withAgility).toEqual({ options: [BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [849] })], skipped: [] });
		expect(threatOnly).toEqual({ options: [], skipped: [] });
	});

	it('values a scope on a bow through ranged DPS, and finds nothing for a wand', () => {
		const bow = makeGear({ [ItemSlot.ItemSlotRanged]: BOW });

		const scoped = buildEnchantOptions(makePlayer(), bow, [], WEIGHTS);
		const noRangedWeight = buildEnchantOptions(makePlayer(), bow, [], WEIGHTS.withPseudoStat(PseudoStat.PseudoStatRangedDps, 0));
		const wand = buildEnchantOptions(makePlayer(), makeGear({ [ItemSlot.ItemSlotRanged]: WAND }), [], WEIGHTS);

		expect(scoped).toEqual({ options: [BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotRanged, effectIds: [664] })], skipped: [] });
		expect(noRangedWeight).toEqual({ options: [], skipped: [] });
		expect(wand).toEqual({ options: [], skipped: [] });
	});
});
