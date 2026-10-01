import { GemColor, ItemSlot } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { wowheadIconUrl } from '@testing/storybook/fixtures/gear';

import { GearChangeIconView } from './GearChangeIconView';

const meta = {
	title: 'Features/Gear/GearChangeIconView',
	component: GearChangeIconView,
	tags: ['autodocs'],
	args: {
		slot: ItemSlot.ItemSlotHead,
		equipped: true,
		iconUrl: wowheadIconUrl('inv_helmet_36'),
		href: '',
		sockets: [],
	},
	argTypes: { sockets: { control: false } },
} satisfies Meta<typeof GearChangeIconView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Unchanged: Story = {};

export const EmptySlot: Story = { args: { equipped: false, iconUrl: '' } };

export const GemChanged: Story = {
	args: {
		sockets: [
			{ socketColor: GemColor.GemColorRed, gemName: 'Bold Ruby', changed: false },
			{ socketColor: GemColor.GemColorYellow, gemName: 'Rigid Citrine', changed: true },
		],
	},
};
