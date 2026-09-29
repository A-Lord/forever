import type { Meta, StoryObj } from '@storybook/react-vite';
import { Table, TableCell, TableRow } from '@ui-kit/Table';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { CharacterStatsPanelView, type CharacterStatsPanelViewProps } from './CharacterStatsPanelView';

const ROWS = [
	['Strength', '320'],
	['Agility', '180'],
	['Stamina', '410'],
	['Attack Power', '1,240'],
	['Melee Crit', '24.50%'],
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

const StatefulPanel = ({ open: initialOpen, mode: initialMode, onToggle, onModeChange, ...props }: CharacterStatsPanelViewProps) => {
	const [open, setOpen] = useState(initialOpen);
	const [mode, setMode] = useState(initialMode);
	return (
		<CharacterStatsPanelView
			{...props}
			open={open}
			mode={mode}
			onToggle={next => {
				setOpen(next);
				onToggle(next);
			}}
			onModeChange={next => {
				setMode(next);
				onModeChange(next);
			}}>
			<SampleTable />
		</CharacterStatsPanelView>
	);
};

const meta = {
	title: 'Features/CharacterStats/Components/CharacterStatsPanelView',
	component: CharacterStatsPanelView,
	tags: ['autodocs'],
	args: {
		open: true,
		onToggle: fn(),
		mode: 'key',
		onModeChange: fn(),
		labels: { title: 'Stats', modeAriaLabel: 'Stats shown', key: 'Key', all: 'All' },
	},
	argTypes: { mode: { control: 'inline-radio', options: ['key', 'all'] }, children: { control: false } },
	render: args => <StatefulPanel {...args} />,
	decorators: [
		Story => (
			<div className="flex w-63 flex-col">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof CharacterStatsPanelView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Open: Story = {
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: 'All' }));
		await expect(args.onModeChange).toHaveBeenCalledWith('all');
		await expect(args.onToggle).not.toHaveBeenCalled();
		await expect(canvas.getByRole('button', { name: 'Stats' })).toHaveAttribute('aria-expanded', 'true');
	},
};

export const Closed: Story = {
	args: { open: false },
	play: async ({ args, canvasElement }) => {
		const trigger = within(canvasElement).getByRole('button', { name: 'Stats' });
		await expect(trigger).toHaveAttribute('aria-expanded', 'false');
		await userEvent.click(trigger);
		await expect(args.onToggle).toHaveBeenCalledWith(true);
		await expect(trigger).toHaveAttribute('aria-expanded', 'true');
	},
};

export const ClosedInScrollingSidebar: Story = {
	args: { open: false },
	decorators: [
		Story => (
			<div className="flex h-48 w-63 flex-col overflow-y-auto">
				{Array.from({ length: 12 }, (_, index) => (
					<p key={index} className="mb-2">
						Sidebar setting {index + 1}
					</p>
				))}
				<Story />
			</div>
		),
	],
};
