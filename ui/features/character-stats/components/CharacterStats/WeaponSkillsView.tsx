import type { WeaponSkill } from '@generated/proto/common';
import { NumberField } from '@ui-kit/TextField';

export interface WeaponSkillField {
	skill: WeaponSkill;
	label: string;
	value: number;
}

export interface WeaponSkillsViewProps {
	fields: ReadonlyArray<WeaponSkillField>;
	onChange: (skill: WeaponSkill, value: number) => void;
}

export const WeaponSkillsView = ({ fields, onChange }: WeaponSkillsViewProps) => (
	<div className="flex flex-col gap-1">
		{fields.map(field => (
			<NumberField
				key={field.skill}
				id={`weapon-skill-${field.skill}`}
				label={field.label}
				layout="split"
				density="compact"
				value={field.value}
				onChange={value => onChange(field.skill, value)}
			/>
		))}
	</div>
);
