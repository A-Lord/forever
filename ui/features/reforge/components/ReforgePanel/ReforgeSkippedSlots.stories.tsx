import type { Meta, StoryObj } from '@storybook/react-vite';

import { ReforgeSkippedSlotsView } from './ReforgeSkippedSlots';

const meta = {
	title: 'Features/Reforge/ReforgePanel/Components/ReforgeSkippedSlots',
	component: ReforgeSkippedSlotsView,
	tags: ['autodocs'],
	args: {
		slots: [
			{ slot: 'Main Hand', enchant: 'Enchant Weapon - Crusader' },
			{ slot: 'Chest', enchant: 'Enchant Chest - Lesser Absorption' },
			{ slot: 'Head' },
		],
	},
	decorators: [
		Story => (
			<div className="max-w-86">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof ReforgeSkippedSlotsView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
