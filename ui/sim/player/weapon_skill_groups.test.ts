import { WeaponSkill } from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import { WEAPON_SKILL_GROUPS, WeaponSkillGroup, weaponSkillsInGroups } from './weapon_skill_groups';

const SKILLS = Object.values(WeaponSkill).filter((skill): skill is WeaponSkill => typeof skill === 'number' && skill !== WeaponSkill.WeaponSkillUnknown);

describe('WEAPON_SKILL_GROUPS', () => {
	it.each(SKILLS.map(skill => [WeaponSkill[skill], skill] as const))('puts %s in exactly one group', (_name, skill) => {
		expect(Object.values(WEAPON_SKILL_GROUPS).filter(skills => skills.includes(skill))).toHaveLength(1);
	});

	it('names no skill outside the weapon skills', () => {
		expect(
			Object.values(WEAPON_SKILL_GROUPS)
				.flat()
				.every(skill => SKILLS.includes(skill)),
		).toBe(true);
	});
});

describe('weaponSkillsInGroups', () => {
	it('takes the skills of every group named, and none for no groups', () => {
		expect(weaponSkillsInGroups([WeaponSkillGroup.Ranged, WeaponSkillGroup.Thrown])).toEqual([
			WeaponSkill.WeaponSkillBows,
			WeaponSkill.WeaponSkillCrossbows,
			WeaponSkill.WeaponSkillGuns,
			WeaponSkill.WeaponSkillThrown,
		]);
		expect(weaponSkillsInGroups([])).toEqual([]);
	});
});
