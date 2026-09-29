import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, waitFor, within } from 'storybook/test';

import { TooltipButton } from './TooltipButton';

const meta = {
	title: 'Kit/TooltipButton',
	component: TooltipButton,
	tags: ['autodocs'],
	args: { tooltip: 'Explains the setting next to it.' },
	argTypes: {
		place: { control: 'select', options: [undefined, 'top', 'right', 'bottom', 'left'] },
		iconStyle: { control: 'select', options: ['regular', 'solid'] },
		tooltip: { control: 'text' },
		className: { control: 'text' },
	},
} satisfies Meta<typeof TooltipButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const OtherGlyph: Story = { args: { icon: 'triangle-exclamation', iconStyle: 'solid', tooltip: 'Some items are missing sockets.' } };
export const InHeading: Story = {
	render: args => (
		<h5 className="flex items-center">
			Consumables
			<TooltipButton {...args} className="ml-2" />
		</h5>
	),
};
export const ShowsTooltip: Story = {
	play: async ({ canvasElement }) => {
		await userEvent.hover(within(canvasElement).getByTestId('tooltip-button'));
		const tooltip = await within(document.body).findByText('Explains the setting next to it.');
		await waitFor(() => expect(tooltip).toBeVisible());
	},
};
