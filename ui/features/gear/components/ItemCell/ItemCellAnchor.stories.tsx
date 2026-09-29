import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { ItemCellAnchor } from './ItemCellAnchor';

const meta = {
	title: 'Features/Gear/ItemCell/Components/ItemCellAnchor',
	component: ItemCellAnchor,
	tags: ['autodocs'],
	args: { children: 'Thunderfury, Blessed Blade of the Windseeker', className: 'text-quality-legendary', role: 'button', onActivate: fn() },
} satisfies Meta<typeof ItemCellAnchor>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Activates: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('button'));
		await expect(args.onActivate).toHaveBeenCalledOnce();
	},
};

export const WowheadLink: Story = { args: { href: 'https://www.wowhead.com/classic/item=19019', role: undefined } };

export const PlainText: Story = { args: { onActivate: undefined, role: undefined } };
