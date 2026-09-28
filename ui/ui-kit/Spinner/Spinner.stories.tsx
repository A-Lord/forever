import type { Meta, StoryObj } from '@storybook/react-vite';

import { Spinner } from './Spinner';

const meta = {
	title: 'Kit/Spinner',
	component: Spinner,
	tags: ['autodocs'],
	argTypes: {
		size: { control: 'select', options: ['md', 'sm'] },
		className: { control: 'text' },
	},
} satisfies Meta<typeof Spinner>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Medium: Story = { args: { size: 'md' } };
export const Small: Story = { args: { size: 'sm' } };
