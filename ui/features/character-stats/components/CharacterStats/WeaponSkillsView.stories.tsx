import { WeaponSkill } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { fn } from 'storybook/test';

import { type WeaponSkillField, WeaponSkillsView, type WeaponSkillsViewProps } from './WeaponSkillsView';

const PALADIN_FIELDS: Array<WeaponSkillField> = [
	{ skill: WeaponSkill.WeaponSkillAxes, label: 'Axes', value: 300 },
	{ skill: WeaponSkill.WeaponSkillTwoHandedAxes, label: 'Two-Handed Axes', value: 300 },
	{ skill: WeaponSkill.WeaponSkillMaces, label: 'Maces', value: 305 },
	{ skill: WeaponSkill.WeaponSkillTwoHandedMaces, label: 'Two-Handed Maces', value: 300 },
	{ skill: WeaponSkill.WeaponSkillSwords, label: 'Swords', value: 300 },
	{ skill: WeaponSkill.WeaponSkillTwoHandedSwords, label: 'Two-Handed Swords', value: 310 },
	{ skill: WeaponSkill.WeaponSkillPolearms, label: 'Polearms', value: 1 },
];

const StatefulView = ({ fields: initialFields, onChange }: WeaponSkillsViewProps) => {
	const [fields, setFields] = useState(initialFields);
	return (
		<WeaponSkillsView
			fields={fields}
			onChange={(skill, value) => {
				setFields(current => current.map(field => (field.skill === skill ? { ...field, value } : field)));
				onChange(skill, value);
			}}
		/>
	);
};

const meta = {
	title: 'Features/CharacterStats/Components/WeaponSkillsView',
	component: WeaponSkillsView,
	tags: ['autodocs'],
	args: { fields: PALADIN_FIELDS, onChange: fn() },
	argTypes: { fields: { control: false } },
	render: args => <StatefulView {...args} />,
	decorators: [
		Story => (
			<div className="flex w-62 flex-col">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof WeaponSkillsView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
