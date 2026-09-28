import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { fn } from 'storybook/test';

import { NumberPicker, type NumberPickerProps } from './NumberPicker';

const NumberPickerStory = ({ value: initialValue, onChange, ...props }: NumberPickerProps) => {
	const [value, setValue] = useState(initialValue);
	return (
		<NumberPicker
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
	title: 'Kit/NumberPicker',
	component: NumberPicker,
	tags: ['autodocs'],
	args: {
		id: 'number-picker-story',
		label: 'Iterations',
		value: 12500,
		onChange: fn(),
	},
	argTypes: {
		layout: { control: 'select', options: [undefined, 'inline', 'split'] },
		disabled: { control: 'boolean' },
		float: { control: 'boolean' },
		positive: { control: 'boolean' },
		showZeroes: { control: 'boolean' },
		value: { control: false },
		className: { control: 'text' },
	},
	render: args => <NumberPickerStory {...args} />,
} satisfies Meta<typeof NumberPicker>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Float: Story = { args: { label: 'Swing Speed', value: 2.6, float: true } };
export const Positive: Story = { args: { label: 'Distance From Target', value: 5, float: true, positive: true } };
export const HideZeroes: Story = { args: { label: 'Min ILvl', value: 0, showZeroes: false } };
export const Split: Story = { args: { label: 'Reaction Time', value: 150, layout: 'split' } };
export const Disabled: Story = { args: { disabled: true } };
