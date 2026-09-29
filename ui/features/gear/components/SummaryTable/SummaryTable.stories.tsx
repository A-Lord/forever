import type { Meta, StoryObj } from '@storybook/react-vite';
import { SummaryTableRow } from '@ui-kit/SummaryTableRow';
import { expect, fn, userEvent, within } from 'storybook/test';

import { SummaryTable } from './SummaryTable';

const meta = {
	title: 'Features/Gear/SummaryTable/SummaryTable',
	component: SummaryTable,
	tags: ['autodocs'],
	args: {
		title: 'Gem Summary',
		modifier: 'summary-table--gems',
		empty: false,
		reset: { label: 'Reset Gems', onReset: fn() },
		children: (
			<>
				<SummaryTableRow>
					<span>Blood Ruby</span>
					<span>4</span>
				</SummaryTableRow>
				<SummaryTableRow>
					<span>Citrine</span>
					<span>2</span>
				</SummaryTableRow>
			</>
		),
	},
	argTypes: { children: { control: false } },
	decorators: [
		Story => (
			<div className="w-80">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof SummaryTable>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('summary-table-reset-button'));
		await expect(args.reset.onReset).toHaveBeenCalledOnce();
	},
};

export const Empty: Story = {
	args: { empty: true },
	parameters: { smoke: false, smokeReason: 'An empty summary renders nothing, so there is nothing to mount' },
};
