import { Class, ConsumableType, ConsumesSpec, HandType, ItemSlot, Profession, Race, Spec, Stat, TristateEffect, WeaponType } from '@generated/proto/common';
import { Consumable } from '@generated/proto/db';
import { ItemEffect } from '@generated/proto/spell';
import { UIItem } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import type { Database } from '@sim/proto/database';
import { describe, expect, it } from 'vitest';

import { type ConsumeConfig, consumeConfigs, type ConsumeToggle } from './utils';

const onUse = (categoryId: number, cooldownMs = 0) =>
	ItemEffect.create({ effect: { oneofKind: 'onUse', onUse: { categoryId, cooldownMs, categoryCooldownMs: 0 } } });
const strength = (value: number) => Object.assign(new Array<number>(42).fill(0), { [Stat.StatStrength]: value });

const CONSUMABLES = [
	Consumable.create({
		id: 13442,
		name: 'Mighty Rage Potion',
		type: ConsumableType.ConsumableTypePotion,
		itemEffects: [onUse(4)],
		classAllowlist: [Class.ClassWarrior, Class.ClassDruid],
		stats: strength(60),
	}),
	Consumable.create({ id: 9421, name: 'Major Healthstone', type: ConsumableType.ConsumableTypeOther, itemEffects: [onUse(1153)] }),
	Consumable.create({ id: 13510, name: 'Flask of the Titans', type: ConsumableType.ConsumableTypeFlask, stats: strength(10) }),
	Consumable.create({ id: 9206, name: 'Elixir of Greater Strength', type: ConsumableType.ConsumableTypeElixir, stats: strength(20) }),
	Consumable.create({ id: 6662, name: 'Elixir of Giant Growth', type: ConsumableType.ConsumableTypeElixir, stats: strength(8) }),
	Consumable.create({ id: 238637, name: 'Prowler Steak', type: ConsumableType.ConsumableTypeOther, itemEffects: [onUse(11)], stats: strength(20) }),
	Consumable.create({ id: 18588, name: 'Ez-Thro Dynamite II', type: ConsumableType.ConsumableTypeExplosive, itemEffects: [onUse(24)] }),
	Consumable.create({
		id: 10646,
		name: 'Goblin Sapper Charge',
		type: ConsumableType.ConsumableTypeExplosive,
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
	windfury: TristateEffect;
	mainHand?: UIItem;
	offHand?: UIItem;
	consumes: ConsumesSpec;
}

const playerOf = (state: Partial<PlayerState> = {}) => {
	const full: PlayerState = {
		cls: Class.ClassWarrior,
		spec: Spec.SpecDpsWarrior,
		professions: [],
		windfury: TristateEffect.TristateEffectMissing,
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
		expect(offered(configs.battleElixir, player)).toEqual([9206, 6662]);
		expect(offered(configs.guardianElixir, player)).toEqual([9206, 6662]);
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
		const { player } = playerOf({ windfury: TristateEffect.TristateEffectRegular });
		expect(consumeConfigs(player, db).mhImbue.showWhen!(player)).toBe(false);
		const feral = playerOf({ cls: Class.ClassDruid, spec: Spec.SpecFeralCatDruid, windfury: TristateEffect.TristateEffectRegular }).player;
		expect(consumeConfigs(feral, db).mhImbue.showWhen!(feral)).toBe(true);
	});

	it('keeps one elixir out of the other elixir picker', () => {
		const { player } = playerOf({ consumes: ConsumesSpec.create({ guardianElixirId: 6662 }) });
		expect(offered(consumeConfigs(player, db).battleElixir, player)).toEqual([9206]);
	});

	it('clears the elixirs for a flask and the flask for an elixir', () => {
		const { player, state } = playerOf({ consumes: ConsumesSpec.create({ battleElixirId: 9206, guardianElixirId: 6662 }) });
		const configs = consumeConfigs(player, db);
		configs.flask.setValue(player, 13510);
		expect([state.consumes.flaskId, state.consumes.battleElixirId, state.consumes.guardianElixirId]).toEqual([13510, 0, 0]);
		configs.battleElixir.setValue(player, 9206);
		expect([state.consumes.flaskId, state.consumes.battleElixirId]).toEqual([0, 9206]);
	});

	it('toggles an item in the extra ids, and a pet scroll in the pet’s', () => {
		const { player, state } = playerOf({ cls: Class.ClassHunter, spec: Spec.SpecHunter });
		const configs = consumeConfigs(player, db);
		configs.scrolls[0].setValue(player, true);
		configs.petScrolls[0].setValue(player, true);
		expect([state.consumes.extraIds, state.consumes.petExtraIds]).toEqual([[10310], [10310]]);
		configs.scrolls[0].setValue(player, false);
		expect(state.consumes.extraIds).toEqual([]);
		expect(toggled(configs.petScrolls, player)).toEqual([10310]);
		const warrior = playerOf().player;
		expect(toggled(consumeConfigs(warrior, db).petScrolls, warrior)).toEqual([]);
	});
});
