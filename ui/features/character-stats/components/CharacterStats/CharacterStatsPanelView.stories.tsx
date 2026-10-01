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
	<Table variant="plain" flush>
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

const StatefulPanel = ({ mode: initialMode, onModeChange, ...props }: CharacterStatsPanelViewProps) => {
	const [mode, setMode] = useState(initialMode);
	return (
		<CharacterStatsPanelView
			{...props}
			mode={mode}
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
		mode: 'key',
		onModeChange: fn(),
		labels: { title: 'Stats', modeAriaLabel: 'Stats shown', key: 'Key', all: 'All' },
	},
	argTypes: { mode: { control: 'inline-radio', options: ['key', 'all'] }, children: { control: false } },
	render: args => <StatefulPanel {...args} />,
	decorators: [
		Story => (
			<div className="flex w-62 flex-col">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof CharacterStatsPanelView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'All' }));
		await expect(args.onModeChange).toHaveBeenCalledWith('all');
	},
};
