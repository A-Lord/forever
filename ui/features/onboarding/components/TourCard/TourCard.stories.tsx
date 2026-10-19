import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { TourCard } from './TourCard';

const meta = {
	title: 'Features/Onboarding/TourCard',
	component: TourCard,
	tags: ['autodocs'],
	args: {
		kind: 'step',
		title: 'Your equipment',
		body: 'Click a slot to pick its item, enchant and gems. Hovering an item shows its tooltip.',
		position: 3,
		total: 13,
		onNext: fn(),
		onPrevious: fn(),
		onSkip: fn(),
	},
	argTypes: {
		kind: { control: 'inline-radio', options: ['welcome', 'step', 'finish'] },
		style: { control: false },
		cardRef: { control: false },
		side: { control: false },
	},
	decorators: [
		Story => (
			<div className="relative h-80">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof TourCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Step: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('tour-card-next'));
		await expect(args.onNext).toHaveBeenCalledOnce();
	},
};

export const Welcome: Story = {
	args: {
		kind: 'welcome',
		title: 'Welcome to WoWSims',
		body: 'Two minutes to see how a sim comes together. Skip any time; the question mark in the toolbar brings this back.',
		onPrevious: undefined,
	},
};

/** A step that waits for the player to run the sim: the primary button passes it over instead. */
export const Waiting: Story = { args: { title: 'Run the sim', body: 'Press Simulate. The tour moves on by itself when the run finishes.', waiting: true } };

export const Finish: Story = {
	args: { kind: 'finish', title: 'You’re set', body: 'If something looks wrong, report it with the bug icon in the toolbar or tell us on Discord.' },
};
