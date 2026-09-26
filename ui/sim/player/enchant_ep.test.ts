import { ItemSlot, PseudoStat, Stat } from '@generated/proto/common';
import { UIEnchant as Enchant, UIItem as Item } from '@generated/proto/ui';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@i18n/localization', () => ({
	translateStat: (stat: unknown) => String(stat),
	translatePseudoStat: (pseudoStat: unknown) => String(pseudoStat),
}));

import * as Mechanics from '../constants/mechanics';
import { Stats } from '../proto/stats';
import { Player } from './player';

const weights = Stats.fromMap({ [Stat.StatStrength]: 2 }, { [PseudoStat.PseudoStatMainHandDps]: 10, [PseudoStat.PseudoStatOffHandDps]: 4 });

const computeEnchantEP = (enchant: Enchant, slot?: ItemSlot, weapon?: Item | null, epWeights: Stats = weights) => {
	const player = { enchantEPCache: new Map<number, number>(), computeStatsEP: (stats: Stats) => stats.computeEP(epWeights) } as unknown as Player<any>;
	const epPerWeaponDps = slot === undefined ? 0 : Player.prototype.computeWeaponDpsEP.call(player, slot);
	return Player.prototype.computeEnchantEP.call(player, enchant, weapon?.weaponSpeed ?? 0, epPerWeaponDps);
};

const striking = Enchant.create({ effectId: 1897, weaponDamage: 5 });
const sword = Item.create({ id: 1, weaponSpeed: 2.5 });

describe('Player.computeEnchantEP', () => {
	it('values flat weapon damage as the DPS it adds at the weapon speed of the slot it enchants', () => {
		expect(computeEnchantEP(striking, ItemSlot.ItemSlotMainHand, sword)).toBeCloseTo((5 / 2.5) * 10);
		expect(computeEnchantEP(striking, ItemSlot.ItemSlotOffHand, sword)).toBeCloseTo((5 / 2.5) * 4);
	});

	it('adds the weapon damage on top of the stats', () => {
		const enchant = Enchant.create({ effectId: 2, weaponDamage: 5, stats: new Stats().withStat(Stat.StatStrength, 3).asProtoArray() });
		expect(computeEnchantEP(enchant, ItemSlot.ItemSlotMainHand, sword)).toBeCloseTo(3 * 2 + (5 / 2.5) * 10);
	});

	it('gives weapon damage nothing without a weapon to read the speed off', () => {
		expect(computeEnchantEP(striking, ItemSlot.ItemSlotMainHand, null)).toBe(0);
		expect(computeEnchantEP(striking)).toBe(0);
	});
});

describe('Player.computeEnchantEP on percent pseudo stats', () => {
	const ratingWeights = Stats.fromMap({ [Stat.StatDodgeRating]: 2, [Stat.StatCritRating]: 3, [Stat.StatHitRating]: 5 }, {});
	const enchantEP = (effectId: number, pseudoStats: Partial<Record<PseudoStat, number>>, epWeights = ratingWeights) =>
		computeEnchantEP(Enchant.create({ effectId, pseudoStats: Stats.fromMap({}, pseudoStats).toProto().pseudoStats }), undefined, undefined, epWeights);

	it('values 2622 Enchant Cloak - Dodge at the dodge rating weight', () => {
		expect(enchantEP(2622, { [PseudoStat.PseudoStatDodgePercent]: 1 })).toBeCloseTo(Mechanics.DODGE_RATING_PER_DODGE_PERCENT * 2);
	});

	it("counts 2717 Might of the Scourge's crit once, though it states melee and the ranged total", () => {
		expect(enchantEP(2717, { [PseudoStat.PseudoStatMeleeCritPercent]: 1, [PseudoStat.PseudoStatRangedCritPercent]: 1 })).toBeCloseTo(
			Mechanics.CRIT_RATING_PER_CRIT_PERCENT * 3,
		);
	});

	it('values the ranged-only 2523 Biznicks 247x128 Accurascope at the hit rating weight', () => {
		expect(enchantEP(2523, { [PseudoStat.PseudoStatRangedHitPercent]: 3 })).toBeCloseTo(3 * Mechanics.HIT_RATING_PER_HIT_PERCENT * 5);
	});

	it('values a melee-only hit percent at the hit rating weight, with no ranged total to take it back', () => {
		expect(enchantEP(1, { [PseudoStat.PseudoStatMeleeHitPercent]: 1 })).toBeCloseTo(Mechanics.HIT_RATING_PER_HIT_PERCENT * 5);
	});

	it("uses a percent pseudo stat's own weight where it has one", () => {
		const ownWeight = Stats.fromMap({ [Stat.StatCritRating]: 3 }, { [PseudoStat.PseudoStatMeleeCritPercent]: 7 });
		expect(enchantEP(2717, { [PseudoStat.PseudoStatMeleeCritPercent]: 1, [PseudoStat.PseudoStatRangedCritPercent]: 1 }, ownWeight)).toBeCloseTo(7);
	});

	it("values 2586 Falcon's Call's hit at a weighted ranged total alone, which carries the melee share", () => {
		const hunterWeights = Stats.fromMap({ [Stat.StatHitRating]: 5 }, { [PseudoStat.PseudoStatRangedHitPercent]: 5 });
		expect(enchantEP(2586, { [PseudoStat.PseudoStatMeleeHitPercent]: 1, [PseudoStat.PseudoStatRangedHitPercent]: 1 }, hunterWeights)).toBeCloseTo(5);
	});

	it('counts melee and spell hit once, as the largest, since one hit rating raises both', () => {
		expect(enchantEP(1, { [PseudoStat.PseudoStatMeleeHitPercent]: 1, [PseudoStat.PseudoStatSpellHitPercent]: 2 })).toBeCloseTo(
			2 * Mechanics.HIT_RATING_PER_HIT_PERCENT * 5,
		);
	});
});

describe('Player.computeEnchantEP on haste pseudo stats', () => {
	const hasteWeights = Stats.fromMap({ [Stat.StatHasteRating]: 2 }, {});
	const enchantEP = (effectId: number, pseudoStats: Partial<Record<PseudoStat, number>>) =>
		computeEnchantEP(Enchant.create({ effectId, pseudoStats: Stats.fromMap({}, pseudoStats).toProto().pseudoStats }), undefined, undefined, hasteWeights);

	it('values 34 Weapon Counterweight at the haste rating weight', () => {
		expect(enchantEP(34, { [PseudoStat.PseudoStatMeleeHastePercent]: 3 })).toBeCloseTo(3 * Mechanics.HASTE_RATING_PER_HASTE_PERCENT * 2);
	});

	it("counts 931 Enchant Gloves - Minor Haste's melee, ranged and cast speed once, as the one haste rating raises all three", () => {
		expect(
			enchantEP(931, {
				[PseudoStat.PseudoStatMeleeHastePercent]: 1,
				[PseudoStat.PseudoStatRangedHastePercent]: 1,
				[PseudoStat.PseudoStatSpellHastePercent]: 1,
			}),
		).toBeCloseTo(Mechanics.HASTE_RATING_PER_HASTE_PERCENT * 2);
	});

	it('values ranged haste beyond the melee haste beside it', () => {
		expect(enchantEP(1, { [PseudoStat.PseudoStatRangedHastePercent]: 2 })).toBeCloseTo(2 * Mechanics.HASTE_RATING_PER_HASTE_PERCENT * 2);
	});
});

describe('Player.computeEnchantEP on haste pseudo stats for a spec weighing ranged haste', () => {
	const rangedHasteWeights = Stats.fromMap({ [Stat.StatHasteRating]: 2 }, { [PseudoStat.PseudoStatRangedHastePercent]: 7 });
	const enchantEP = (effectId: number, pseudoStats: Partial<Record<PseudoStat, number>>) =>
		computeEnchantEP(
			Enchant.create({ effectId, pseudoStats: Stats.fromMap({}, pseudoStats).toProto().pseudoStats }),
			undefined,
			undefined,
			rangedHasteWeights,
		);

	it('keeps 34 Weapon Counterweight at the haste rating weight, as its melee haste is no part of the ranged total', () => {
		expect(enchantEP(34, { [PseudoStat.PseudoStatMeleeHastePercent]: 3, [PseudoStat.PseudoStatRangedHastePercent]: 0 })).toBeCloseTo(
			3 * Mechanics.HASTE_RATING_PER_HASTE_PERCENT * 2,
		);
	});

	it('counts melee and ranged haste together once, at the ranged haste weight', () => {
		expect(enchantEP(1, { [PseudoStat.PseudoStatMeleeHastePercent]: 1, [PseudoStat.PseudoStatRangedHastePercent]: 1 })).toBeCloseTo(7);
	});

	it('values ranged haste alone at the ranged haste weight', () => {
		expect(enchantEP(1, { [PseudoStat.PseudoStatRangedHastePercent]: 2 })).toBeCloseTo(2 * 7);
	});
});
