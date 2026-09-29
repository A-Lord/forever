import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '@ui-kit/Button';
import { expect, fn, userEvent, within } from 'storybook/test';

import { Drawer } from './Drawer';

const SHEET = 'inset-x-0 bottom-0 max-h-[50vh] overflow-y-auto border-t border-border bg-background p-4';

const meta = {
	title: 'Kit/Drawer',
	component: Drawer,
	tags: ['autodocs'],
	args: {
		trigger: <Button>Open drawer</Button>,
		onOpenChange: fn(),
		className: SHEET,
		children: <p className="m-0">Drawer contents.</p>,
	},
	argTypes: {
		trigger: { control: false },
		children: { control: false },
		headerChildren: { control: false },
		title: { control: 'text' },
		style: { control: false },
		ignoreOutsidePress: { control: false },
	},
} satisfies Meta<typeof Drawer>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const WithHeader: Story = {
	args: {
		title: 'Filters',
		headerChildren: <span className="ml-auto text-sm text-border">3 active</span>,
	},
};

export const NonModal: Story = {
	args: { modal: false, children: <p className="m-0">The page behind stays interactive.</p> },
};

export const Bottom: Story = {
	args: {
		side: 'bottom',
		className: undefined,
		title: 'Stats',
		children: Array.from({ length: 40 }, (_, index) => (
			<p key={index} className="m-0 py-1">
				Row {index + 1}
			</p>
		)),
	},
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: 'Open drawer' }));
		const drawer = await canvas.findByRole('dialog', { name: 'Stats' });
		await expect(drawer).toHaveAttribute('data-side', 'bottom');
		await expect(canvas.getByTestId('sim-drawer-backdrop')).toBeInTheDocument();
		await expect(args.onOpenChange).toHaveBeenCalledWith(true);
	},
};

export const Opens: Story = {
	args: { title: 'Filters' },
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: 'Open drawer' }));
		const drawer = await canvas.findByRole('dialog', { name: 'Filters' });
		await expect(canvasElement.querySelector('[data-testid="storybook-portal"]')?.contains(drawer)).toBe(true);
		await expect(args.onOpenChange).toHaveBeenCalledWith(true);
	},
};
