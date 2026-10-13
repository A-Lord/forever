import type { Meta, StoryObj } from '@storybook/react-vite';

import { StatSweepTimeRemaining } from './StatSweepTimeRemaining';

const meta = {
	title: 'Features/StatWeights/StatSweep/StatSweepTimeRemaining',
	component: StatSweepTimeRemaining,
	tags: ['autodocs'],
	args: { seconds: 95 },
} satisfies Meta<typeof StatSweepTimeRemaining>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Estimated: Story = {};

export const NoEstimateYet: Story = { args: { seconds: undefined } };
