import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { EpPrecisionOptions } from './EpPrecisionOptions';

const meta = {
	title: 'Features/StatWeights/EpPrecisionOptions',
	component: EpPrecisionOptions,
	tags: ['autodocs'],
	args: {
		value: { precisionMode: 'iterations', iterations: 0, targetEpCi: 0.01, maxIterationMultiplier: 0 },
		onChange: fn(),
		defaultIterations: 10000,
	},
	argTypes: { value: { control: 'object' } },
} satisfies Meta<typeof EpPrecisionOptions>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Iterations: Story = {};

export const TargetPrecision: Story = {
	args: { value: { precisionMode: 'target', iterations: 0, targetEpCi: 0.01, maxIterationMultiplier: 0 } },
};

export const Disabled: Story = {
	args: { value: { precisionMode: 'target', iterations: 0, targetEpCi: 0.005, maxIterationMultiplier: 4 }, disabled: true },
};
