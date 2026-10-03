import { Debuffs } from '@generated/proto/buffs';
import { Class, PseudoStat, Spec, Stat } from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import { ActionId } from '../proto/action_id';
import { Stats } from '../proto/stats';
import { dropLoneRanks, Player } from './player';

// getDebuffStats reads nothing but the raid's debuffs, so the method is called against that one
// dependency rather than against a constructed Player, which needs a whole sim behind it.
const debuffStats = (debuffs: Partial<Debuffs>): Stats =>
	Player.prototype.getDebuffStats.call({
		sim: { raid: { getDebuffs: () => Debuffs.create(debuffs) } },
	} as unknown as Player<any>);

describe('Player.getDebuffStats', () => {
	// Judgement of the Crusader raises Holy damage taken and nothing else: Improved Seal of the
	// Crusader, the crit half, has no node in any Forever trait tree, so the sheet credits no crit.
	// Faerie Fire stands for a debuff that moves no stat on the sheet at all.
	it.each<[string, Partial<Debuffs>]>([
		['Judgement of the Crusader', { judgementOfTheCrusader: true }],
		['Faerie Fire', { faerieFire: true }],
	])('credits %s with no crit', (_name, debuffs) => {
		const stats = debuffStats(debuffs);

		expect(stats.getPseudoStat(PseudoStat.PseudoStatMeleeCritPercent)).toBe(0);
		expect(stats.getPseudoStat(PseudoStat.PseudoStatRangedCritPercent)).toBe(0);
		expect(stats.getPseudoStat(PseudoStat.PseudoStatSpellCritPercent)).toBe(0);
	});

	// The number has to be the one HuntersMarkValue states in sim/core/debuffs_auto_gen.go,
	// or the sheet and the simulation disagree about the same debuff.
	it("credits Hunter's Mark with the ranged attack power spell 14325 states", () => {
		expect(debuffStats({ huntersMark: true }).getStat(Stat.StatRangedAttackPower)).toBe(71);
		expect(debuffStats({ huntersMark: false }).getStat(Stat.StatRangedAttackPower)).toBe(0);
	});
});

// A rogue with one weapon and no hit, expertise or crit, behind the target unless placed in front.
// A suppression left out is a pseudo-stat the sim has not sent.
const critCapInfo = (targetLevel: number, { inFront = false, critSuppression }: { inFront?: boolean; critSuppression?: number } = {}) => {
	const pseudoStats: Array<number> = [];
	if (critSuppression !== undefined) pseudoStats[PseudoStat.PseudoStatMeleeCritSuppressionPercent] = critSuppression;
	return Player.prototype.getMeleeCritCapInfo.call({
		slice: () => ({ currentStats: { finalStats: { pseudoStats } } }),
		getDebuffStats: () => new Stats(),
		sim: { encounter: { primaryTarget: { level: targetLevel } } },
		isDualWielding: () => false,
		getClass: () => Class.ClassRogue,
		getInFrontOfTarget: () => inFront,
		getSpec: () => Spec.SpecRogue,
	} as unknown as Player<any>);
};

describe('Player.getMeleeCritCapInfo', () => {
	it.each<[number, number, number, number, number]>([
		[58, 4, 4, 0, -0.4],
		[60, 5, 5, 10, 0],
		[62, 6, 6, 30, 2],
		[63, 9, 6.5, 40, 3],
	])('against a level %i target: %f%% hit cap, %f%% dodge, %f%% glancing, %f%% crit suppression', (level, hitCap, dodge, glancing, suppression) => {
		const info = critCapInfo(level);

		expect(info.meleeHitCap).toBeCloseTo(hitCap);
		expect(info.expertiseCap).toBeCloseTo(dodge);
		expect(info.glancing).toBeCloseTo(glancing);
		expect(info.suppression).toBeCloseTo(suppression);
		expect(info.baseCritCap).toBeCloseTo(100 - glancing + suppression - hitCap - dodge);
	});

	it('adds the parry of a level 63 target to the expertise cap from in front', () => {
		expect(critCapInfo(63, { inFront: true }).expertiseCap).toBeCloseTo(6.5 + 14);
	});

	it.each([4.8, 0])('takes the crit suppression the sim sends, %f%%, over the level table', critSuppression => {
		const info = critCapInfo(63, { critSuppression });

		expect(info.suppression).toBe(critSuppression);
		expect(info.baseCritCap).toBeCloseTo(100 - 40 + critSuppression - 9 - 6.5);
	});
});

describe('dropLoneRanks', () => {
	// Spell ids stand in for names: 2120 and 2121 are two ranks of one spell, 25304 has only one.
	const names: Record<number, string> = { 2120: 'Flamestrike', 2121: 'Flamestrike', 25304: 'Frostbolt' };
	const ranksOf = (ids: Array<ActionId>) =>
		dropLoneRanks(
			ids.map(id => ({ id })),
			id => names[id.spellId],
		).map(({ id }) => id.rank);

	it('keeps the ranks of a spell registered at several', () => {
		expect(ranksOf([ActionId.fromSpellId(2120, 1), ActionId.fromSpellId(2121, 2)])).toEqual([1, 2]);
	});

	it('drops the rank of a spell registered at one', () => {
		expect(ranksOf([ActionId.fromSpellId(25304, 11), ActionId.fromSpellId(2120, 1), ActionId.fromSpellId(2121, 2)])).toEqual([0, 1, 2]);
	});

	// A spell and its DoT, registered as a second spell under another tag, count apart.
	it('counts one spell under two tags as two', () => {
		expect(ranksOf([ActionId.fromSpellId(25304, 11), ActionId.fromSpellId(25304, 11, 1)])).toEqual([0, 0]);
	});
});
