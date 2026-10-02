import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '@ui-kit/Button';
import { expect, fn, userEvent, within } from 'storybook/test';

import { Drawer } from './Drawer';

const meta = {
	title: 'Kit/Drawer',
	component: Drawer,
	tags: ['autodocs'],
	args: {
		trigger: <Button>Open drawer</Button>,
		onOpenChange: fn(),
		className: 'inset-x-0 bottom-0 border-t border-border bg-background p-4',
		children: <p className="m-0">Drawer contents.</p>,
	},
	argTypes: {
		trigger: { control: false },
		children: { control: false },
		style: { control: false },
		ignoreOutsidePress: { control: false },
	},
} satisfies Meta<typeof Drawer>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const NonModal: Story = {
	args: { modal: false, children: <p className="m-0">The page behind stays interactive.</p> },
};

export const Opens: Story = {
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: 'Open drawer' }));
		const drawer = await canvas.findByRole('dialog');
		await expect(canvasElement.querySelector('[data-testid="storybook-portal"]')?.contains(drawer)).toBe(true);
		await expect(args.onOpenChange).toHaveBeenCalledWith(true);
	},
};
