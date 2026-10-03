import type { WeaponSkill } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateWeaponSkill } from '@i18n/localization';
import { usePlayer, useSpecConfig } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { getClassWeaponSkills } from '@sim/player/classes/capabilities';
import { MAX_WEAPON_SKILL, weaponSkillValue } from '@sim/player/player';
import { weaponSkillsInGroups } from '@sim/player/weapon_skill_groups';
import { AccordionItem } from '@ui-kit/Accordion';
import { Table, TableCell, TableRow } from '@ui-kit/Table';
import type { ReactNode } from 'react';

import { BonusLink } from './BonusLink';
import type { CharacterStatsMode } from './CharacterStatsHeaderView';
import { bonusStatClass } from './utils/stat_display';
import { WeaponSkillPicker } from './WeaponSkillPicker';

export interface WeaponSkillRow {
	skill: WeaponSkill;
	label: string;
	value: number;
	/** The value's distance from the level maximum, which colours it. */
	bonus: number;
	bonusLink?: ReactNode;
}

export interface WeaponSkillsViewProps {
	rows: ReadonlyArray<WeaponSkillRow>;
}

export const WeaponSkillsView = ({ rows }: WeaponSkillsViewProps) => (
	<Table variant="plain" flush>
		<tbody>
			{rows.map(row => (
				<TableRow key={row.skill} className="ui-character-stats-row">
					<TableCell className="ui-character-stats-label">{row.label}</TableCell>
					<TableCell className="ui-character-stats-value">
						<div className="ui-stat-value-link-container" data-tone={row.bonus === 0 ? 'neutral' : row.bonus > 0 ? 'positive' : 'negative'}>
							<span className={bonusStatClass(row.bonus)}>{row.value}</span>
						</div>
						{row.bonusLink}
					</TableCell>
				</TableRow>
			))}
		</tbody>
	</Table>
);

export interface WeaponSkillsProps {
	mode: CharacterStatsMode;
}

export const WeaponSkills = ({ mode }: WeaponSkillsProps) => {
	const player = usePlayer();
	const { displayWeaponSkillGroups } = useSpecConfig();
	const weaponSkills = usePlayerStore('weaponSkills');
	const classSkills = getClassWeaponSkills(player.getClass());
	const keySkills = weaponSkillsInGroups(displayWeaponSkillGroups ?? []);
	const shown = mode === 'all' ? classSkills : classSkills.filter(skill => keySkills.includes(skill));
	if (shown.length === 0) return null;

	const bonusPrefix = i18n.t('sidebar.character_stats.bonus_prefix');
	const rows = shown.map(skill => {
		const label = translateWeaponSkill(skill);
		const bonusLabel = `${bonusPrefix} ${label}`;
		const value = weaponSkillValue(weaponSkills, skill);
		return {
			skill,
			label,
			value,
			bonus: value - MAX_WEAPON_SKILL,
			bonusLink: <BonusLink label={bonusLabel} picker={onCommit => <WeaponSkillPicker skill={skill} label={bonusLabel} onCommit={onCommit} />} />,
		};
	});

	return (
		<AccordionItem
			value="weapon-skill"
			title={i18n.t('sidebar.character_stats.weapon_skill')}
			panelClassName="pb-1"
			testId="character-stats-group-weapon-skill">
			<WeaponSkillsView rows={rows} />
		</AccordionItem>
	);
};
