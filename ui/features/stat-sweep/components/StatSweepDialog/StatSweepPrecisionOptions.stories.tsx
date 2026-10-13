import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { StatSweepPrecisionOptions } from './StatSweepPrecisionOptions';

const meta = {
	title: 'Features/StatSweep/StatSweepPrecisionOptions',
	component: StatSweepPrecisionOptions,
	tags: ['autodocs'],
	args: {
		value: { precisionMode: 'iterations', iterations: 0, targetCi: 1, maxIterationMultiplier: 0 },
		onChange: fn(),
		metric: 'dps',
		defaultIterations: 10000,
	},
	argTypes: { value: { control: 'object' }, metric: { control: 'inline-radio', options: ['dps', 'hps', 'tps', 'dtps', 'tmi'] } },
	decorators: [
		Story => (
			<div className="w-64">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof StatSweepPrecisionOptions>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Iterations: Story = {};

export const TargetPrecision: Story = {
	args: { value: { precisionMode: 'target', iterations: 0, targetCi: 1, maxIterationMultiplier: 0 } },
};

export const TargetThreat: Story = {
	args: { value: { precisionMode: 'target', iterations: 0, targetCi: 2, maxIterationMultiplier: 4 }, metric: 'tps' },
};

export const Disabled: Story = {
	args: { value: { precisionMode: 'target', iterations: 0, targetCi: 0.5, maxIterationMultiplier: 8 }, disabled: true },
};
