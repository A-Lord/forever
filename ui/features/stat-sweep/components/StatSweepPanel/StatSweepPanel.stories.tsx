import { PrecisionOptions } from '@features/stat-weights/components/PrecisionOptions';
import i18n from '@i18n/config';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeHitSweepPoints, sweepSeries } from '@testing/storybook/fixtures/stat-sweep';
import { fn } from 'storybook/test';

import { StatSweepChart } from '../StatSweepChart';
import { StatSweepHistory } from '../StatSweepHistory';
import { StatSweepPanel } from './StatSweepPanel';

const precisionText = (key: string) => i18n.t(`sidebar.buttons.stat_sweep.precision.${key}`, { metric: i18n.t('sidebar.results.metrics.dps.label') });

const HISTORY = [
	{ id: 'hit-new', label: 'Hit Rating ±100', createdAt: Date.UTC(2026, 9, 12, 14, 3), overlaid: true, colorSlot: 1, metric: 'dps' as const },
	{ id: 'crit', label: 'Crit Rating ±50', createdAt: Date.UTC(2026, 9, 10, 18, 15), overlaid: false, metric: 'tps' as const },
];

const meta = {
	title: 'Features/StatSweep/StatSweepPanel',
	component: StatSweepPanel,
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
			<PrecisionOptions
				value={{ precisionMode: 'iterations', iterations: 0, targetCi: 1, maxIterationMultiplier: 0 }}
				onChange={fn()}
				defaultIterations={10000}
				idPrefix="stat-sweep-precision"
				targetLabel={precisionText('target_ci')}
				targetTooltip={precisionText('target_ci_tooltip')}
				iterationsTooltip={precisionText('iterations_tooltip')}
				maxMultiplierTooltip={precisionText('max_multiplier_tooltip')}
				maxDecimalDigits={2}
			/>
		),
		chart: <StatSweepChart series={[]} metric="dps" metrics={['dps', 'tps']} onMetricChange={fn()} showSpread onShowSpreadChange={fn()} />,
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
				metric="dps"
				metrics={['dps', 'tps']}
				onMetricChange={fn()}
				showSpread
				onShowSpreadChange={fn()}
			/>
		),
	},
	parameters: { smoke: false, smokeReason: 'chart.js needs a canvas' },
};
