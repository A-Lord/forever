import type { Meta, StoryObj } from '@storybook/react-vite';
import { Field } from '@ui-kit/Field';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import { Input } from '@ui-kit/FormControl';
import { IconPicker } from '@ui-kit/IconPicker';
import { fn } from 'storybook/test';

import { PickerGroup } from './PickerGroup';

const ICONS = 'https://wow.zamimg.com/images/wow/icons/large';
const icon = (file: string, name: string, itemId: number): ResolvedIcon => ({
	iconUrl: `${ICONS}/${file}.jpg`,
	name,
	href: `https://www.wowhead.com/classic/item=${itemId}`,
});

const CONSUMES = [
	icon('inv_potion_62', 'Flask of the Titans', 13510),
	icon('inv_potion_97', 'Elixir of the Mongoose', 13452),
	icon('inv_misc_food_64', 'Smoked Desert Dumplings', 20452),
	icon('inv_stone_02', 'Dense Sharpening Stone', 12404),
];

const FIELDS = [
	{ id: 'picker-group-story-duration', label: 'Duration', value: '180' },
	{ id: 'picker-group-story-variation', label: 'Duration +/-', value: '5' },
	{ id: 'picker-group-story-execute', label: 'Execute Phase %', value: '20' },
];

const meta = {
	title: 'Kit/PickerGroup',
	component: PickerGroup,
	tags: ['autodocs'],
	argTypes: {
		variant: { control: 'select', options: [undefined, 'icons'] },
		className: { control: 'text' },
		children: { control: false },
	},
} satisfies Meta<typeof PickerGroup>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	args: {
		children: FIELDS.map(({ id, label, value }) => (
			<Field key={id} id={id} label={label}>
				<Input id={id} defaultValue={value} />
			</Field>
		)),
	},
};
export const Icons: Story = {
	args: {
		variant: 'icons',
		children: CONSUMES.map((consume, index) => <IconPicker key={consume.name} states={2} icon={consume} value={index % 2} onChange={fn()} />),
	},
};
