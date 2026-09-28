import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { AdaptiveStringPicker } from './AdaptiveStringPicker';
import type { AdaptiveStringPickerProps } from './types';

const AdaptiveStringPickerStory = ({ value: initialValue, onChange, ...props }: AdaptiveStringPickerProps) => {
	const [value, setValue] = useState(initialValue);
	return (
		<AdaptiveStringPicker
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
	title: 'Kit/AdaptiveStringPicker',
	component: AdaptiveStringPicker,
	tags: ['autodocs'],
	args: {
		id: 'adaptive-string-picker-story',
		label: 'Variable Name',
		value: 'burst',
		onChange: fn(),
	},
	argTypes: {
		layout: { control: 'select', options: [undefined, 'inline', 'split'] },
		disabled: { control: 'boolean' },
		value: { control: false },
		className: { control: 'text' },
	},
	render: args => <AdaptiveStringPickerStory {...args} />,
} satisfies Meta<typeof AdaptiveStringPicker>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		const input = within(canvasElement).getByRole('textbox');
		await userEvent.clear(input);
		await userEvent.type(input, 'execute');
		await userEvent.tab();
		await expect(args.onChange).toHaveBeenLastCalledWith('execute');
	},
};
export const Long: Story = { args: { value: 'a variable name long enough to widen the field' } };
