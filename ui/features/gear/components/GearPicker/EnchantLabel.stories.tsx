import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeUIEnchant } from '@testing/storybook/fixtures/protos';
import { fn } from 'storybook/test';

import { EnchantLabel } from './EnchantLabel';

const meta = {
	title: 'Features/Gear/GearPicker/Components/EnchantLabel',
	component: EnchantLabel,
	tags: ['autodocs'],
	args: { className: 'ui-item-picker-label-muted', enchant: makeUIEnchant(), onActivate: fn() },
	argTypes: { enchant: { control: false } },
} satisfies Meta<typeof EnchantLabel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
