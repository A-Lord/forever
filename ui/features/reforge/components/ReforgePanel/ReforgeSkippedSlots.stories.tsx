import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { ReforgeSkippedSlotsView } from './ReforgeSkippedSlots';

const meta = {
	title: 'Features/Reforge/ReforgePanel/Components/ReforgeSkippedSlots',
	component: ReforgeSkippedSlotsView,
	tags: ['autodocs'],
	args: {
		procSlots: ['Main Hand', 'Chest'],
		unlistedSlots: ['Head'],
		onOpenBatchSim: fn(),
	},
	decorators: [
		Story => (
			<div className="max-w-86">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof ReforgeSkippedSlotsView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('reforge-open-batch-sim'));
		await expect(args.onOpenBatchSim).toHaveBeenCalledOnce();
	},
};

/** In the Batch Sim tab's own copy of the panel there is no tab to open. */
export const WithoutBatchSimLink: Story = {
	args: { onOpenBatchSim: undefined },
};
