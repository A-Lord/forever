import type { Meta, StoryObj } from '@storybook/react-vite';

import { TooltipRow } from './TooltipRow';

const meta = {
	title: 'Features/CharacterStats/TooltipRow',
	component: TooltipRow,
	tags: ['autodocs'],
	args: { label: 'Gear', value: '+180' },
	decorators: [
		Story => (
			<div className="max-w-56">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof TooltipRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Success: Story = { args: { label: 'Crush', value: '103.00%', className: 'text-success' } };
export const Danger: Story = { args: { label: 'Shear', value: '60.70%', className: 'text-danger' } };
