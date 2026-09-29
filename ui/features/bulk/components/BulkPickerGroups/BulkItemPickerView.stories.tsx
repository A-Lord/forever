import { ItemSlot } from '@generated/proto/common';
import { EquippedItem } from '@sim/proto/equipped_item';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { withSimHost } from '@testing/storybook/fixtures/models';
import { makeUIItem } from '@testing/storybook/fixtures/protos';
import { fn } from 'storybook/test';

import { BulkItemPickerView } from './BulkItemPickerView';

const NEEDS_DATABASE = { smoke: false, smokeReason: 'ActionId.fill() needs the item database for the item’s icon and name.' };

const meta = {
	title: 'Features/Bulk/BulkItemPickerView',
	component: BulkItemPickerView,
	tags: ['autodocs'],
	decorators: [withSimHost()],
	parameters: NEEDS_DATABASE,
	args: { slot: ItemSlot.ItemSlotMainHand, item: new EquippedItem({ item: makeUIItem() }), state: 'editable', onOpen: fn(), onRemove: fn() },
	argTypes: { item: { control: false }, state: { control: 'select', options: ['frozen', 'equipped', 'editable'] } },
} satisfies Meta<typeof BulkItemPickerView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Editable: Story = {};

export const Equipped: Story = { args: { state: 'equipped', onRemove: undefined } };

export const Frozen: Story = { args: { state: 'frozen', onRemove: undefined } };
