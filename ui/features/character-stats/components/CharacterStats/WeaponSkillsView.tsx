import type { WeaponSkill } from '@generated/proto/common';
import { Table, TableCell, TableRow } from '@ui-kit/Table';
import type { ReactNode } from 'react';

import { bonusStatClass } from './utils/stat_display';

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
