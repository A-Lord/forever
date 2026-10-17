import { HitKind } from '@generated/proto/api';
import { ActionMetrics, SimResult } from '@sim/proto/sim_result';
import { RANKED_SPELLS, type SpellActionFixture, spellActionsRun, stubRankedSpellIcons, stubSpellIcons } from '@testing/results-metrics';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@i18n/localization', async () => (await import('@testing/sim/localization')).localizationStub);

import { groupRanks, mergeActionRow, rankParentId } from './rank_groups';

const { FIREBALL_1, FIREBALL_2, FROSTBOLT } = RANKED_SPELLS;

const groupsOf = async (spells: Array<SpellActionFixture>) => {
	const { request, result } = spellActionsRun(spells);
	const player = (await SimResult.makeNew(request, result)).getFirstPlayer()!;
	return ActionMetrics.groupById(player.getDamageActions());
};

const spellIds = (groups: Array<Array<ActionMetrics>>) => groups.map(group => group.map(action => action.actionId.spellId));

describe('groupRanks', () => {
	beforeEach(() => {
		stubRankedSpellIcons();
	});

	it('joins the ranks of a spell into one group, each rank merged over its tags', async () => {
		const groups = await groupsOf([
			{ spellId: FIREBALL_1, damage: 100 },
			{ spellId: FROSTBOLT, damage: 50 },
			{ spellId: FIREBALL_2, damage: 300 },
			{ spellId: FIREBALL_2, tag: 1, damage: 30 },
		]);
		expect(spellIds(groups)).toEqual([[FIREBALL_1], [FROSTBOLT], [FIREBALL_2, FIREBALL_2]]);

		const grouped = groupRanks(groups);
		expect(spellIds(grouped)).toEqual([[FIREBALL_1, FIREBALL_2], [FROSTBOLT]]);
		expect(grouped[0][0]).toBe(groups[0][0]);
		expect(grouped[0][1].damage).toBe(330);
		expect(grouped[0][1].actionId.tag).toBe(0);
		expect(grouped[0][1].actionId.rank).toBe(2);
		expect(grouped[1]).toBe(groups[1]);
	});

	it('leaves a spell with one rank, and a group carrying another spell, as they are', async () => {
		const groups = await groupsOf([
			{ spellId: FIREBALL_2, damage: 300 },
			{ spellId: FROSTBOLT, damage: 50 },
		]);
		const same = (before: Array<Array<ActionMetrics>>) => {
			const after = groupRanks(before);
			return after.length === before.length && after.every((group, index) => group === before[index]);
		};
		expect(same(groups)).toBe(true);

		const withRider = [[...groups[0], ...groups[1]], ...(await groupsOf([{ spellId: FIREBALL_1, damage: 100 }]))];
		expect(same(withRider)).toBe(true);
	});

	it('keeps a ranked spell apart from an unranked spell of the same name', async () => {
		stubSpellIcons({ [FIREBALL_2]: { name: 'Fireball', rank: 2 }, [FROSTBOLT]: { name: 'Fireball' } });
		const groups = await groupsOf([
			{ spellId: FIREBALL_2, damage: 300 },
			{ spellId: FROSTBOLT, damage: 50 },
		]);

		expect(spellIds(groupRanks(groups))).toEqual([[FIREBALL_2], [FROSTBOLT]]);
		expect(rankParentId([groups[0][0], groups[1][0]])).toBeUndefined();
	});
});

describe('mergeActionRow', () => {
	beforeEach(() => {
		stubRankedSpellIcons();
	});

	it('names the row of joined ranks after the highest rank without its rank, and merges the counts and ranges', async () => {
		const [ranks] = groupRanks(
			await groupsOf([
				{ spellId: FIREBALL_1, damage: 100 },
				{ spellId: FIREBALL_2, damage: 300 },
			]),
		);
		const parent = mergeActionRow(ranks);

		expect(parent.name).toBe('Fireball');
		expect(parent.actionId.spellId).toBe(FIREBALL_2);
		expect(parent.actionId.hasRank).toBe(false);
		expect(parent.damage).toBe(400);
		expect(parent.hits).toBe(ranks[0].hits + ranks[1].hits);
		expect(parent.range(HitKind.HitKindHit)!.aggregatorData!.n).toBe(2 * ranks[0].range(HitKind.HitKindHit)!.aggregatorData!.n);
	});

	it('keeps the id of the first member, without its tag, for the tags of one spell', async () => {
		const [tags] = await groupsOf([
			{ spellId: FIREBALL_2, damage: 300 },
			{ spellId: FIREBALL_2, tag: 1, damage: 30 },
		]);

		expect(rankParentId(tags)).toBeUndefined();
		expect(mergeActionRow(tags).actionId.equals(tags[0].actionId.withoutTag())).toBe(true);
	});
});
