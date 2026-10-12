import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeHitSweepPoints } from '@testing/storybook/fixtures/stat-sweep';

import { StatSweepChart } from './StatSweepChart';

const CANVAS = { smoke: false, smokeReason: 'chart.js needs a canvas' };

const meta = {
	title: 'Features/StatWeights/StatSweep/StatSweepChart',
	component: StatSweepChart,
	tags: ['autodocs'],
	args: { series: [], statLabel: 'Hit Rating', currentValue: 120 },
	argTypes: { series: { control: false } },
	decorators: [
		Story => (
			<div className="max-w-3xl">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof StatSweepChart>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Empty: Story = {};

export const SingleRun: Story = {
	args: { series: [{ id: 'current', label: 'Hit Rating ±100', colorSlot: 0, points: makeHitSweepPoints(120) }] },
	parameters: CANVAS,
};

export const Overlay: Story = {
	args: {
		series: [
			{ id: 'current', label: 'Hit Rating ±100', colorSlot: 0, points: makeHitSweepPoints(120) },
			{ id: 'older', label: 'Hit Rating ±100 (old gear)', colorSlot: 1, points: makeHitSweepPoints(90, 10_000) },
		],
	},
	parameters: CANVAS,
};
