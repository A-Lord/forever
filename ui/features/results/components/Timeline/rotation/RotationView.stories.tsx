import type { Meta, StoryObj } from '@storybook/react-vite';
import { warriorRotationModel } from '@testing/storybook/fixtures/results-views';
import { type ReactNode, useState } from 'react';

import { DrStickySlotContext } from '../../DetailedResults/context/DrStickySlotContext';
import { RotationView } from './RotationView';

// The ruler and the zoom buttons portal into the detailed results' sticky slot.
const StickySlot = ({ children }: { children: ReactNode }) => {
	const [slot, setSlot] = useState<HTMLDivElement | null>(null);
	return (
		<>
			<div ref={setSlot} />
			<DrStickySlotContext.Provider value={slot}>{children}</DrStickySlotContext.Provider>
		</>
	);
};

const meta = {
	title: 'Features/Results/Timeline/Components/RotationView',
	component: RotationView,
	tags: ['autodocs'],
	args: { model: warriorRotationModel(), active: true },
	argTypes: { model: { control: false } },
	decorators: [
		Story => (
			<StickySlot>
				<Story />
			</StickySlot>
		),
	],
} satisfies Meta<typeof RotationView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Warrior: Story = {};

export const NoResult: Story = { args: { model: null } };
