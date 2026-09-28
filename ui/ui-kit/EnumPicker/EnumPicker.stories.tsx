import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { EnumPicker, type EnumPickerProps } from './EnumPicker';

const EnumPickerStory = ({ value: initialValue, onChange, ...props }: EnumPickerProps) => {
	const [value, setValue] = useState(initialValue);
	return (
		<EnumPicker
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
	title: 'Kit/EnumPicker',
	component: EnumPicker,
	tags: ['autodocs'],
	args: {
		id: 'enum-picker-story',
		label: 'Mob Type',
		values: [
			{ name: 'Unknown', value: 0 },
			{ name: 'Beast', value: 1 },
			{ name: 'Demon', value: 2, tooltip: 'Takes extra damage from demon slaying effects.' },
			{ name: 'Undead', value: 8 },
		],
		value: 2,
		onChange: fn(),
	},
	argTypes: {
		layout: { control: 'select', options: [undefined, 'inline', 'split'] },
		disabled: { control: 'boolean' },
		values: { control: false },
		value: { control: false },
		className: { control: 'text' },
	},
	render: args => <EnumPickerStory {...args} />,
} satisfies Meta<typeof EnumPicker>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.selectOptions(within(canvasElement).getByRole('combobox'), '8');
		await expect(args.onChange).toHaveBeenLastCalledWith(8);
	},
};
export const Disabled: Story = { args: { disabled: true } };
