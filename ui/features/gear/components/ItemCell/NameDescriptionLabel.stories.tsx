import type { Meta, StoryObj } from '@storybook/react-vite';

import { NameDescriptionLabel } from './NameDescriptionLabel';

const meta = {
	title: 'Features/Gear/NameDescriptionLabel',
	component: NameDescriptionLabel,
	tags: ['autodocs'],
	args: { nameDescription: 'Heroic' },
	decorators: [
		Story => (
			<span className="text-quality-epic">
				Band of Accuria
				<Story />
			</span>
		),
	],
} satisfies Meta<typeof NameDescriptionLabel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Flush: Story = { args: { flush: true } };
