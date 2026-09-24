import { Class, ConsumableType, ConsumesSpec, HandType, ItemSlot, Profession, Race, Spec, Stat, WeaponType } from '@generated/proto/common';
import { Consumable } from '@generated/proto/db';
import { ItemEffect } from '@generated/proto/spell';
import { UIItem } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import type { Database } from '@sim/proto/database';
import { iconEnumPickerShown } from '@ui-kit/IconEnumPicker';
import { describe, expect, it } from 'vitest';

import { type ConsumeConfig, consumeConfigs, type ConsumeToggle } from './utils';

const onUse = (categoryId: number, cooldownMs = 0) =>
	ItemEffect.create({ effect: { oneofKind: 'onUse', onUse: { categoryId, cooldownMs, categoryCooldownMs: 0 } } });
const statArray = (stat: Stat, value: number) => Object.assign(new Array<number>(42).fill(0), { [stat]: value });
const strength = (value: number) => statArray(Stat.StatStrength, value);
const elixir = (id: number, name: string, stats: Array<number>, elixirGroupAura: number, elixirGroupMisc: number) =>
	Consumable.create({ id, name, type: ConsumableType.ConsumableTypeElixir, stats, elixirGroupAura, elixirGroupMisc });

const CONSUMABLES = [
	Consumable.create({
		id: 13442,
		name: 'Mighty Rage Potion',
		type: ConsumableType.ConsumableTypePotion,
		categoryId: 4,
		itemEffects: [onUse(4)],
		classAllowlist: [Class.ClassWarrior, Class.ClassDruid],
		stats: strength(60),
	}),
	Consumable.create({ id: 9421, name: 'Major Healthstone', type: ConsumableType.ConsumableTypeConjured, categoryId: 1153, itemEffects: [onUse(1153)] }),
	Consumable.create({ id: 13510, name: 'Flask of the Titans', type: ConsumableType.ConsumableTypeFlask, stats: strength(10) }),
	elixir(9206, 'Elixir of Greater Strength', strength(25), 29, 0),
	elixir(6662, 'Elixir of Giant Growth', strength(8), 29, 0),
	elixir(3383, 'Elixir of Wisdom', statArray(Stat.StatIntellect, 6), 29, 3),
	elixir(12820, 'Winterfall Firewater', statArray(Stat.StatAttackPower, 35), 99, 0),
	Consumable.create({
		id: 238637,
		name: 'Prowler Steak',
		type: ConsumableType.ConsumableTypeFood,
		categoryId: 11,
		itemEffects: [onUse(11)],
		stats: strength(20),
	}),
	Consumable.create({ id: 18588, name: 'Ez-Thro Dynamite II', type: ConsumableType.ConsumableTypeExplosive, categoryId: 24, itemEffects: [onUse(24)] }),
	Consumable.create({
		id: 10646,
		name: 'Goblin Sapper Charge',
		type: ConsumableType.ConsumableTypeExplosive,
		categoryId: 24,
		itemEffects: [onUse(24, 300000)],
		requiredProfession: Profession.Engineering,
	}),
	Consumable.create({ id: 12404, name: 'Dense Sharpening Stone', type: ConsumableType.ConsumableTypeImbue, weaponSubclassMask: 33219 }),
	Consumable.create({ id: 228980, name: 'Conductive Shield Coating', type: ConsumableType.ConsumableTypeImbue, weaponInventoryTypeMask: 16384 }),
	Consumable.create({ id: 2289, name: 'Scroll of Strength II', type: ConsumableType.ConsumableTypeScroll, stats: strength(9) }),
	Consumable.create({ id: 10310, name: 'Scroll of Strength IV', type: ConsumableType.ConsumableTypeScroll, stats: strength(17) }),
	Consumable.create({ id: 4376, name: 'Flame Deflector', type: ConsumableType.ConsumableTypeDevice }),
];

const db = { getConsumables: () => CONSUMABLES, getSpellEffect: () => undefined } as unknown as Database;

interface PlayerState {
	cls: Class;
	spec: Spec;
	professions: Array<Profession>;
	windfury: boolean;
	mainHand?: UIItem;
	offHand?: UIItem;
	consumes: ConsumesSpec;
}

const playerOf = (state: Partial<PlayerState> = {}) => {
	const full: PlayerState = {
		cls: Class.ClassWarrior,
		spec: Spec.SpecDpsWarrior,
		professions: [],
		windfury: false,
		consumes: ConsumesSpec.create(),
		...state,
	};
	const player = {
		getClass: () => full.cls,
		getSpec: () => full.spec,
		getRace: () => Race.RaceOrc,
		hasProfession: (profession: Profession) => full.professions.includes(profession),
		getEquippedItem: (slot: ItemSlot) => {
			const item = slot === ItemSlot.ItemSlotMainHand ? full.mainHand : slot === ItemSlot.ItemSlotOffHand ? full.offHand : undefined;
			return item ? { item } : null;
		},
		getSpecConfig: () => ({ epStats: [Stat.StatStrength, Stat.StatAttackPower] }),
		getParty: () => ({ getBuffs: () => ({ windfuryTotem: full.windfury }) }),
		getConsumes: () => ConsumesSpec.clone(full.consumes),
		setConsumes: (consumes: ConsumesSpec) => (full.consumes = ConsumesSpec.clone(consumes)),
	} as unknown as Player<any>;
	return { player, state: full };
};

const offered = (config: ConsumeConfig, player: Player<any>) =>
	config.values.filter(value => value.value !== 0 && (!value.showWhen || value.showWhen(player))).map(value => value.value);
const toggled = (toggles: ReadonlyArray<ConsumeToggle>, player: Player<any>) =>
	toggles.filter(toggle => !toggle.showWhen || toggle.showWhen(player)).map(toggle => toggle.actionId.itemId);

describe('consumeConfigs', () => {
	it('offers the Mighty Rage Potion to a class the client allows it to', () => {
		const warrior = playerOf().player;
		expect(offered(consumeConfigs(warrior, db).potion, warrior)).toEqual([13442]);
		const paladin = playerOf({ cls: Class.ClassPaladin, spec: Spec.SpecRetributionPaladin }).player;
		expect(offered(consumeConfigs(paladin, db).potion, paladin)).toEqual([]);
	});

	it('puts each consumable in its picker', () => {
		const { player } = playerOf({ professions: [Profession.Engineering] });
		const configs = consumeConfigs(player, db);
		expect(offered(configs.conjured, player)).toEqual([]);
		expect(offered(configs.flask, player)).toEqual([13510]);
		expect(configs.elixirs.map(config => offered(config, player))).toEqual([[9206, 6662], [], [12820]]);
		expect(offered(configs.food, player)).toEqual([238637]);
		expect(offered(configs.explosive, player)).toEqual([18588]);
		expect(toggled(configs.explosiveExtras, player)).toEqual([10646]);
		expect(toggled(configs.scrolls, player)).toEqual([10310]);
		expect(toggled(configs.miscellaneous, player)).toEqual([4376]);
	});

	it('offers the Goblin Sapper Charge to an engineer only', () => {
		const { player } = playerOf();
		expect(toggled(consumeConfigs(player, db).explosiveExtras, player)).toEqual([]);
	});

	it('offers each hand the imbues its item takes, a shield coating on a shield', () => {
		const sword = UIItem.create({ weaponType: WeaponType.WeaponTypeSword, handType: HandType.HandTypeOneHand });
		const shield = UIItem.create({ weaponType: WeaponType.WeaponTypeShield, handType: HandType.HandTypeOffHand });
		const { player } = playerOf({ spec: Spec.SpecProtectionWarrior, mainHand: sword, offHand: shield });
		const configs = consumeConfigs(player, db);
		expect(offered(configs.mhImbue, player)).toEqual([12404]);
		expect(offered(configs.ohImbue, player)).toEqual([228980]);
	});

	it('hides the main-hand imbue under Windfury Totem', () => {
		const { player } = playerOf({ windfury: true });
		expect(consumeConfigs(player, db).mhImbue.showWhen!(player)).toBe(false);
		const feral = playerOf({ cls: Class.ClassDruid, spec: Spec.SpecFeralCatDruid, windfury: true }).player;
		expect(consumeConfigs(feral, db).mhImbue.showWhen!(feral)).toBe(true);
	});

	it('names each elixir group by its aura and hides one that offers the player nothing', () => {
		const { player } = playerOf();
		const configs = consumeConfigs(player, db).elixirs;
		expect(configs.map(config => config.tooltip)).toEqual(['Strength', 'Intellect', 'Attack Power']);
		expect(configs.map(config => iconEnumPickerShown(config, player))).toEqual([true, false, true]);
	});

	it('picks one elixir of each group, and any number of groups', () => {
		const { player, state } = playerOf();
		const [strengthGroup, , attackPowerGroup] = consumeConfigs(player, db).elixirs;
		strengthGroup.setValue(player, 9206);
		attackPowerGroup.setValue(player, 12820);
		expect(state.consumes.elixirIds).toEqual([9206, 12820]);
		strengthGroup.setValue(player, 6662);
		expect(state.consumes.elixirIds).toEqual([12820, 6662]);
		expect(strengthGroup.getValue(player)).toBe(6662);
		strengthGroup.setValue(player, 0);
		expect(state.consumes.elixirIds).toEqual([12820]);
	});

	it('clears the elixirs for a flask and the flask for an elixir', () => {
		const { player, state } = playerOf({ consumes: ConsumesSpec.create({ elixirIds: [9206, 12820] }) });
		const configs = consumeConfigs(player, db);
		configs.flask.setValue(player, 13510);
		expect([state.consumes.flaskId, state.consumes.elixirIds]).toEqual([13510, []]);
		configs.elixirs[0].setValue(player, 9206);
		expect([state.consumes.flaskId, state.consumes.elixirIds]).toEqual([0, [9206]]);
	});

	it('toggles an item in the extra ids', () => {
		const { player, state } = playerOf({ cls: Class.ClassHunter, spec: Spec.SpecHunter });
		const scroll = consumeConfigs(player, db).scrolls.find(toggle => toggle.actionId.itemId === 10310)!;
		scroll.setValue(player, true);
		expect(state.consumes.extraIds).toEqual([10310]);
		scroll.setValue(player, false);
		expect(state.consumes.extraIds).toEqual([]);
	});

	it('offers the highest rank the player can use, below a top rank it cannot', () => {
		const ranks = [
			Consumable.create({ id: 2289, name: 'Scroll of Strength II', type: ConsumableType.ConsumableTypeScroll, stats: strength(9) }),
			Consumable.create({
				id: 10310,
				name: 'Scroll of Strength IV',
				type: ConsumableType.ConsumableTypeScroll,
				stats: strength(17),
				classAllowlist: [Class.ClassRogue],
			}),
			Consumable.create({ id: 12404, name: 'Dense Sharpening Stone II', type: ConsumableType.ConsumableTypeImbue, weaponSubclassMask: 1 << 7 }),
			Consumable.create({ id: 12643, name: 'Dense Sharpening Stone III', type: ConsumableType.ConsumableTypeImbue, weaponSubclassMask: 1 << 15 }),
		];
		const rankDb = { getConsumables: () => ranks, getSpellEffect: () => undefined } as unknown as Database;
		const sword = UIItem.create({ weaponType: WeaponType.WeaponTypeSword, handType: HandType.HandTypeOneHand });
		const dagger = UIItem.create({ weaponType: WeaponType.WeaponTypeDagger, handType: HandType.HandTypeOneHand });

		const warrior = playerOf({ mainHand: sword }).player;
		const configs = consumeConfigs(warrior, rankDb);
		expect(toggled(configs.scrolls, warrior)).toEqual([2289]);
		expect(offered(configs.mhImbue, warrior)).toEqual([12404]);

		const rogue = playerOf({ cls: Class.ClassRogue, mainHand: dagger }).player;
		const rogueConfigs = consumeConfigs(rogue, rankDb);
		expect(toggled(rogueConfigs.scrolls, rogue)).toEqual([10310]);
		expect(offered(rogueConfigs.mhImbue, rogue)).toEqual([12643]);
	});
});
