import { ItemSlot } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { fn } from 'storybook/test';

import { StoryBoolean } from '../../../../testing/storybook/fixtures/settings_reforge_bulk';
import { BulkSettingsView, type BulkSettingsViewProps } from './BulkSettingsView';
import { CombinationsCountView } from './CombinationsCountView';

const pair = (id: string, label: string, first: [string, ItemSlot], second: [string, ItemSlot]) => ({
	id,
	label,
	values: [
		{ name: 'None', value: -1 },
		{ name: first[0], value: first[1] },
		{ name: second[0], value: second[1] },
	],
	value: -1,
	onChange: fn(),
});

const WEAPON_TYPES = (
	<div>
		<h6 className="mb-2">Main hand weapon types</h6>
		<div className="grid grid-cols-2 gap-1">
			{['Sword', 'Axe', 'Mace', 'Dagger'].map(type => (
				<StoryBoolean key={type} id={`story-weapon-type-${type}`} label={type} layout="inline" initial={type === 'Sword'} />
			))}
		</div>
	</div>
);

const StatefulBulkSettings = (props: BulkSettingsViewProps) => {
	const [openGroups, setOpenGroups] = useState(props.openGroups);
	return (
		<BulkSettingsView
			{...props}
			openGroups={openGroups}
			onOpenGroupsChange={groups => {
				setOpenGroups(groups);
				props.onOpenGroupsChange(groups);
			}}
		/>
	);
};

const meta = {
	title: 'Features/Bulk/BulkSettingsView',
	component: BulkSettingsView,
	tags: ['autodocs'],
	render: args => <StatefulBulkSettings {...args} />,
	args: {
		combinationsCount: <CombinationsCountView pending={false} combinations={48} iterations={96000} iterationsLimit={1000000} />,
		canRun: true,
		onRun: fn(),
		openGroups: ['options', 'freezes'],
		onOpenGroupsChange: fn(),
		useLegacyBulkSim: false,
		onUseLegacyBulkSimChange: fn(),
		reforgeSettings: null,
		frozenItems: [
			pair('story-freeze-ring', 'Freeze ring', ['Finger 1', ItemSlot.ItemSlotFinger1], ['Finger 2', ItemSlot.ItemSlotFinger2]),
			pair('story-freeze-trinket', 'Freeze trinket', ['Trinket 1', ItemSlot.ItemSlotTrinket1], ['Trinket 2', ItemSlot.ItemSlotTrinket2]),
		],
		weaponFreeze: null,
		onFrozenWeaponSlotChange: fn(),
	},
	argTypes: {
		combinationsCount: { control: false },
		reforgeSettings: { control: false },
		frozenItems: { control: false },
		weaponFreeze: { control: false },
	},
} satisfies Meta<typeof BulkSettingsView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const DualWield: Story = { args: { weaponFreeze: { value: -1, weaponTypes: WEAPON_TYPES } } };

export const WithReforge: Story = {
	args: {
		openGroups: ['options', 'reforge'],
		reforgeSettings: <StoryBoolean id="story-bulk-reforge-custom-ep" label="Use custom EP values" layout="inline" />,
	},
};

export const CannotRun: Story = { args: { canRun: false } };
