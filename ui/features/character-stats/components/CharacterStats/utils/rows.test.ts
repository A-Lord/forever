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

	it('shows the off-hand rows, followed by the dual-wield footnote, only while dual wielding', () => {
		const statList = [
			PseudoStat.PseudoStatMeleeHitPercent,
			PseudoStat.PseudoStatMeleeCritPercent,
			PseudoStat.PseudoStatOffHandHitPercent,
			PseudoStat.PseudoStatOffHandCritPercent,
			PseudoStat.PseudoStatMeleeHastePercent,
		].map(pseudoStat => UnitStat.fromPseudoStat(pseudoStat));

		const dualWielding = buildRows(fakePlayer({}), statList, true);
		expect(dualWielding.shownStats.map(stat => stat.getKey())).toEqual(statList.map(stat => stat.getKey()));
		expect(kinds(dualWielding, 'Physical')).toEqual(['stat', 'stat', 'stat', 'stat', 'dual-wield-penalty', 'stat']);

		const oneWeapon = buildRows(fakePlayer({}), statList, false);
		expect(oneWeapon.shownStats.map(stat => stat.getKey())).toEqual([statList[0], statList[1], statList[4]].map(stat => stat.getKey()));
		expect(kinds(oneWeapon, 'Physical')).toEqual(['stat', 'stat', 'stat']);
	});

	// The melee crit cap goes after the whole table, not next to the crit row.
	it('trails the melee crit cap as its own group for a melee dps spec', () => {
		const layout = buildRows(fakePlayer({ isMeleeDpsSpec: true }), [UnitStat.fromStat(Stat.StatHealth)]);
		expect(layout.groups.map(group => group.key)).toEqual(['Primary', 'MeleeCritCap']);
		expect(kinds(layout, 'MeleeCritCap')).toEqual(['melee-crit-cap']);
	});
});
