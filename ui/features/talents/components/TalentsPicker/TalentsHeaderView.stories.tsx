import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { TalentsHeaderView } from './TalentsHeaderView';

const meta = {
	title: 'Features/Talents/TalentsHeaderView',
	component: TalentsHeaderView,
	tags: ['autodocs'],
	args: { pointsRemaining: 51, copied: false, onCopy: fn() },
	decorators: [
		Story => (
			<div className="w-96">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof TalentsHeaderView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Unspent: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('copy-button'));
		await expect(args.onCopy).toHaveBeenCalledOnce();
	},
};

export const PartlySpent: Story = { args: { pointsRemaining: 15 } };

export const AllSpent: Story = { args: { pointsRemaining: 0 } };

export const Copied: Story = { args: { pointsRemaining: 0, copied: true } };
