import { HitKind } from '@generated/proto/api';
import { ActionMetrics, SimResult } from '@sim/proto/sim_result';
import { RANKED_SPELLS, type SpellActionFixture, spellActionsRun, stubRankedSpellIcons, stubSpellIcons } from '@testing/results-metrics';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@i18n/localization', async () => (await import('@testing/sim/localization')).localizationStub);

import type { MetricRow } from './grouping';
import { buildActionRows, mergeActionRow, rankParentId } from './rank_groups';

const { FIREBALL_1, FIREBALL_2, FROSTBOLT } = RANKED_SPELLS;

const groupsOf = async (spells: Array<SpellActionFixture>) => {
	const { request, result } = spellActionsRun(spells);
	const player = (await SimResult.makeNew(request, result)).getFirstPlayer()!;
	return ActionMetrics.groupById(player.getDamageActions());
};

const collapse = () => true;

interface RowShape {
	spellId: number;
	rank: number;
	tag: number;
	subRows?: Array<RowShape>;
}

const shape = (rows: Array<MetricRow<ActionMetrics>>): Array<RowShape> =>
	rows.map(({ metric: { actionId }, subRows }) => ({
		spellId: actionId.spellId,
		rank: actionId.rank,
		tag: actionId.tag,
		...(subRows && { subRows: shape(subRows) }),
	}));

describe('buildActionRows', () => {
	beforeEach(() => {
		stubRankedSpellIcons();
	});

	it('puts the ranks of a spell under one row, and the tags of a rank under that rank', async () => {
		const groups = await groupsOf([
			{ spellId: FIREBALL_1, damage: 100 },
			{ spellId: FROSTBOLT, damage: 50 },
			{ spellId: FIREBALL_2, damage: 300 },
			{ spellId: FIREBALL_2, tag: 1, damage: 30 },
		]);
		const rows = buildActionRows(groups, collapse, true);

		expect(shape(rows)).toEqual([
			{
				spellId: FIREBALL_2,
				rank: 0,
				tag: 0,
				subRows: [
					{ spellId: FIREBALL_1, rank: 1, tag: 0 },
					{
						spellId: FIREBALL_2,
						rank: 2,
						tag: 0,
						subRows: [
							{ spellId: FIREBALL_2, rank: 2, tag: 0 },
							{ spellId: FIREBALL_2, rank: 2, tag: 1 },
						],
					},
				],
			},
			{ spellId: FROSTBOLT, rank: 0, tag: 0 },
		]);
		expect(rows[0].metric.name).toBe('Fireball');
		expect(rows[0].metric.damage).toBe(430);
		expect(rows[0].subRows![1].metric.damage).toBe(330);
		expect(rows[0].subRows![0].metric).toBe(groups[0][0]);
		expect(rows[0].subRows![1].subRows![1].metric).toBe(groups[2][1]);
		expect(rows[1].metric).toBe(groups[1][0]);
	});

	it('builds a row per spell id, ranks apart, while ranks are not grouped', async () => {
		const groups = await groupsOf([
			{ spellId: FIREBALL_1, damage: 100 },
			{ spellId: FIREBALL_2, damage: 300 },
			{ spellId: FIREBALL_2, tag: 1, damage: 30 },
		]);

		expect(shape(buildActionRows(groups, collapse, false))).toEqual([
			{ spellId: FIREBALL_1, rank: 1, tag: 0 },
			{
				spellId: FIREBALL_2,
				rank: 2,
				tag: 0,
				subRows: [
					{ spellId: FIREBALL_2, rank: 2, tag: 0 },
					{ spellId: FIREBALL_2, rank: 2, tag: 1 },
				],
			},
		]);
	});

	it('leaves a spell with one rank, and a group carrying another spell, as they are', async () => {
		const groups = await groupsOf([
			{ spellId: FIREBALL_2, damage: 300 },
			{ spellId: FROSTBOLT, damage: 50 },
		]);
		const withRider = [[...groups[0], ...groups[1]], ...(await groupsOf([{ spellId: FIREBALL_1, damage: 100 }]))];

		for (const sample of [groups, withRider]) {
			expect(shape(buildActionRows(sample, collapse, true))).toEqual(shape(buildActionRows(sample, collapse, false)));
		}
	});

	it('keeps an unranked spell apart from the ranks of a spell of the same name', async () => {
		stubSpellIcons({
			[FIREBALL_1]: { name: 'Fireball', rank: 1 },
			[FIREBALL_2]: { name: 'Fireball', rank: 2 },
			[FROSTBOLT]: { name: 'Fireball' },
		});
		const groups = await groupsOf([
			{ spellId: FIREBALL_1, damage: 100 },
			{ spellId: FROSTBOLT, damage: 50 },
			{ spellId: FIREBALL_2, damage: 300 },
		]);

		expect(shape(buildActionRows(groups, collapse, true))).toEqual([
			{
				spellId: FIREBALL_2,
				rank: 0,
				tag: 0,
				subRows: [
					{ spellId: FIREBALL_1, rank: 1, tag: 0 },
					{ spellId: FIREBALL_2, rank: 2, tag: 0 },
				],
			},
			{ spellId: FROSTBOLT, rank: 0, tag: 0 },
		]);
	});

	it('drops the tag of a rank whose only member carries one, like the ranks merged over their tags', async () => {
		const groups = await groupsOf([
			{ spellId: FIREBALL_1, tag: 1, damage: 100 },
			{ spellId: FIREBALL_2, damage: 300 },
		]);

		expect(shape(buildActionRows(groups, collapse, true))[0].subRows![0]).toEqual({ spellId: FIREBALL_1, rank: 1, tag: 0 });
	});
});

describe('mergeActionRow', () => {
	beforeEach(() => {
		stubRankedSpellIcons();
	});

	it('names the row of joined ranks after the highest rank without its rank, and merges the counts and ranges', async () => {
		const groups = await groupsOf([
			{ spellId: FIREBALL_1, damage: 100 },
			{ spellId: FIREBALL_2, damage: 300 },
		]);
		const ranks = groups.flat();
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
