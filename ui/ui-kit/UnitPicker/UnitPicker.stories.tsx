import { UnitReference, UnitReference_Type as UnitType } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import type { UnitValue } from './types';
import { UnitPicker, type UnitPickerProps } from './UnitPicker';

const ICONS = 'https://wow.zamimg.com/images/wow/icons';
const target = (index: number) => UnitReference.create({ type: UnitType.Target, index });

const OPTIONS: Array<UnitValue> = [
	{ value: undefined, text: 'Self', iconUrl: 'fa-user' },
	{ value: UnitReference.create({ type: UnitType.CurrentTarget }), text: 'Current Target', iconUrl: 'fa-bullseye' },
	{ value: target(0), text: 'Target 1', iconUrl: `${ICONS}/large/spell_shadow_metamorphosis.jpg`, color: 'warrior' },
	{ value: target(1), text: 'Target 2', iconUrl: `${ICONS}/large/spell_shadow_metamorphosis.jpg` },
	{ value: UnitReference.create({ type: UnitType.Pet, index: 0 }), text: 'Wolf', iconUrl: `${ICONS}/medium/ability_hunter_pet_wolf.jpg` },
];

const UnitPickerStory = ({ value: initialValue, onChange, ...props }: UnitPickerProps) => {
	const [value, setValue] = useState(initialValue);
	return (
		<UnitPicker
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
	title: 'Kit/UnitPicker',
	component: UnitPicker,
	tags: ['autodocs'],
	args: {
		options: OPTIONS,
		value: target(0),
		onChange: fn(),
	},
	argTypes: {
		hideLabelWhenDefault: { control: 'boolean' },
		options: { control: false },
		value: { control: false },
		className: { control: 'text' },
	},
	render: args => <UnitPickerStory {...args} />,
} satisfies Meta<typeof UnitPicker>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ImageIcon: Story = {};
export const GlyphIcon: Story = { args: { value: UnitReference.create({ type: UnitType.CurrentTarget }) } };
export const DefaultUnitIconOnly: Story = { args: { value: undefined, hideLabelWhenDefault: true } };
export const NoMatch: Story = { args: { value: target(9) } };
export const Picks: Story = {
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByTestId('dropdown-picker-button'));
		const items = await canvas.findAllByTestId('dropdown-picker-item');
		await expect(within(items[4]).getByTestId('unit-picker-item-icon').getAttribute('src')).toBe(`${ICONS}/medium/ability_hunter_pet_wolf.jpg`);
		await userEvent.click(items[4]);
		await expect(args.onChange).toHaveBeenCalledWith(OPTIONS[4].value);
		await expect(canvas.getByTestId('dropdown-picker-button').textContent).toBe('Wolf');
	},
};
