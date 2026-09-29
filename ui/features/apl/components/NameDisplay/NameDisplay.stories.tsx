import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { NameDisplay } from './NameDisplay';

const meta = {
	title: 'Features/Apl/NameDisplay',
	component: NameDisplay,
	tags: ['autodocs'],
	args: { name: 'Execute Phase', onRename: fn() },
} satisfies Meta<typeof NameDisplay>;

export default meta;
type Story = StoryObj<typeof meta>;

/** An action group's or a value variable's name. */
export const Default: Story = {};

/** A variable placeholder's name, inside a field. */
export const Compact: Story = { args: { name: 'target', compact: true } };
