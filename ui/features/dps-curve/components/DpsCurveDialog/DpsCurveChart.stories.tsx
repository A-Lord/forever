import type { Meta, StoryObj } from '@storybook/react-vite';
import { curveSeries, makeCritCurvePoints, makeHitCurvePoints } from '@testing/storybook/fixtures/dps-curve';

import { DpsCurveChart } from './DpsCurveChart';

const CANVAS = { smoke: false, smokeReason: 'chart.js needs a canvas' };

const meta = {
	title: 'Features/DpsCurve/DpsCurveChart',
	component: DpsCurveChart,
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
} satisfies Meta<typeof DpsCurveChart>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Empty: Story = {};

export const SingleRun: Story = {
	args: { series: [curveSeries('current', 'Hit Rating', 120, makeHitCurvePoints(120))] },
	parameters: CANVAS,
};

export const Overlay: Story = {
	args: {
		series: [
			curveSeries('current', 'Hit Rating', 120, makeHitCurvePoints(120)),
			curveSeries('older', 'Hit Rating', 90, makeHitCurvePoints(90, 10_000), 1, 'Hit Rating ±100 (old gear)'),
		],
	},
	parameters: CANVAS,
};

export const OverlayOtherStat: Story = {
	args: {
		series: [curveSeries('hit', 'Hit Rating', 120, makeHitCurvePoints(120)), curveSeries('crit', 'Crit Rating', 210, makeCritCurvePoints(210), 1)],
	},
	parameters: CANVAS,
};
