import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { Field } from './Field';

const meta = {
	title: 'Kit/Field',
	component: Field,
	tags: ['autodocs'],
	args: {
		id: 'field-story',
		label: 'Cast Delay',
		children: <input id="field-story" className="ui-input" defaultValue="0.1" />,
	},
	argTypes: {
		layout: { control: 'select', options: [undefined, 'inline', 'split'] },
		disabled: { control: 'boolean' },
		iconField: { control: 'boolean' },
		grow: { control: 'boolean' },
		className: { control: 'text' },
		children: { control: false },
		leading: { control: false },
		after: { control: false },
		clear: { control: false },
		ref: { control: false },
	},
} satisfies Meta<typeof Field>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Inline: Story = { args: { layout: 'inline' } };
export const Split: Story = { args: { layout: 'split' } };
export const WithTooltip: Story = { args: { labelTooltip: 'Seconds of <b>human</b> reaction time before each cast.' } };
export const WithDescription: Story = { args: { description: 'Delay before the first cast.' } };
export const IconField: Story = {
	args: { iconField: true, className: 'ui-icon-field', leading: <span className="ui-icon-picker-swatch" />, children: undefined },
};
export const WithClear: Story = {
	args: {
		label: 'Search',
		children: <input id="field-story" className="ui-input" defaultValue="Thunderfury" />,
		clear: { show: true, label: 'Clear', onClear: fn() },
	},
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'Clear' }));
		await expect(args.clear?.onClear).toHaveBeenCalledOnce();
	},
};
export const Disabled: Story = { args: { disabled: true } };
