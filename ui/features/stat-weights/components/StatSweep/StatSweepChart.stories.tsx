import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeCritSweepPoints, makeHitSweepPoints, sweepSeries } from '@testing/storybook/fixtures/stat-sweep';

import { StatSweepChart } from './StatSweepChart';

const CANVAS = { smoke: false, smokeReason: 'chart.js needs a canvas' };

const meta = {
	title: 'Features/StatWeights/StatSweep/StatSweepChart',
	component: StatSweepChart,
	tags: ['autodocs'],
	args: { series: [] },
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
	args: { series: [sweepSeries('current', 'Hit Rating', 120, makeHitSweepPoints(120))] },
	parameters: CANVAS,
};

export const Overlay: Story = {
	args: {
		series: [
			sweepSeries('current', 'Hit Rating', 120, makeHitSweepPoints(120)),
			sweepSeries('older', 'Hit Rating', 90, makeHitSweepPoints(90, 10_000), 1, 'Hit Rating ±100 (old gear)'),
		],
	},
	parameters: CANVAS,
};

export const OverlayOtherStat: Story = {
	args: {
		series: [sweepSeries('hit', 'Hit Rating', 120, makeHitSweepPoints(120)), sweepSeries('crit', 'Crit Rating', 210, makeCritSweepPoints(210), 1)],
	},
	parameters: CANVAS,
};
