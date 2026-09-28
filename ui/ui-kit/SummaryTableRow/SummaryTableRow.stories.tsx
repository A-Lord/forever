import type { Meta, StoryObj } from '@storybook/react-vite';

import { SummaryTableRow } from './SummaryTableRow';

const meta = {
	title: 'Kit/SummaryTableRow',
	component: SummaryTableRow,
	tags: ['autodocs'],
	argTypes: { children: { control: false } },
	render: args => (
		<div className="flex w-72 flex-col gap-1">
			<SummaryTableRow {...args}>
				<span className="text-quality-rare">Runed Crimson Spinel</span>
				<span>4</span>
			</SummaryTableRow>
			<SummaryTableRow {...args}>
				<span className="text-quality-rare">Brilliant Lionseye</span>
				<span>2</span>
			</SummaryTableRow>
		</div>
	),
} satisfies Meta<typeof SummaryTableRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
