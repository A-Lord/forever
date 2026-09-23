// Which consumables the settings offer a spec, and in which picker, from the fields db.json carries
// for each. `Player.getConsumes(true)` sends the potions and conjured items this finds, so it sits
// below the view layer; the pickers in the settings feature read the same functions.
import { Class, ConsumableType, HandType, ItemSlot, Profession, Race, Stat, WeaponType } from '@generated/proto/common';
import { Consumable } from '@generated/proto/db';
import { ResourceType, SpellEffect } from '@generated/proto/spell';
import { UIItem as Item } from '@generated/proto/ui';

import type { SpecConfigData } from '../player/player';
import type { EquippedItem } from '../proto/equipped_item';

// The spell categories an on-use shares its cooldown through (SpellCategory.ID).
const COMBAT_POTION_CATEGORY = 4; // Item - Combat Consumable (Potion)
const FOOD_CATEGORY = 11; // Item - Food
const DRINK_CATEGORY = 59; // Item - Drink
const CONJURED_CATEGORY = 1153; // Item - Combat Consumable (Non-Aggressive)

export enum ConsumableSlot {
	Potion,
	Conjured,
	Flask,
	Elixir,
	Food,
	Explosive,
	Imbue,
	PetFood,
	// Everything else, picked any number at a time into extra_ids.
	Extra,
}

export interface ConsumableUser {
	getClass(): Class;
	getRace(): Race;
	hasProfession(profession: Profession): boolean;
	getEquippedItem(slot: ItemSlot): EquippedItem | null;
	getSpecConfig(): Pick<SpecConfigData<any>, 'epStats' | 'consumableStats'>;
}

export interface SpellEffectSource {
	getSpellEffect(effectId: number): SpellEffect | undefined;
}

// The generator writes the category the on-use shares, the item's or, where the item states none, its
// spell's, as the consumable's.
const onUseCategory = (consumable: Consumable): number => consumable.categoryId;

const ELIXIR_TYPES = [ConsumableType.ConsumableTypeElixir, ConsumableType.ConsumableTypeBattleElixir, ConsumableType.ConsumableTypeGuardianElixir];

export function consumableSlot(consumable: Consumable): ConsumableSlot {
	const category = onUseCategory(consumable);
	if (category === CONJURED_CATEGORY) return ConsumableSlot.Conjured;
	if (consumable.type === ConsumableType.ConsumableTypePotion || category === COMBAT_POTION_CATEGORY) return ConsumableSlot.Potion;
	if (consumable.type === ConsumableType.ConsumableTypeFlask) return ConsumableSlot.Flask;
	if (ELIXIR_TYPES.includes(consumable.type)) return ConsumableSlot.Elixir;
	if (consumable.type === ConsumableType.ConsumableTypeFood || category === FOOD_CATEGORY || category === DRINK_CATEGORY) return ConsumableSlot.Food;
	if (consumable.type === ConsumableType.ConsumableTypeExplosive) return ConsumableSlot.Explosive;
	if (consumable.type === ConsumableType.ConsumableTypeImbue) return ConsumableSlot.Imbue;
	if (consumable.type === ConsumableType.ConsumableTypePetFood) return ConsumableSlot.PetFood;
	return ConsumableSlot.Extra;
}

// An explosive with a cooldown of its own besides the one explosives share is used beside the
// spec's explosive, so it is picked into extra_ids.
export const explosiveUsedBeside = (consumable: Consumable): boolean =>
	consumable.itemEffects.some(effect => effect.effect.oneofKind === 'onUse' && effect.effect.onUse.cooldownMs > 0);

// ChrRaces.ID of each race: bit ID - 1 of ItemSparse.AllowableRace.
const CLIENT_RACE_IDS: Partial<Record<Race, number>> = {
	[Race.RaceHuman]: 1,
	[Race.RaceOrc]: 2,
	[Race.RaceDwarf]: 3,
	[Race.RaceNightElf]: 4,
	[Race.RaceUndead]: 5,
	[Race.RaceTauren]: 6,
	[Race.RaceGnome]: 7,
	[Race.RaceTroll]: 8,
	[Race.RaceBloodElf]: 10,
	[Race.RaceDraenei]: 11,
};

// ItemSubclassWeapon of an equipped weapon: the bit a temporary enchant's weapon subclass mask sets
// for it. None for a shield or an off-hand frill.
const clientWeaponSubclass = (item: Item): number | undefined => {
	const twoHand = item.handType === HandType.HandTypeTwoHand;
	switch (item.weaponType) {
		case WeaponType.WeaponTypeAxe:
			return twoHand ? 1 : 0;
		case WeaponType.WeaponTypeMace:
			return twoHand ? 5 : 4;
		case WeaponType.WeaponTypeSword:
			return twoHand ? 8 : 7;
		case WeaponType.WeaponTypePolearm:
			return 6;
		case WeaponType.WeaponTypeStaff:
			return 10;
		case WeaponType.WeaponTypeFist:
			return 13;
		case WeaponType.WeaponTypeDagger:
			return 15;
		default:
			return undefined;
	}
};

// InventoryType of an item in a hand: the bit a temporary enchant's inventory type mask sets for it.
const clientInventoryType = (item: Item): number | undefined => {
	if (item.weaponType === WeaponType.WeaponTypeShield) return 14;
	if (item.weaponType === WeaponType.WeaponTypeOffHand) return 23;
	switch (item.handType) {
		case HandType.HandTypeOneHand:
			return 13;
		case HandType.HandTypeTwoHand:
			return 17;
		case HandType.HandTypeMainHand:
			return 21;
		case HandType.HandTypeOffHand:
			return 22;
		default:
			return undefined;
	}
};

const maskAllows = (mask: number, bit: number | undefined): boolean => mask === 0 || (bit !== undefined && (mask & (1 << bit)) !== 0);

export function handTakesImbue(consumable: Consumable, item: Item | undefined): boolean {
	if (!item) return false;
	return maskAllows(consumable.weaponSubclassMask, clientWeaponSubclass(item)) && maskAllows(consumable.weaponInventoryTypeMask, clientInventoryType(item));
}

// The stats a spec weighs consumables by. Stamina brings health and the resistances with it, and
// Attack Power the flat physical damage a Frenzy Potion or Bogling Root adds.
export function consumableStatsOf(specConfig: Pick<SpecConfigData<any>, 'epStats' | 'consumableStats'>): Set<Stat> {
	const stats = new Set([...(specConfig.consumableStats ?? []), ...specConfig.epStats]);
	if (stats.has(Stat.StatAttackPower)) stats.add(Stat.StatPhysicalDamage);
	if (stats.has(Stat.StatStamina)) {
		for (const stat of [
			Stat.StatHealth,
			Stat.StatArcaneResistance,
			Stat.StatFireResistance,
			Stat.StatFrostResistance,
			Stat.StatNatureResistance,
			Stat.StatShadowResistance,
		]) {
			stats.add(stat);
		}
	}
	return stats;
}

const restoredResources = (consumable: Consumable, db: SpellEffectSource): Array<ResourceType> =>
	consumable.effectIds.flatMap(effectId => {
		const miscValue = db.getSpellEffect(effectId)?.miscValue0;
		return miscValue?.oneofKind === 'resourceType' ? [miscValue.resourceType] : [];
	});

const restoresForSpec = (resource: ResourceType, stats: Set<Stat>): boolean => {
	switch (resource) {
		case ResourceType.ResourceTypeMana:
			return stats.has(Stat.StatMana) || stats.has(Stat.StatMP5) || stats.has(Stat.StatIntellect);
		case ResourceType.ResourceTypeHealth:
			return stats.has(Stat.StatHealth);
		default:
			return stats.has(Stat.StatAttackPower);
	}
};

// Types whose consumables exist to grant a stat: one granting none the spec weighs does nothing for it.
const STAT_TYPES = [
	ConsumableType.ConsumableTypeUnknown,
	ConsumableType.ConsumableTypeFlask,
	ConsumableType.ConsumableTypeFood,
	ConsumableType.ConsumableTypeScroll,
	ConsumableType.ConsumableTypeOther,
	...ELIXIR_TYPES,
];

// Whether the settings offer the consumable to this player: the class, race and profession the
// client allows it to, and either a stat the spec weighs, a resource the spec spends, or, for a
// consumable that grants neither, an effect of another kind (damage, an absorb, a weapon enchant).
// A temporary enchant is checked against a hand with handTakesImbue.
export function isConsumableRelevant(
	consumable: Consumable,
	user: ConsumableUser,
	db: SpellEffectSource,
	stats = consumableStatsOf(user.getSpecConfig()),
): boolean {
	if (consumable.classAllowlist.length && !consumable.classAllowlist.includes(user.getClass())) return false;
	if (consumable.requiredProfession !== Profession.ProfessionUnknown && !user.hasProfession(consumable.requiredProfession)) return false;
	if (consumable.allowableRace) {
		const raceId = CLIENT_RACE_IDS[user.getRace()];
		if (raceId === undefined || (consumable.allowableRace & (1n << BigInt(raceId - 1))) === 0n) return false;
	}

	const grantedStats = consumable.stats.flatMap((value, stat) => (value > 0 ? [stat as Stat] : []));
	const restores = restoredResources(consumable, db);
	if (consumable.buffsMainStat || grantedStats.some(stat => stats.has(stat)) || restores.some(resource => restoresForSpec(resource, stats))) return true;
	return !grantedStats.length && !restores.length && !STAT_TYPES.includes(consumable.type);
}

// Roman-numeral ranks of one consumable share a name up to the numeral; the settings offer the
// highest.
const RANK_SUFFIX = / (I|II|III|IV|V|VI|VII|VIII|IX|X)$/;
const ROMAN = ['I', 'II', 'III', 'IV', 'V', 'VI', 'VII', 'VIII', 'IX', 'X'];

export function highestRanks(consumables: ReadonlyArray<Consumable>): Array<Consumable> {
	const rankOf = (consumable: Consumable) => {
		const match = consumable.name.match(RANK_SUFFIX);
		return { base: `${consumable.type}:${match ? consumable.name.slice(0, match.index) : consumable.name}`, rank: match ? ROMAN.indexOf(match[1]) + 1 : 1 };
	};
	const best = new Map<string, { rank: number; consumable: Consumable }>();
	for (const consumable of consumables) {
		const { base, rank } = rankOf(consumable);
		const current = best.get(base);
		if (!current || rank > current.rank) best.set(base, { rank, consumable });
	}
	const kept = new Set([...best.values()].map(entry => entry.consumable));
	return consumables.filter(consumable => kept.has(consumable));
}

// The ids of the potions and conjured items the spec could use, which a sim request lists for the
// rotation to reach.
export function relevantConsumableIds(
	consumables: ReadonlyArray<Consumable>,
	slot: ConsumableSlot,
	user: ConsumableUser,
	db: SpellEffectSource,
): Array<number> {
	const stats = consumableStatsOf(user.getSpecConfig());
	return highestRanks(consumables.filter(consumable => consumableSlot(consumable) === slot))
		.filter(consumable => isConsumableRelevant(consumable, user, db, stats))
		.map(consumable => consumable.id);
}
