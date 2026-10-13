import type { Meta, StoryObj } from '@storybook/react-vite';

import { DpsCurveTimeRemaining } from './DpsCurveTimeRemaining';

const meta = {
	title: 'Features/DpsCurve/DpsCurveTimeRemaining',
	component: DpsCurveTimeRemaining,
	tags: ['autodocs'],
	args: { seconds: 95 },
} satisfies Meta<typeof DpsCurveTimeRemaining>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Estimated: Story = {};

export const NoEstimateYet: Story = { args: { seconds: undefined } };
