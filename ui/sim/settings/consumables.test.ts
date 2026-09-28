import { Class, ConsumableType, HandType, ItemSlot, Profession, PseudoStat, Race, Stat, WeaponType } from '@generated/proto/common';
import { Consumable } from '@generated/proto/db';
import { ItemEffect, ResourceType, SpellEffect } from '@generated/proto/spell';
import { UIItem } from '@generated/proto/ui';
import { describe, expect, it } from 'vitest';

import type { EquippedItem } from '../proto/equipped_item';
import { UnitStat } from '../proto/stats';
import {
	ConsumableSlot,
	consumableSlot,
	type ConsumableUser,
	explosiveUsedBeside,
	handTakesImbue,
	higherRanks,
	highestRanks,
	isConsumableRelevant,
	relevanceFor,
	relevantConsumableIds,
} from './consumables';

const stats = (granted: Partial<Record<Stat, number>>) => {
	const out = new Array<number>(42).fill(0);
	for (const [stat, value] of Object.entries(granted)) out[Number(stat)] = value!;
	return out;
};

const pseudoStats = (granted: Partial<Record<PseudoStat, number>>) => {
	const out = new Array<number>(36).fill(0);
	for (const [pseudoStat, value] of Object.entries(granted)) out[Number(pseudoStat)] = value!;
	return out;
};

const onUse = (categoryId: number, cooldownMs = 0) =>
	ItemEffect.create({ effect: { oneofKind: 'onUse', onUse: { categoryId, cooldownMs, categoryCooldownMs: 0 } } });

const consumable = (fields: Partial<Consumable>) => Consumable.create({ id: 1, name: 'Test', ...fields });

const effects = new Map<number, SpellEffect>([
	[1, SpellEffect.create({ id: 1, miscValue0: { oneofKind: 'resourceType', resourceType: ResourceType.ResourceTypeMana } })],
	[2, SpellEffect.create({ id: 2, miscValue0: { oneofKind: 'resourceType', resourceType: ResourceType.ResourceTypeHealth } })],
	[3, SpellEffect.create({ id: 3, miscValue0: { oneofKind: 'resourceType', resourceType: ResourceType.ResourceTypeRage } })],
]);
const db = { getSpellEffect: (id: number) => effects.get(id) };

const user = (
	fields: Partial<{
		cls: Class;
		race: Race;
		professions: Profession[];
		epStats: Stat[];
		epPseudoStats: PseudoStat[];
		displayPseudoStats: PseudoStat[];
		mainHand: UIItem;
		offHand: UIItem;
	}> = {},
): ConsumableUser => ({
	getClass: () => fields.cls ?? Class.ClassWarrior,
	getRace: () => fields.race ?? Race.RaceOrc,
	hasProfession: profession => (fields.professions ?? []).includes(profession),
	getEquippedItem: slot => {
		const item = slot === ItemSlot.ItemSlotMainHand ? fields.mainHand : slot === ItemSlot.ItemSlotOffHand ? fields.offHand : undefined;
		return item ? ({ item } as EquippedItem) : null;
	},
	getSpecConfig: () => ({
		epStats: fields.epStats ?? [Stat.StatStrength, Stat.StatAttackPower],
		epPseudoStats: fields.epPseudoStats,
		displayStats: (fields.displayPseudoStats ?? []).map(pseudoStat => UnitStat.fromPseudoStat(pseudoStat)),
	}),
});

describe('consumableSlot', () => {
	it('puts each consumable in the picker of its type, and what no picker takes among the extras', () => {
		expect(consumableSlot(consumable({ type: ConsumableType.ConsumableTypeConjured }))).toBe(ConsumableSlot.Conjured);
		expect(consumableSlot(consumable({ type: ConsumableType.ConsumableTypePotion }))).toBe(ConsumableSlot.Potion);
		expect(consumableSlot(consumable({ type: ConsumableType.ConsumableTypeFlask }))).toBe(ConsumableSlot.Flask);
		expect(consumableSlot(consumable({ type: ConsumableType.ConsumableTypeElixir }))).toBe(ConsumableSlot.Elixir);
		expect(consumableSlot(consumable({ type: ConsumableType.ConsumableTypeFood }))).toBe(ConsumableSlot.Food);
		expect(consumableSlot(consumable({ type: ConsumableType.ConsumableTypeExplosive }))).toBe(ConsumableSlot.Explosive);
		expect(consumableSlot(consumable({ type: ConsumableType.ConsumableTypeImbue }))).toBe(ConsumableSlot.Imbue);
		expect(consumableSlot(consumable({ type: ConsumableType.ConsumableTypeScroll }))).toBe(ConsumableSlot.Extra);
		expect(consumableSlot(consumable({ type: ConsumableType.ConsumableTypeDevice }))).toBe(ConsumableSlot.Extra);
	});
});

describe('isConsumableRelevant', () => {
	it('keeps a consumable off a class, a profession or a race the game files do not allow it', () => {
		const mightyRage = consumable({ classAllowlist: [Class.ClassWarrior, Class.ClassDruid], stats: stats({ [Stat.StatStrength]: 60 }) });
		expect(isConsumableRelevant(mightyRage, user(), db)).toBe(true);
		expect(isConsumableRelevant(mightyRage, user({ cls: Class.ClassRogue }), db)).toBe(false);

		const sapper = consumable({ type: ConsumableType.ConsumableTypeExplosive, requiredProfession: Profession.Engineering });
		expect(isConsumableRelevant(sapper, user(), db)).toBe(false);
		expect(isConsumableRelevant(sapper, user({ professions: [Profession.Engineering] }), db)).toBe(true);

		const orcOnly = consumable({ type: ConsumableType.ConsumableTypeExplosive, raceAllowlist: [Race.RaceOrc] });
		expect(isConsumableRelevant(orcOnly, user(), db)).toBe(true);
		expect(isConsumableRelevant(orcOnly, user({ race: Race.RaceHuman }), db)).toBe(false);
	});

	it('offers a stat consumable by a stat the spec weighs, Stamina bringing the resistances', () => {
		const giants = consumable({ type: ConsumableType.ConsumableTypeElixir, stats: stats({ [Stat.StatStrength]: 25 }) });
		const arthas = consumable({ type: ConsumableType.ConsumableTypeElixir, stats: stats({ [Stat.StatShadowResistance]: 10 }) });
		expect(isConsumableRelevant(giants, user(), db)).toBe(true);
		expect(isConsumableRelevant(giants, user({ epStats: [Stat.StatIntellect] }), db)).toBe(false);
		expect(isConsumableRelevant(arthas, user(), db)).toBe(false);
		expect(isConsumableRelevant(arthas, user({ epStats: [Stat.StatStamina] }), db)).toBe(true);
	});

	it('offers flat physical damage to a spec that weighs Attack Power', () => {
		const frenzy = consumable({ type: ConsumableType.ConsumableTypePotion, stats: stats({ [Stat.StatPhysicalDamage]: 40 }) });
		expect(isConsumableRelevant(frenzy, user({ epStats: [Stat.StatAttackPower] }), db)).toBe(true);
		expect(isConsumableRelevant(frenzy, user({ epStats: [Stat.StatIntellect] }), db)).toBe(false);
	});

	it('offers a one-channel percent to a spec that shows or weighs it, not by the rating the spec weighs', () => {
		// Dark Desire 22237: A_MOD_HIT_CHANCE 2, melee and ranged hit alone.
		const darkDesire = consumable({
			type: ConsumableType.ConsumableTypeOther,
			pseudoStats: pseudoStats({ [PseudoStat.PseudoStatMeleeHitPercent]: 2, [PseudoStat.PseudoStatRangedHitPercent]: 2 }),
		});
		const caster = { epStats: [Stat.StatSpellDamage, Stat.StatHitRating], displayPseudoStats: [PseudoStat.PseudoStatSpellHitPercent] };
		expect(isConsumableRelevant(darkDesire, user(caster), db)).toBe(false);
		expect(isConsumableRelevant(darkDesire, user({ ...caster, displayPseudoStats: [PseudoStat.PseudoStatMeleeHitPercent] }), db)).toBe(true);
		expect(isConsumableRelevant(darkDesire, user({ ...caster, epPseudoStats: [PseudoStat.PseudoStatRangedHitPercent] }), db)).toBe(true);
	});

	it('offers an all-channel crit percent to a melee spec and to a caster that shows spell crit', () => {
		// Elixir of the Mongoose 13452: A_MOD_CRIT_PCT 2, melee, ranged and spell crit alike.
		const mongoose = consumable({
			type: ConsumableType.ConsumableTypeElixir,
			pseudoStats: pseudoStats({
				[PseudoStat.PseudoStatMeleeCritPercent]: 2,
				[PseudoStat.PseudoStatRangedCritPercent]: 2,
				[PseudoStat.PseudoStatSpellCritPercent]: 2,
			}),
		});
		const rogue = user({ cls: Class.ClassRogue, epStats: [Stat.StatAgility], displayPseudoStats: [PseudoStat.PseudoStatMeleeCritPercent] });
		const mage = user({ cls: Class.ClassMage, epStats: [Stat.StatIntellect], displayPseudoStats: [PseudoStat.PseudoStatSpellCritPercent] });
		expect(isConsumableRelevant(mongoose, rogue, db)).toBe(true);
		expect(isConsumableRelevant(mongoose, mage, db)).toBe(true);
	});

	it('offers a restore by the resource the spec spends', () => {
		const mana = consumable({ type: ConsumableType.ConsumableTypePotion, effectIds: [1] });
		const health = consumable({ type: ConsumableType.ConsumableTypePotion, effectIds: [2] });
		const rage = consumable({ type: ConsumableType.ConsumableTypePotion, effectIds: [3] });
		expect(isConsumableRelevant(mana, user(), db)).toBe(false);
		expect(isConsumableRelevant(mana, user({ epStats: [Stat.StatIntellect] }), db)).toBe(true);
		expect(isConsumableRelevant(health, user(), db)).toBe(false);
		expect(isConsumableRelevant(health, user({ epStats: [Stat.StatStamina] }), db)).toBe(true);
		expect(isConsumableRelevant(rage, user(), db)).toBe(true);
		expect(isConsumableRelevant(rage, user({ epStats: [Stat.StatIntellect] }), db)).toBe(false);
	});

	it('offers a consumable of another kind to everyone, but not a stat consumable that grants nothing', () => {
		expect(isConsumableRelevant(consumable({ type: ConsumableType.ConsumableTypeDevice }), user(), db)).toBe(true);
		expect(isConsumableRelevant(consumable({ type: ConsumableType.ConsumableTypeExplosive }), user(), db)).toBe(true);
		expect(isConsumableRelevant(consumable({ type: ConsumableType.ConsumableTypeOther }), user(), db)).toBe(false);
		expect(isConsumableRelevant(consumable({ type: ConsumableType.ConsumableTypeElixir }), user(), db)).toBe(false);
	});
});

describe('handTakesImbue', () => {
	const item = (weaponType: WeaponType, handType = HandType.HandTypeOneHand) => UIItem.create({ weaponType, handType });
	// Dense Sharpening Stone 12404: SpellEquippedItems subclass mask 33219.
	const sharpeningStone = consumable({ type: ConsumableType.ConsumableTypeImbue, weaponSubclassMask: 33219 });
	// Dense Weightstone 12643: 9264.
	const weightstone = consumable({ type: ConsumableType.ConsumableTypeImbue, weaponSubclassMask: 9264 });
	// Brilliant Wizard Oil 20749: inventory type mask 6430720.
	const oil = consumable({ type: ConsumableType.ConsumableTypeImbue, weaponInventoryTypeMask: 6430720 });
	// Conductive Shield Coating 228980: inventory type mask 16384.
	const coating = consumable({ type: ConsumableType.ConsumableTypeImbue, weaponInventoryTypeMask: 16384 });

	it('takes a stone by the weapon subclass', () => {
		expect(handTakesImbue(sharpeningStone, item(WeaponType.WeaponTypeSword))).toBe(true);
		expect(handTakesImbue(sharpeningStone, item(WeaponType.WeaponTypeAxe, HandType.HandTypeTwoHand))).toBe(true);
		expect(handTakesImbue(sharpeningStone, item(WeaponType.WeaponTypeMace))).toBe(false);
		expect(handTakesImbue(weightstone, item(WeaponType.WeaponTypeMace, HandType.HandTypeTwoHand))).toBe(true);
		expect(handTakesImbue(weightstone, item(WeaponType.WeaponTypeDagger))).toBe(false);
		expect(handTakesImbue(sharpeningStone, item(WeaponType.WeaponTypeShield))).toBe(false);
	});

	it('takes an oil or a shield coating by the inventory type', () => {
		expect(handTakesImbue(oil, item(WeaponType.WeaponTypeDagger, HandType.HandTypeMainHand))).toBe(true);
		expect(handTakesImbue(oil, item(WeaponType.WeaponTypeStaff, HandType.HandTypeTwoHand))).toBe(true);
		expect(handTakesImbue(oil, item(WeaponType.WeaponTypeShield, HandType.HandTypeOffHand))).toBe(false);
		expect(handTakesImbue(coating, item(WeaponType.WeaponTypeShield, HandType.HandTypeOffHand))).toBe(true);
		expect(handTakesImbue(coating, item(WeaponType.WeaponTypeSword, HandType.HandTypeOffHand))).toBe(false);
	});

	it('takes nothing on an empty hand', () => {
		expect(handTakesImbue(oil, undefined)).toBe(false);
	});
});

describe('relevanceFor', () => {
	it('answers as isConsumableRelevant does, and again once the class, race or a required profession changes', () => {
		const sapper = consumable({ type: ConsumableType.ConsumableTypeExplosive, requiredProfession: Profession.Engineering });
		const orcOnly = consumable({ id: 2, type: ConsumableType.ConsumableTypeExplosive, raceAllowlist: [Race.RaceOrc] });
		const relevant = relevanceFor([sapper, orcOnly], db, new Set([Stat.StatStrength]), new Set());
		expect(relevant(sapper, user())).toBe(false);
		expect(relevant(sapper, user({ professions: [Profession.Engineering] }))).toBe(true);
		expect(relevant(orcOnly, user({ professions: [Profession.Engineering] }))).toBe(true);
		expect(relevant(orcOnly, user({ race: Race.RaceHuman, professions: [Profession.Engineering] }))).toBe(false);
	});
});

describe('highestRanks', () => {
	it('keeps the highest numeral of a name, and every consumable whose name has none', () => {
		const ranks = [
			consumable({ id: 3012, name: 'Scroll of Agility', type: ConsumableType.ConsumableTypeScroll }),
			consumable({ id: 1477, name: 'Scroll of Agility II', type: ConsumableType.ConsumableTypeScroll }),
			consumable({ id: 10309, name: 'Scroll of Agility IV', type: ConsumableType.ConsumableTypeScroll }),
			consumable({ id: 4425, name: 'Scroll of Agility III', type: ConsumableType.ConsumableTypeScroll }),
			consumable({ id: 5633, name: 'Great Rage Potion', type: ConsumableType.ConsumableTypePotion }),
			consumable({ id: 13442, name: 'Mighty Rage Potion', type: ConsumableType.ConsumableTypePotion }),
		];
		expect(highestRanks(ranks).map(c => c.id)).toEqual([10309, 5633, 13442]);
		const higher = higherRanks(ranks);
		expect(higher.get(1477)!.map(c => c.id)).toEqual([10309, 4425]);
		expect(higher.get(10309)).toEqual([]);
		expect(higher.get(5633)).toEqual([]);
	});
});

describe('explosiveUsedBeside', () => {
	it('names an explosive with a cooldown of its own', () => {
		expect(explosiveUsedBeside(consumable({ itemEffects: [onUse(24, 300000)] }))).toBe(true);
		expect(explosiveUsedBeside(consumable({ itemEffects: [onUse(24)] }))).toBe(false);
	});
});

describe('relevantConsumableIds', () => {
	it('lists the relevant consumables of one slot, the highest rank of each', () => {
		const consumables = [
			consumable({ id: 13442, name: 'Mighty Rage Potion', type: ConsumableType.ConsumableTypePotion, categoryId: 4, effectIds: [3] }),
			consumable({ id: 13444, name: 'Major Mana Potion', type: ConsumableType.ConsumableTypePotion, categoryId: 4, effectIds: [1] }),
			consumable({ id: 9421, name: 'Major Healthstone', type: ConsumableType.ConsumableTypeConjured, categoryId: 1153, effectIds: [2] }),
		];
		expect(relevantConsumableIds(consumables, ConsumableSlot.Potion, user(), db)).toEqual([13442]);
		expect(relevantConsumableIds(consumables, ConsumableSlot.Conjured, user({ epStats: [Stat.StatStamina] }), db)).toEqual([9421]);
	});

	it('picks the rank among the relevant ones, so an unusable top rank leaves the one below', () => {
		const consumables = [
			consumable({ id: 1, name: 'Rage Potion', type: ConsumableType.ConsumableTypePotion, categoryId: 4, effectIds: [3] }),
			consumable({
				id: 2,
				name: 'Rage Potion II',
				type: ConsumableType.ConsumableTypePotion,
				categoryId: 4,
				effectIds: [3],
				classAllowlist: [Class.ClassRogue],
			}),
		];
		expect(relevantConsumableIds(consumables, ConsumableSlot.Potion, user(), db)).toEqual([1]);
		expect(relevantConsumableIds(consumables, ConsumableSlot.Potion, user({ cls: Class.ClassRogue }), db)).toEqual([2]);
	});
});
