import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeHitSweepPoints, sweepSeries } from '@testing/storybook/fixtures/stat-sweep';
import { fn } from 'storybook/test';

import { EpPrecisionOptions } from '../EpPrecisionOptions/EpPrecisionOptions';
import { StatSweepChart } from './StatSweepChart';
import { StatSweepHistory } from './StatSweepHistory';
import { StatSweepPanel } from './StatSweepPanel';

const HISTORY = [
	{ id: 'hit-new', label: 'Hit Rating ±100', createdAt: Date.UTC(2026, 9, 12, 14, 3), overlaid: true, colorSlot: 1 },
	{ id: 'crit', label: 'Crit Rating ±50', createdAt: Date.UTC(2026, 9, 10, 18, 15), overlaid: false },
];

const meta = {
	title: 'Features/StatWeights/StatSweep/StatSweepPanel',
	component: StatSweepPanel,
	tags: ['autodocs'],
	args: {
		statOptions: [
			{ value: 'stat-hit', label: 'Hit Rating' },
			{ value: 'stat-crit', label: 'Crit Rating' },
			{ value: 'stat-agility', label: 'Agility' },
		],
		stat: 'stat-hit',
		onStatChange: fn(),
		range: 100,
		onRangeChange: fn(),
		step: 10,
		onStepChange: fn(),
		precision: (
			<EpPrecisionOptions
				value={{ precisionMode: 'iterations', iterations: 0, targetEpCi: 0.01, maxIterationMultiplier: 0 }}
				onChange={fn()}
				defaultIterations={10000}
			/>
		),
		chart: <StatSweepChart series={[]} />,
		history: <StatSweepHistory runs={HISTORY} onOverlayChange={fn()} onDelete={fn()} />,
		running: false,
	},
	argTypes: { statOptions: { control: false }, precision: { control: false }, chart: { control: false }, history: { control: false } },
} satisfies Meta<typeof StatSweepPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Idle: Story = {};

export const Running: Story = {
	args: {
		running: true,
		chart: (
			<StatSweepChart
				series={[
					sweepSeries('current', 'Hit Rating', 120, makeHitSweepPoints(120).slice(0, 7)),
					sweepSeries('hit-new', 'Hit Rating', 90, makeHitSweepPoints(90), 1),
				]}
			/>
		),
	},
	parameters: { smoke: false, smokeReason: 'chart.js needs a canvas' },
};
