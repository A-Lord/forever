import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '@ui-kit/Button';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { Dialog, type DialogProps, type DialogSize } from './Dialog';

const SIZES: Array<DialogSize> = ['sm', 'md', 'lg', 'xl'];

interface DialogStoryProps extends DialogProps {
	trigger?: string;
}

const DialogStory = ({ trigger = 'Open dialog', open: initialOpen, onOpenChange, ...props }: DialogStoryProps) => {
	const [open, setOpen] = useState(initialOpen);
	return (
		<>
			<Button onClick={() => setOpen(true)}>{trigger}</Button>
			<Dialog
				{...props}
				open={open}
				onOpenChange={next => {
					setOpen(next);
					onOpenChange(next);
				}}
			/>
		</>
	);
};

const meta = {
	title: 'Kit/Dialog',
	component: Dialog,
	args: {
		open: false,
		onOpenChange: fn(),
		title: 'Options',
		children: <p>Dialog body.</p>,
	},
	argTypes: {
		size: { control: 'select', options: SIZES },
		children: { control: false },
		footer: { control: false },
		headerChildren: { control: false },
		container: { control: false },
	},
	render: args => <DialogStory {...args} />,
} satisfies Meta<typeof Dialog>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Sizes: Story = {
	render: args => (
		<div className="flex gap-2">
			{SIZES.map(size => (
				<DialogStory key={size} {...args} size={size} title={`${size} dialog`} trigger={size} />
			))}
		</div>
	),
};

export const WithFooter: Story = {
	args: {
		footer: (
			<>
				<Button variant="cancel">Cancel</Button>
				<Button>Save</Button>
			</>
		),
	},
};

export const Elevated: Story = {
	args: {
		children: <DialogStory open={false} onOpenChange={fn()} elevated title="Elevated" trigger="Open elevated dialog" />,
	},
};

export const Opens: Story = {
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: 'Open dialog' }));
		const dialog = await canvas.findByRole('dialog');
		await expect(canvasElement.querySelector('[data-testid="storybook-portal"]')?.contains(dialog)).toBe(true);
	},
};
