import { AuraStats as AuraStatsProto, SpellStats as SpellStatsProto, UnitMetadata as UnitMetadataProto } from '@generated/proto/api';
import { Debuffs } from '@generated/proto/buffs';
import { ActionID as ActionIdProto, PseudoStat, Stat } from '@generated/proto/common';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ActionId } from '../proto/action_id';
import { Stats } from '../proto/stats';
import { Player, UnitMetadata } from './player';

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

describe('UnitMetadata ranks', () => {
	// 2120 and 2121 are two ranks of Flamestrike, 25304 the only rank of Frostbolt the sim registers.
	const names: Record<number, string> = { 2120: 'Flamestrike', 2121: 'Flamestrike', 25304: 'Frostbolt' };
	beforeEach(() => {
		vi.spyOn(ActionId, 'getTooltipData').mockImplementation(async id => ({ id: id.spellId, name: names[id.spellId], icon: '' }) as never);
	});

	const actionId = (spellId: number, rank: number, tag = 0) => ActionIdProto.create({ rawId: { oneofKind: 'spellId', spellId }, rank, tag });
	const ranksOf = async (ids: Array<ActionIdProto>) => {
		const metadata = new UnitMetadata();
		await metadata.update(
			UnitMetadataProto.create({
				spells: ids.map(id => SpellStatsProto.create({ id })),
				auras: ids.map(id => AuraStatsProto.create({ id })),
			}),
		);
		const rankMap = (entries: Array<{ id: ActionId }>) => Object.fromEntries(entries.map(({ id }) => [`${id.spellId}|${id.tag}`, id.rank]));
		return { spells: rankMap(metadata.getSpells()), auras: rankMap(metadata.getAuras()) };
	};

	it('keeps the ranks of a spell registered at several', async () => {
		const ranks = { '2120|0': 1, '2121|0': 2 };
		expect(await ranksOf([actionId(2120, 1), actionId(2121, 2)])).toEqual({ spells: ranks, auras: ranks });
	});

	it('drops the rank of a spell registered at one', async () => {
		const ranks = { '25304|0': 0, '2120|0': 1, '2121|0': 2 };
		expect(await ranksOf([actionId(25304, 11), actionId(2120, 1), actionId(2121, 2)])).toEqual({ spells: ranks, auras: ranks });
	});

	// A spell and its DoT, registered as a second spell under another tag, count apart.
	it('counts one spell under two tags as two', async () => {
		const ranks = { '25304|0': 0, '25304|1': 0 };
		expect(await ranksOf([actionId(25304, 11), actionId(25304, 11, 1)])).toEqual({ spells: ranks, auras: ranks });
	});
});
