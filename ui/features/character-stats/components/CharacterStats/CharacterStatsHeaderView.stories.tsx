import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { CharacterStatsHeaderView, type CharacterStatsHeaderViewProps } from './CharacterStatsHeaderView';

const StatefulHeader = ({ mode: initial, onModeChange, ...props }: CharacterStatsHeaderViewProps) => {
	const [mode, setMode] = useState(initial);
	return (
		<CharacterStatsHeaderView
			{...props}
			mode={mode}
			onModeChange={next => {
				setMode(next);
				onModeChange(next);
			}}
		/>
	);
};

const meta = {
	title: 'Features/CharacterStats/Components/CharacterStatsHeaderView',
	component: CharacterStatsHeaderView,
	tags: ['autodocs'],
	args: { mode: 'key', onModeChange: fn(), labels: { title: 'Stats', modeAriaLabel: 'Stats shown', key: 'Key', all: 'All' } },
	argTypes: { mode: { control: 'inline-radio', options: ['key', 'all'] } },
	render: args => <StatefulHeader {...args} />,
	decorators: [
		Story => (
			<div className="w-62">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof CharacterStatsHeaderView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Key: Story = {
	play: async ({ args, canvasElement }) => {
		const all = within(canvasElement).getByRole('button', { name: 'All' });
		await userEvent.click(all);
		await expect(args.onModeChange).toHaveBeenCalledWith('all');
		await expect(all).toHaveAttribute('aria-pressed', 'true');
	},
};

export const All: Story = { args: { mode: 'all' } };
