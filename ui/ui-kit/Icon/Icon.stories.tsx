import type { Meta, StoryObj } from '@storybook/react-vite';

import { Icon } from './Icon';
import { ICON_ALIASES, type IconAlias, type IconSize, type IconStyle } from './types';

const STYLES: Array<{ style: IconStyle; name: 'circle-question' | 'github' }> = [
	{ style: 'solid', name: 'circle-question' },
	{ style: 'regular', name: 'circle-question' },
	{ style: 'brands', name: 'github' },
	{ style: 'base', name: 'circle-question' },
];

const SIZES: IconSize[] = ['2xs', 'xs', 'sm', '1x', 'lg', 'xl', '2xl', '2x', '3x', '4x', '5x'];

const meta = {
	title: 'Kit/Icon',
	component: Icon,
	tags: ['autodocs'],
	args: { name: 'star' },
	argTypes: {
		style: { control: 'select', options: ['solid', 'regular', 'brands', 'base'] },
		size: { control: 'select', options: [undefined, ...SIZES] },
		spin: { control: 'boolean' },
		title: { control: 'text' },
		className: { control: 'text' },
	},
} satisfies Meta<typeof Icon>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Styles: Story = {
	render: () => (
		<div className="flex items-end gap-6">
			{STYLES.map(({ style, name }) => (
				<div key={style} className="flex flex-col items-center gap-1">
					<Icon name={name} style={style} size="2x" />
					<code className="text-ui">{style}</code>
				</div>
			))}
		</div>
	),
};

export const Sizes: Story = {
	render: args => (
		<div className="flex flex-wrap items-end gap-6">
			{SIZES.map(size => (
				<div key={size} className="flex flex-col items-center gap-1">
					<Icon {...args} size={size} />
					<code className="text-ui">{size}</code>
				</div>
			))}
		</div>
	),
};

export const Spin: Story = { args: { name: 'spinner', size: '2x', spin: true } };

export const Aliases: Story = {
	render: () => (
		<div className="grid grid-cols-[auto_auto] items-center gap-x-6 gap-y-2">
			{Object.entries(ICON_ALIASES).map(([alias, name]) => (
				<div key={alias} className="contents">
					<code className="text-ui">
						{alias} → {name}
					</code>
					<Icon name={alias as IconAlias} size="lg" />
				</div>
			))}
		</div>
	),
};
