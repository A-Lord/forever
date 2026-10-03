import { Class, WeaponSkill } from '@generated/proto/common';
import { translateWeaponSkill } from '@i18n/localization';
import { getClassWeaponSkills } from '@sim/player/classes/capabilities';
import { Player } from '@sim/player/player';
import { createSimStore, PLAYER_FIELDS, type PlayerSlice, seedKeyed, zeroVersions } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { act, fireEvent, screen } from '@testing-library/react';
import { Accordion } from '@ui-kit/Accordion';
import { describe, expect, it } from 'vitest';

import { WeaponSkills } from './WeaponSkills';

const KEY = 3;

// The Player's own setters run against a seeded slice; its constructor needs a whole sim behind it.
const renderForPaladin = () => {
	const store = createSimStore();
	seedKeyed(store, 'players', KEY, { weaponSkills: [], v: zeroVersions(PLAYER_FIELDS) } as unknown as PlayerSlice);
	const player = Object.assign(Object.create(Player.prototype), { sim: { store }, storeKey: KEY, playerSpec: { classID: Class.ClassPaladin } });
	renderWithHost(
		<Accordion value={['weapon-skill']} onValueChange={() => {}}>
			<WeaponSkills />
		</Accordion>,
		{ player },
	);
	return { stored: () => store.getState().players[KEY].weaponSkills };
};

const field = (skill: WeaponSkill) => screen.getByRole('textbox', { name: translateWeaponSkill(skill) }) as HTMLInputElement;

const commit = async (skill: WeaponSkill, text: string) => {
	const input = field(skill);
	await act(async () => {
		input.value = text;
		fireEvent.change(input);
	});
};

describe('WeaponSkills', () => {
	it('shows a field per weapon skill of the class, each at the level 60 maximum', () => {
		renderForPaladin();

		const skills = getClassWeaponSkills(Class.ClassPaladin);
		expect(screen.getAllByRole('textbox')).toHaveLength(skills.length);
		for (const skill of skills) expect(field(skill).value).toBe('300');
	});

	it('stores a skill set off the maximum, and drops it again when set back to the maximum', async () => {
		const { stored } = renderForPaladin();

		await commit(WeaponSkill.WeaponSkillTwoHandedSwords, '305');
		await commit(WeaponSkill.WeaponSkillMaces, '290');
		expect(stored()).toEqual([
			{ skill: WeaponSkill.WeaponSkillMaces, value: 290 },
			{ skill: WeaponSkill.WeaponSkillTwoHandedSwords, value: 305 },
		]);
		expect(field(WeaponSkill.WeaponSkillTwoHandedSwords).value).toBe('305');

		await commit(WeaponSkill.WeaponSkillTwoHandedSwords, '300');
		expect(stored()).toEqual([{ skill: WeaponSkill.WeaponSkillMaces, value: 290 }]);
		expect(field(WeaponSkill.WeaponSkillTwoHandedSwords).value).toBe('300');
	});

	it('keeps a skill between 1 and 350', async () => {
		const { stored } = renderForPaladin();

		await commit(WeaponSkill.WeaponSkillMaces, '400');
		expect(field(WeaponSkill.WeaponSkillMaces).value).toBe('350');

		await commit(WeaponSkill.WeaponSkillMaces, '0');
		expect(stored()).toEqual([{ skill: WeaponSkill.WeaponSkillMaces, value: 1 }]);
	});
});
