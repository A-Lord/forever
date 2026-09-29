import type { Meta, StoryObj } from '@storybook/react-vite';
import { ToggleGroup } from '@ui-kit/ToggleGroup';
import { useState } from 'react';
import { expect, fireEvent, within } from 'storybook/test';

import { Accordion } from './Accordion';
import { AccordionItem } from './AccordionItem';

interface AccordionStoryProps {
	initial: Array<string>;
}

const AccordionStory = ({ initial }: AccordionStoryProps) => {
	const [open, setOpen] = useState(initial);
	return (
		<Accordion value={open} onValueChange={setOpen}>
			<AccordionItem value="consumes" title="Consumables" testId="story-consumes">
				<p>Flask of Supreme Power, Elixir of Greater Firepower</p>
			</AccordionItem>
			<AccordionItem value="buffs" title="Raid buffs" testId="story-buffs">
				<p>Arcane Intellect, Mark of the Wild, Blessing of Kings</p>
			</AccordionItem>
			<AccordionItem value="debuffs" title="Debuffs" testId="story-debuffs">
				<p>Curse of Elements, Improved Scorch</p>
			</AccordionItem>
		</Accordion>
	);
};

const HeaderRowStory = ({ initial }: AccordionStoryProps) => {
	const [open, setOpen] = useState(initial);
	const [mode, setMode] = useState<'key' | 'all'>('key');
	return (
		<div className="w-62">
			<Accordion value={open} onValueChange={setOpen}>
				<AccordionItem
					value="stats"
					title="Stats"
					panelFirst
					testId="story-stats"
					header={
						<div className="flex flex-1 items-center justify-between gap-2">
							<span className="font-bold">Stats</span>
							<ToggleGroup
								size="xs"
								aria-label="Stats shown"
								value={mode}
								onChange={setMode}
								options={[
									{ value: 'key', label: 'Key' },
									{ value: 'all', label: 'All' },
								]}
							/>
						</div>
					}>
					<p>Strength 320, Agility 180, Stamina 410</p>
				</AccordionItem>
			</Accordion>
		</div>
	);
};

const meta = {
	title: 'Kit/Accordion',
	component: AccordionStory,
	tags: ['autodocs'],
	args: { initial: ['consumes'] },
} satisfies Meta<typeof AccordionStory>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Multiple: Story = { args: { initial: ['consumes', 'debuffs'] } };

export const Opens: Story = {
	args: { initial: [] },
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.queryByText(/Curse of Elements/)).toBeNull();

		await fireEvent.click(canvas.getByTestId('story-debuffs-trigger'));

		await expect(canvas.getByText(/Curse of Elements/)).toBeTruthy();
	},
};

export const HeaderRow: Story = {
	args: { initial: [] },
	parameters: {
		docs: {
			description: {
				story: 'The trigger spans the whole header row: a click on the title or anywhere else in the row toggles the panel, while Key and All take their own clicks. Tab reaches the trigger first, then the toggle group.',
			},
		},
	},
	render: args => <HeaderRowStory {...args} />,
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await fireEvent.click(canvas.getByRole('button', { name: 'All' }));
		await expect(canvas.queryByText(/Strength 320/)).toBeNull();

		await fireEvent.click(canvas.getByTestId('story-stats-trigger'));

		await expect(canvas.getByText(/Strength 320/)).toBeTruthy();
	},
};
