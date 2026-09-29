import type { Meta, StoryObj } from '@storybook/react-vite';
import type { ReactNode } from 'react';
import { expect, fn, screen, userEvent, within } from 'storybook/test';

import { Tree } from './index';
import type { TreeDensity, TreeLayout, TreeOrientation } from './types';

const DENSITIES: Array<TreeDensity> = ['default', 'compact'];
const LAYOUTS: Array<TreeLayout> = ['default', 'inline'];
const ORIENTATIONS: Array<TreeOrientation> = ['vertical', 'horizontal'];

interface TreeStoryProps {
	density?: TreeDensity;
	layout?: TreeLayout;
	orientation?: TreeOrientation;
	disabled?: boolean;
	defaultOpen?: boolean;
	onAdd?: () => void;
	onDelete?: () => void;
}

const ROWS = ['Opener', 'Main priority', 'Cooldowns'];

const Row = ({
	id,
	title,
	defaultOpen,
	onDelete,
	children,
}: {
	id: string;
	title: string;
	defaultOpen?: boolean;
	onDelete?: () => void;
	children?: ReactNode;
}) => (
	<Tree.Item id={id} defaultOpen={defaultOpen}>
		<Tree.Item.Header>
			<Tree.Item.Handle label={`Move ${title}`} />
			<Tree.Item.Title>{title}</Tree.Item.Title>
			<Tree.Item.Trigger label={`Toggle ${title}`} />
			<Tree.Item.Actions label={`${title} actions`}>
				<Tree.Item.Action icon="copy" label="Copy" />
				<Tree.Item.Action icon="times" label="Delete" tone="danger" onClick={onDelete} />
			</Tree.Item.Actions>
		</Tree.Item.Header>
		<Tree.Item.Body>The body of {title}.</Tree.Item.Body>
		{children}
	</Tree.Item>
);

const TreeStory = ({ density, layout, orientation, disabled, defaultOpen, onAdd = () => {}, onDelete }: TreeStoryProps) => (
	<Tree.Root density={density} layout={layout} orientation={orientation} disabled={disabled}>
		<Tree.Items>
			{ROWS.map(title => (
				<Row key={title} id={title} title={title} defaultOpen={defaultOpen} onDelete={onDelete} />
			))}
		</Tree.Items>
		<Tree.AddButton label="New row" onClick={onAdd} />
	</Tree.Root>
);

const InlineTree = ({ onDelete }: { onDelete?: () => void }) => (
	<Tree.Root layout="inline">
		<Tree.Items>
			{ROWS.map(title => (
				<Tree.Item key={title} id={title}>
					<Tree.Item.Header>
						<Tree.Item.Handle label={`Move ${title}`} />
						<Tree.Item.Actions label={`${title} actions`}>
							<Tree.Item.Action icon="times" label="Delete" tone="danger" onClick={onDelete} />
						</Tree.Item.Actions>
					</Tree.Item.Header>
					<Tree.Item.Body>The body of {title}, on the header row.</Tree.Item.Body>
				</Tree.Item>
			))}
		</Tree.Items>
		<Tree.AddButton label="New row" icon="plus" onClick={() => {}} />
	</Tree.Root>
);

const meta = {
	title: 'Kit/Tree',
	component: TreeStory,
	tags: ['autodocs'],
	args: { onAdd: fn(), onDelete: fn() },
	argTypes: {
		density: { control: 'select', options: DENSITIES },
		layout: { control: 'select', options: LAYOUTS },
		orientation: { control: 'select', options: ORIENTATIONS },
	},
} satisfies Meta<typeof TreeStory>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Nested: Story = {
	render: ({ onDelete }) => (
		<Tree.Root>
			<Tree.Items>
				<Row id="group" title="Group" onDelete={onDelete}>
					<Tree.Item.Children>
						<Tree.Items>
							<Row id="sequence" title="Sequence" onDelete={onDelete}>
								<Tree.Item.Children>
									<Tree.Items>
										<Row id="step-1" title="Step 1" onDelete={onDelete} />
										<Row id="step-2" title="Step 2" onDelete={onDelete} />
									</Tree.Items>
								</Tree.Item.Children>
							</Row>
						</Tree.Items>
					</Tree.Item.Children>
				</Row>
			</Tree.Items>
		</Tree.Root>
	),
};

export const Compact: Story = { args: { density: 'compact' } };

export const Inline: Story = {
	render: ({ onDelete }) => <InlineTree onDelete={onDelete} />,
};

export const Horizontal: Story = { args: { orientation: 'horizontal' } };

export const Collapsed: Story = {
	args: { defaultOpen: false },
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const toggle = canvas.getByRole('button', { name: 'Toggle Opener' });
		await expect(toggle).toHaveAttribute('aria-expanded', 'false');
		await userEvent.click(toggle);
		await expect(toggle).toHaveAttribute('aria-expanded', 'true');
	},
};

export const Disabled: Story = { args: { disabled: true } };

export const WithActions: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'Opener actions' }));
		await userEvent.click(await screen.findByRole('button', { name: 'Delete' }));
		await expect(args.onDelete).toHaveBeenCalledTimes(1);
	},
};

export const Narrow: Story = {
	render: ({ onDelete }) => (
		<div className="w-64">
			<InlineTree onDelete={onDelete} />
		</div>
	),
};
