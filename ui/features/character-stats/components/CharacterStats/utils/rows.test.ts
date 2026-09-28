import { PseudoStat, Stat } from '@generated/proto/common';
import type { Player } from '@sim/player/player';
import { UnitStat } from '@sim/proto/stats';
import { describe, expect, it } from 'vitest';

import { buildRows } from './rows';

const fakePlayer = (spec: { isTankSpec?: boolean; isMeleeDpsSpec?: boolean }) =>
	({ getPlayerSpec: () => ({ isTankSpec: false, isMeleeDpsSpec: false, ...spec }) }) as unknown as Player<any>;

const kinds = (layout: ReturnType<typeof buildRows>, key: string) => layout.groups.find(group => group.key === key)?.rows.map(row => row.kind);

describe('buildRows', () => {
	it('keeps group order and includes the Resistance group', () => {
		const statList = [
			UnitStat.fromStat(Stat.StatHealth),
			UnitStat.fromStat(Stat.StatAttackPower),
			UnitStat.fromStat(Stat.StatSpellDamage),
			UnitStat.fromStat(Stat.StatArmor),
			UnitStat.fromStat(Stat.StatShadowResistance),
		];
		expect(buildRows(fakePlayer({}), statList).groups.map(group => group.key)).toEqual(['Primary', 'Physical', 'Spell', 'Defense', 'Resistance']);
	});

	it('drops a group nothing in it is displayed for', () => {
		const layout = buildRows(fakePlayer({}), [UnitStat.fromStat(Stat.StatHealth)]);
		expect(layout.groups.map(group => group.key)).toEqual(['Primary']);
	});

	it('renders the split melee/ranged hit, crit and speed pseudo stats apart', () => {
		const split = [
			PseudoStat.PseudoStatMeleeHitPercent,
			PseudoStat.PseudoStatMeleeCritPercent,
			PseudoStat.PseudoStatRangedHitPercent,
			PseudoStat.PseudoStatRangedCritPercent,
			PseudoStat.PseudoStatMeleeSpeedMultiplier,
			PseudoStat.PseudoStatRangedSpeedMultiplier,
		].map(pseudoStat => UnitStat.fromPseudoStat(pseudoStat));

		const layout = buildRows(fakePlayer({}), split);
		expect(layout.shownStats.map(stat => stat.getKey())).toEqual(split.map(stat => stat.getKey()));
	});

	it('renders the six school hit percents under the Spell group', () => {
		const schools = [
			PseudoStat.PseudoStatSchoolHitPercentArcane,
			PseudoStat.PseudoStatSchoolHitPercentFire,
			PseudoStat.PseudoStatSchoolHitPercentFrost,
			PseudoStat.PseudoStatSchoolHitPercentHoly,
			PseudoStat.PseudoStatSchoolHitPercentNature,
			PseudoStat.PseudoStatSchoolHitPercentShadow,
		].map(pseudoStat => UnitStat.fromPseudoStat(pseudoStat));

		const layout = buildRows(fakePlayer({}), schools);
		expect(kinds(layout, 'Spell')).toEqual(Array(6).fill('stat'));
	});

	it('hangs the tank readouts off the end of the Defense group', () => {
		const layout = buildRows(fakePlayer({ isTankSpec: true }), [UnitStat.fromStat(Stat.StatDefenseRating)]);
		expect(kinds(layout, 'Defense')).toEqual(['stat', 'miss', 'avoidance', 'crit-immunity']);
	});

	it('leaves the tank readouts out for a non-tank', () => {
		const layout = buildRows(fakePlayer({}), [UnitStat.fromStat(Stat.StatDefenseRating)]);
		expect(kinds(layout, 'Defense')).toEqual(['stat']);
	});

	it('folds the off-hand pseudo stats into the main-hand rows instead of giving them their own row', () => {
		const statList = [
			PseudoStat.PseudoStatMeleeHitPercent,
			PseudoStat.PseudoStatMeleeCritPercent,
			PseudoStat.PseudoStatOffHandHitPercent,
			PseudoStat.PseudoStatOffHandCritPercent,
			PseudoStat.PseudoStatMeleeHastePercent,
		].map(pseudoStat => UnitStat.fromPseudoStat(pseudoStat));

		const layout = buildRows(fakePlayer({}), statList);
		expect(kinds(layout, 'Physical')).toEqual(['stat', 'stat', 'stat']);
		expect(layout.shownStats.map(stat => stat.getKey())).toEqual(
			[PseudoStat.PseudoStatMeleeHitPercent, PseudoStat.PseudoStatMeleeCritPercent, PseudoStat.PseudoStatMeleeHastePercent].map(pseudoStat =>
				UnitStat.fromPseudoStat(pseudoStat).getKey(),
			),
		);

		const physical = layout.groups.find(group => group.key === 'Physical')!;
		const meleeHitRow = physical.rows[0];
		const meleeCritRow = physical.rows[1];
		expect(meleeHitRow.kind === 'stat' && meleeHitRow.displayStat.offHandStat?.equalsPseudoStat(PseudoStat.PseudoStatOffHandHitPercent)).toBe(true);
		expect(meleeCritRow.kind === 'stat' && meleeCritRow.displayStat.offHandStat?.equalsPseudoStat(PseudoStat.PseudoStatOffHandCritPercent)).toBe(true);
	});

	it('puts the melee crit cap directly after the Melee Crit row for a melee dps spec', () => {
		const statList = [PseudoStat.PseudoStatMeleeHitPercent, PseudoStat.PseudoStatMeleeCritPercent, PseudoStat.PseudoStatMeleeHastePercent].map(pseudoStat =>
			UnitStat.fromPseudoStat(pseudoStat),
		);

		const layout = buildRows(fakePlayer({ isMeleeDpsSpec: true }), statList);
		expect(layout.groups.map(group => group.key)).toEqual(['Physical']);
		expect(kinds(layout, 'Physical')).toEqual(['stat', 'stat', 'melee-crit-cap', 'stat']);
	});

	it('leaves the melee crit cap out for a non-melee-dps spec', () => {
		const statList = [UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent), UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeCritPercent)];
		const layout = buildRows(fakePlayer({}), statList);
		expect(kinds(layout, 'Physical')).toEqual(['stat', 'stat']);
	});
});
