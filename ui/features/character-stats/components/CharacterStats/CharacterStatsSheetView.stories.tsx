import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '@ui-kit/Button';
import { Table, TableCell, TableRow } from '@ui-kit/Table';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { CharacterStatsSheetView, type CharacterStatsSheetViewProps } from './CharacterStatsSheetView';

const ROWS = Array.from({ length: 40 }, (_, index) => [`Stat ${index + 1}`, `${(index + 1) * 10}`]);

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

const StatefulSheet = ({ open: initialOpen, mode: initialMode, onOpenChange, onModeChange, ...props }: CharacterStatsSheetViewProps) => {
	const [open, setOpen] = useState(initialOpen);
	const [mode, setMode] = useState(initialMode);
	return (
		<CharacterStatsSheetView
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
		</CharacterStatsSheetView>
	);
};

const meta = {
	title: 'Features/CharacterStats/Components/CharacterStatsSheetView',
	component: CharacterStatsSheetView,
	tags: ['autodocs'],
	args: {
		open: true,
		onOpenChange: fn(),
		trigger: <Button>Open stats</Button>,
		mode: 'key',
		onModeChange: fn(),
		labels: { title: 'Stats', modeAriaLabel: 'Stats shown', key: 'Key', all: 'All' },
	},
	argTypes: { mode: { control: 'inline-radio', options: ['key', 'all'] }, trigger: { control: false }, children: { control: false } },
	render: args => <StatefulSheet {...args} />,
} satisfies Meta<typeof CharacterStatsSheetView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Open: Story = {
	play: async ({ args, canvasElement }) => {
		const sheet = within(await within(canvasElement).findByRole('dialog', { name: 'Stats' }));
		await userEvent.click(sheet.getByRole('button', { name: 'All' }));
		await expect(args.onModeChange).toHaveBeenCalledWith('all');
		await expect(args.onOpenChange).not.toHaveBeenCalled();
		await userEvent.click(sheet.getByRole('button', { name: 'Stats' }));
		await expect(args.onOpenChange).toHaveBeenCalledWith(false);
	},
};
