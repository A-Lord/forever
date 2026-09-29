import { ItemSlot } from '@generated/proto/common';
import { Gear } from '@sim/proto/gear';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { withSimHost } from '@testing/storybook/fixtures/models';
import { expect, fireEvent, fn, within } from 'storybook/test';

import { SlotRail } from './SlotRail';

const meta = {
	title: 'Features/Gear/SlotRail',
	component: SlotRail,
	tags: ['autodocs'],
	args: { gear: new Gear({}), currentSlot: ItemSlot.ItemSlotHead, onOpen: fn() },
	argTypes: { gear: { control: false } },
	decorators: [
		Story => (
			<div className="relative ml-24">
				<Story />
			</div>
		),
		withSimHost(),
	],
} satisfies Meta<typeof SlotRail>;

export default meta;
type Story = StoryObj<typeof meta>;

export const EmptyGear: Story = {
	play: async ({ args, canvasElement }) => {
		// A click without the hover: the rail's tooltip opening on hover would land after the story ends.
		fireEvent.click(within(canvasElement).getAllByTestId('item-picker-icon')[2]);
		await expect(args.onOpen).toHaveBeenCalledWith(ItemSlot.ItemSlotShoulder);
	},
};

export const OnTheMainHand: Story = { args: { currentSlot: ItemSlot.ItemSlotMainHand } };
