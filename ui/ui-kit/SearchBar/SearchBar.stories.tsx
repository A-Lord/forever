import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { expect, fn, userEvent, waitFor, within } from 'storybook/test';

import { SearchBar, type SearchBarProps } from './SearchBar';

const SearchBarStory = ({ value: initialValue, onChange, ...props }: SearchBarProps) => {
	const [value, setValue] = useState(initialValue);
	return (
		<SearchBar
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
	title: 'Kit/SearchBar',
	component: SearchBar,
	tags: ['autodocs'],
	args: {
		id: 'search-bar-story',
		label: 'Search',
		placeholder: 'Search items',
		value: '',
		onChange: fn(),
	},
	argTypes: {
		value: { control: false },
		children: { control: false },
		debounceMs: { control: 'number' },
		clearable: { control: 'boolean' },
		grow: { control: 'boolean' },
		className: { control: 'text' },
	},
	render: args => <SearchBarStory {...args} />,
} satisfies Meta<typeof SearchBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Clearable: Story = {
	args: { value: 'Thunderfury', clearable: true, clearLabel: 'Clear' },
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: 'Clear' }));
		await expect(args.onChange).toHaveBeenLastCalledWith('');
		await expect(canvas.getByRole('textbox')).toHaveValue('');
	},
};
export const Debounced: Story = {
	args: { debounceMs: 150 },
	play: async ({ args, canvasElement }) => {
		await userEvent.type(within(canvasElement).getByRole('textbox'), 'Ashkandi');
		await expect(args.onChange).not.toHaveBeenCalled();
		await waitFor(() => expect(args.onChange).toHaveBeenCalledWith('Ashkandi'));
		await expect(args.onChange).toHaveBeenCalledOnce();
	},
};
