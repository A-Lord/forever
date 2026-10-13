import type { Meta, StoryObj } from '@storybook/react-vite';
import { curveSeries, makeHitCurvePoints } from '@testing/storybook/fixtures/dps-curve';
import { fn } from 'storybook/test';

import { DpsCurveChart } from './DpsCurveChart';
import { DpsCurveHistory } from './DpsCurveHistory';
import { DpsCurvePanel } from './DpsCurvePanel';
import { DpsCurvePrecisionOptions } from './DpsCurvePrecisionOptions';

const HISTORY = [
	{ id: 'hit-new', label: 'Hit Rating ±100', createdAt: Date.UTC(2026, 9, 12, 14, 3), overlaid: true, colorSlot: 1 },
	{ id: 'crit', label: 'Crit Rating ±50', createdAt: Date.UTC(2026, 9, 10, 18, 15), overlaid: false },
];

const meta = {
	title: 'Features/DpsCurve/DpsCurvePanel',
	component: DpsCurvePanel,
	tags: ['autodocs'],
	args: {
		statOptions: [
			{ value: 'stat-hit', label: 'Hit Rating' },
			{ value: 'stat-crit', label: 'Crit Rating' },
			{ value: 'stat-agility', label: 'Agility' },
		],
		stats: ['stat-hit', 'stat-crit'],
		onStatsChange: fn(),
		range: 100,
		onRangeChange: fn(),
		step: 10,
		onStepChange: fn(),
		precision: (
			<DpsCurvePrecisionOptions
				value={{ precisionMode: 'iterations', iterations: 0, targetDpsCi: 1, maxIterationMultiplier: 0 }}
				onChange={fn()}
				defaultIterations={10000}
			/>
		),
		chart: <DpsCurveChart series={[]} />,
		history: <DpsCurveHistory runs={HISTORY} onOverlayChange={fn()} onDelete={fn()} />,
		running: false,
	},
	argTypes: { statOptions: { control: false }, precision: { control: false }, chart: { control: false }, history: { control: false } },
} satisfies Meta<typeof DpsCurvePanel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Idle: Story = {};

export const Running: Story = {
	args: {
		running: true,
		chart: (
			<DpsCurveChart
				series={[
					curveSeries('current', 'Hit Rating', 120, makeHitCurvePoints(120).slice(0, 7)),
					curveSeries('hit-new', 'Hit Rating', 90, makeHitCurvePoints(90), 1),
				]}
			/>
		),
	},
	parameters: { smoke: false, smokeReason: 'chart.js needs a canvas' },
};
