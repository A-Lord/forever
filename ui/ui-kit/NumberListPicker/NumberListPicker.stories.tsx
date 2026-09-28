import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { NumberListPicker } from './NumberListPicker';
import type { NumberListPickerProps } from './types';

const NumberListPickerStory = ({ value: initialValue, onChange, ...props }: NumberListPickerProps) => {
	const [value, setValue] = useState(initialValue);
	return (
		<NumberListPicker
			{...props}
			value={value}
			onChange={next => {
				setValue(next);
				onChange(next);
			}}
		/>
	);
};

const meta = {
	title: 'Kit/NumberListPicker',
	component: NumberListPicker,
	tags: ['autodocs'],
	args: {
		id: 'number-list-picker-story',
		label: 'Timings',
		value: [0, 90.5, 180],
		onChange: fn(),
	},
	argTypes: {
		layout: { control: 'select', options: [undefined, 'inline', 'split'] },
		disabled: { control: 'boolean' },
		value: { control: false },
		className: { control: 'text' },
	},
	render: args => <NumberListPickerStory {...args} />,
} satisfies Meta<typeof NumberListPicker>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		const input = within(canvasElement).getByRole('textbox');
		await userEvent.clear(input);
		await userEvent.type(input, '30, x, 60,');
		await userEvent.tab();
		await expect(args.onChange).toHaveBeenLastCalledWith([30, 60]);
		await expect(input).toHaveValue('30,60');
	},
};
export const Placeholder: Story = { args: { value: [], placeholder: 'e.g. 30,120' } };
