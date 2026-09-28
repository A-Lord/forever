import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { DropdownPicker } from './DropdownPicker';
import type { DropdownOption, DropdownPickerProps } from './types';

const ICONS = 'https://wow.zamimg.com/images/wow/icons/medium';
const icon = (name: string) => <img className="ui-unit-icon" src={`${ICONS}/${name}.jpg`} alt="" />;

const KINDS: Array<DropdownOption<string>> = [
	{ value: 'cast', label: 'Cast', tooltip: 'Cast a spell.' },
	{ value: 'wait', label: 'Wait', tooltip: 'Wait for a duration.' },
	{ value: 'sequence', label: 'Sequence', tooltip: 'Run a list of actions in order.' },
];

const SPELLS: Array<DropdownOption<string>> = [
	{ value: '', label: 'None' },
	{ value: 'fireball', label: 'Fireball', icon: icon('spell_fire_flamebolt') },
	{ value: 'frostbolt', label: 'Frostbolt', icon: icon('spell_frost_frostbolt02') },
	{ value: 'arcane-missiles', label: 'Arcane Missiles', icon: icon('spell_nature_starfall') },
];

const GROUPED: Array<DropdownOption<string>> = [
	{ value: 'none', label: 'None' },
	{ value: 'and', label: 'All of', submenu: ['Logic'] },
	{ value: 'or', label: 'Any of', submenu: ['Logic'] },
	{ value: 'mana', label: 'Current Mana', submenu: ['Resources'] },
	{ value: 'mana-percent', label: 'Current Mana %', submenu: ['Resources'] },
];

const DropdownPickerStory = ({ value: initialValue, onChange, ...props }: DropdownPickerProps<string>) => {
	const [value, setValue] = useState(initialValue);
	return (
		<DropdownPicker<string>
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
	title: 'Kit/DropdownPicker',
	component: DropdownPicker<string>,
	tags: ['autodocs'],
	args: {
		id: 'dropdown-picker-story',
		options: KINDS,
		value: 'cast',
		onChange: fn(),
		equals: (a: string | undefined, b: string | undefined) => a === b,
		defaultLabel: 'Pick an action',
	},
	argTypes: {
		layout: { control: 'select', options: [undefined, 'inline', 'split'] },
		side: { control: 'select', options: [undefined, 'top', 'bottom', 'left', 'right'] },
		disabled: { control: 'boolean' },
		options: { control: false },
		value: { control: false },
		equals: { control: false },
		hideLabelWhenDefault: { control: false },
		className: { control: 'text' },
	},
	render: args => <DropdownPickerStory {...args} />,
} satisfies Meta<typeof DropdownPicker<string>>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByTestId('dropdown-picker-button'));
		const items = await within(canvasElement.ownerDocument.body).findAllByTestId('dropdown-picker-item');
		await userEvent.click(items[1]);
		await expect(args.onChange).toHaveBeenCalledWith('wait');
		await expect(canvas.getByTestId('dropdown-picker-button').textContent).toBe('Wait');
	},
};
export const WithIcons: Story = { args: { options: SPELLS, value: 'frostbolt', defaultLabel: 'Pick a spell' } };
export const WithSubmenu: Story = { args: { options: GROUPED, value: 'mana', defaultLabel: 'Pick a value' } };
export const HideLabelWhenDefault: Story = {
	args: { options: SPELLS, value: '', defaultLabel: 'Pick a spell', hideLabelWhenDefault: (value: string) => !value },
};
export const WithChrome: Story = {
	args: { label: 'Action', labelTooltip: 'What this entry of the priority list does.', description: 'Picked from the action kinds.', layout: 'inline' },
};
