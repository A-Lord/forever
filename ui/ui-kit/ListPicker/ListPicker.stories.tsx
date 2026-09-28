import type { Meta, StoryObj } from '@storybook/react-vite';
import { AdaptiveStringPicker } from '@ui-kit/AdaptiveStringPicker';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Input } from '@ui-kit/FormControl';
import { NumberPicker } from '@ui-kit/NumberPicker';
import { useId, useState } from 'react';
import { expect, fireEvent, fn, within } from 'storybook/test';

import { ListPicker } from './ListPicker';
import type { ListPickerProps } from './types';

interface Target {
	name: string;
}

const targets = (...names: Array<string>): Array<Target> => names.map(name => ({ name }));

type ListPickerStoryProps = Omit<ListPickerProps<Target>, 'items' | 'renderItem' | 'renderItemHeader'> & {
	initial: Array<Target>;
	withHeader?: boolean;
};

const ListPickerStory = ({ initial, onChange, withHeader, ...props }: ListPickerStoryProps) => {
	const [items, setItems] = useState(initial);
	const listProps = {
		...props,
		items,
		onChange: (next: Array<Target>) => {
			setItems(next);
			onChange(next);
		},
		renderItem: (item: Target, _index: number, onItemChange: (next: Target) => void) => (
			<Input className="w-full" data-testid="target-name" value={item.name} onChange={event => onItemChange({ name: event.target.value })} />
		),
		renderItemHeader: withHeader ? (item: Target) => <span className="text-sm">{item.name.length} letters</span> : undefined,
	} as ListPickerProps<Target>;
	return <ListPicker<Target> {...listProps} />;
};

const meta = {
	title: 'Kit/ListPicker',
	component: ListPickerStory,
	tags: ['autodocs'],
	args: {
		itemLabel: 'Target',
		initial: targets('Patchwerk', 'Grobbulus'),
		newItem: () => ({ name: 'New target' }),
		copyItem: (old: Target) => ({ ...old }),
		onChange: fn(),
	},
	argTypes: {
		disabled: { control: 'boolean' },
		initial: { control: false },
		className: { control: 'text' },
	},
} satisfies Meta<typeof ListPickerStory>;

export default meta;
type Story = StoryObj<typeof meta>;

const names = (element: HTMLElement) =>
	within(element)
		.queryAllByTestId('target-name')
		.map(input => (input as HTMLInputElement).value);

export const Default: Story = {
	args: { title: 'Targets', titleTooltip: 'One row per target.' },
	play: async ({ args, canvasElement }) => {
		await fireEvent.click(within(canvasElement).getByTestId('list-picker-new-button'));
		await expect(args.onChange).toHaveBeenLastCalledWith(targets('Patchwerk', 'Grobbulus', 'New target'));
		await expect(names(canvasElement)).toEqual(['Patchwerk', 'Grobbulus', 'New target']);
	},
};

export const Compact: Story = { args: { isCompact: true, allowedActions: [] } };

export const Horizontal: Story = { args: { horizontalLayout: true, actions: { create: { useIcon: true } } } };

export const InlineMenuBar: Story = { args: { inlineMenuBar: true, minimumItems: 1 } };

export const WithHeader: Story = { args: { withHeader: true } };

interface Spell {
	name: string;
	delay: number;
	enabled: boolean;
}

const DensityRows = ({ density }: { density?: 'compact' }) => {
	const rowsId = useId();
	const [items, setItems] = useState<Array<Spell>>([
		{ name: 'Fireball', delay: 0.5, enabled: true },
		{ name: 'Frostbolt', delay: 1, enabled: false },
	]);
	return (
		<ListPicker<Spell>
			items={items}
			onChange={setItems}
			density={density}
			itemLabel="Spell"
			inlineMenuBar
			actions={{ create: { useIcon: true } }}
			newItem={() => ({ name: 'Arcane Blast', delay: 0, enabled: true })}
			renderItem={(item, index, onItemChange) => (
				<div className="flex flex-row flex-wrap items-center gap-2">
					<AdaptiveStringPicker
						id={`${rowsId}-name-${index}`}
						label="Name"
						layout="inline"
						value={item.name}
						onChange={name => onItemChange({ ...item, name })}
					/>
					<NumberPicker
						id={`${rowsId}-delay-${index}`}
						label="Delay"
						layout="inline"
						float
						value={item.delay}
						onChange={delay => onItemChange({ ...item, delay })}
					/>
					<BooleanPicker
						id={`${rowsId}-enabled-${index}`}
						label="Enabled"
						value={item.enabled}
						onChange={enabled => onItemChange({ ...item, enabled })}
					/>
				</div>
			)}
		/>
	);
};

/** The same rows at the default density and at `density="compact"`, the layout the rotation editor uses. */
export const Density: Story = {
	render: () => (
		<div className="flex flex-col gap-6">
			<DensityRows />
			<DensityRows density="compact" />
		</div>
	),
};

// A synthetic drag brings no data transfer of its own; the list only ever reads these three members.
const dataTransfer = () => ({ setDragImage: () => {}, dropEffect: '', effectAllowed: '' });

export const DragBetweenLists: Story = {
	args: { dragGroup: 'targets', sameGroupOnly: true, inlineMenuBar: true },
	render: args => (
		<div className="flex flex-col gap-2">
			<div data-testid="left-list">
				<ListPickerStory {...args} />
			</div>
			<div data-testid="right-list">
				<ListPickerStory {...args} initial={targets('Thaddius')} />
			</div>
		</div>
	),
	play: async ({ canvasElement }) => {
		const left = within(canvasElement).getByTestId('left-list');
		const right = within(canvasElement).getByTestId('right-list');
		const source = within(left).getAllByTestId('list-picker-item-container')[0];
		const target = within(right).getAllByTestId('list-picker-item-container')[0];

		await fireEvent.mouseDown(source);
		await fireEvent.dragStart(source, { dataTransfer: dataTransfer(), clientY: 0 });
		await fireEvent.drop(target, { dataTransfer: dataTransfer(), clientY: 0 });

		await expect(names(left)).toEqual(['Grobbulus']);
		await expect(names(right)).toEqual(['Patchwerk', 'Thaddius']);
	},
};
