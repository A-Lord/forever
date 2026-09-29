import type { Meta, StoryObj } from '@storybook/react-vite';
import { Table, TableCell, TableRow } from '@ui-kit/Table';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { CharacterStatsBarView, type CharacterStatsBarViewProps } from './CharacterStatsBarView';

const ROWS = [
	['Strength', '320'],
	['Agility', '180'],
	['Stamina', '410'],
	['Attack Power', '1,240'],
	['Melee Hit', '6.00%'],
	['Melee Crit', '24.50%'],
	['Expertise', '12'],
	['Armor Penetration', '140'],
	['Haste', '3.20%'],
	['Armor', '7,420'],
	['Defense', '350'],
	['Dodge', '9.80%'],
	['Parry', '8.10%'],
	['Block', '5.00%'],
	['Health', '8,950'],
];

const SampleTable = () => (
	<Table variant="plain" flush className="p-2.5">
		<tbody>
			{ROWS.map(([label, value]) => (
				<TableRow key={label} className="ui-character-stats-row">
					<TableCell className="ui-character-stats-label">{label}</TableCell>
					<TableCell className="ui-character-stats-value">{value}</TableCell>
				</TableRow>
			))}
		</tbody>
	</Table>
);

const StatefulBar = ({ open: initialOpen, mode: initialMode, onOpenChange, onModeChange, ...props }: CharacterStatsBarViewProps) => {
	const [open, setOpen] = useState(initialOpen);
	const [mode, setMode] = useState(initialMode);
	return (
		<CharacterStatsBarView
			{...props}
			open={open}
			mode={mode}
			onOpenChange={next => {
				setOpen(next);
				onOpenChange(next);
			}}
			onModeChange={next => {
				setMode(next);
				onModeChange(next);
			}}>
			<SampleTable />
		</CharacterStatsBarView>
	);
};

const meta = {
	title: 'Features/CharacterStats/Components/CharacterStatsBarView',
	component: CharacterStatsBarView,
	tags: ['autodocs'],
	args: {
		open: false,
		onOpenChange: fn(),
		mode: 'key',
		onModeChange: fn(),
		labels: { title: 'Stats', modeAriaLabel: 'Stats shown', key: 'Key', all: 'All' },
	},
	argTypes: { mode: { control: 'inline-radio', options: ['key', 'all'] }, children: { control: false } },
	render: args => <StatefulBar {...args} />,
} satisfies Meta<typeof CharacterStatsBarView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Closed: Story = {
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: 'All' }));
		await expect(args.onModeChange).toHaveBeenCalledWith('all');
		await expect(args.onOpenChange).not.toHaveBeenCalled();
		await expect(canvas.getByRole('button', { name: 'Stats' })).toHaveAttribute('aria-expanded', 'false');
	},
};

export const Opens: Story = {
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: 'Stats' }));
		await expect(args.onOpenChange).toHaveBeenCalledWith(true);
		await expect(await canvas.findByRole('dialog', { name: 'Stats' })).toBeInTheDocument();
	},
};

export const Open: Story = {
	args: { open: true },
};
