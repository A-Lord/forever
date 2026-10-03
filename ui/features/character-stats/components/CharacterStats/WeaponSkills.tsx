import i18n from '@i18n/config';
import { translateWeaponSkill } from '@i18n/localization';
import { usePlayer } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { getClassWeaponSkills } from '@sim/player/classes/capabilities';
import { weaponSkillValue } from '@sim/player/player';
import { AccordionItem } from '@ui-kit/Accordion';

import { WeaponSkillsView } from './WeaponSkillsView';

// The maximum at level 60 is 300; the field goes past it for theorycrafting.
const WEAPON_SKILL_FIELD_MIN = 1;
const WEAPON_SKILL_FIELD_MAX = 350;

export const WeaponSkills = () => {
	const player = usePlayer();
	const weaponSkills = usePlayerStore('weaponSkills');
	const fields = getClassWeaponSkills(player.getClass()).map(skill => ({
		skill,
		label: translateWeaponSkill(skill),
		value: weaponSkillValue(weaponSkills, skill),
	}));

	return (
		<AccordionItem value="weapon-skill" title={i18n.t('sidebar.character_stats.weapon_skill')} testId="character-stats-group-weapon-skill">
			<WeaponSkillsView
				fields={fields}
				onChange={(skill, value) => player.setWeaponSkill(skill, Math.min(Math.max(value, WEAPON_SKILL_FIELD_MIN), WEAPON_SKILL_FIELD_MAX))}
			/>
		</AccordionItem>
	);
};
