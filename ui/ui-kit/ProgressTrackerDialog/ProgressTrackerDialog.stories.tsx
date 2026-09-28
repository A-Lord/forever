import type { Meta, StoryObj } from '@storybook/react-vite';
import { useEffect, useRef } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { ProgressTrackerDialog, type ProgressTrackerDialogProps } from './ProgressTrackerDialog';
import type { ProgressTrackerHandle } from './types';

const WithProgressStory = (props: ProgressTrackerDialogProps) => {
	const handle = useRef<ProgressTrackerHandle>(null);
	useEffect(() => handle.current?.setProgress({ title: '3 / 12 simulations complete', current: 30, total: 120 }), []);
	return <ProgressTrackerDialog {...props} ref={handle} />;
};

const meta = {
	title: 'Kit/ProgressTrackerDialog',
	component: ProgressTrackerDialog,
	tags: ['autodocs'],
	args: {
		open: true,
		title: 'Calculate Stat Weights',
		state: { stage: 'running' },
	},
	argTypes: {
		state: { control: false },
		warning: { control: false },
		ref: { control: false },
	},
} satisfies Meta<typeof ProgressTrackerDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Running: Story = {
	args: { state: { stage: 'running', message: 'Preparing the batch.' } },
};

export const WithProgress: Story = {
	args: { hasProgressBar: true },
	render: args => <WithProgressStory {...args} />,
};

export const WithWarning: Story = {
	args: {
		title: 'Optimizing Gems',
		warning: <p className="m-0">Gemming can take a while. You may cancel at any time.</p>,
	},
};

export const Cancelable: Story = {
	args: { onCancel: fn() },
	play: async ({ args, canvasElement }) => {
		const dialog = await within(canvasElement).findByRole('dialog', { name: 'Calculate Stat Weights' });
		await expect(canvasElement.querySelector('[data-testid="storybook-portal"]')?.contains(dialog)).toBe(true);
		await userEvent.click(within(dialog).getByTestId('progress-tracker-modal-cancel-btn'));
		await expect(args.onCancel).toHaveBeenCalledTimes(1);
	},
};
