import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { AbortButton } from './AbortButton';

const meta = {
	title: 'Features/Results/SimResultsPanel/Components/AbortButton',
	component: AbortButton,
	tags: ['autodocs'],
	args: { onAbort: fn() },
} satisfies Meta<typeof AbortButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Idle: Story = {};

export const Stopping: Story = {
	play: async ({ args, canvasElement }) => {
		const button = within(canvasElement).getByRole('button');
		await userEvent.click(button);
		await expect(args.onAbort).toHaveBeenCalledOnce();
		await expect(button).toBeDisabled();
	},
};
