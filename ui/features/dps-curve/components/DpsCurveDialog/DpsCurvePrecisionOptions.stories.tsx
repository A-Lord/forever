import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { DpsCurvePrecisionOptions } from './DpsCurvePrecisionOptions';

const meta = {
	title: 'Features/DpsCurve/DpsCurvePrecisionOptions',
	component: DpsCurvePrecisionOptions,
	tags: ['autodocs'],
	args: {
		value: { precisionMode: 'iterations', iterations: 0, targetDpsCi: 1, maxIterationMultiplier: 0 },
		onChange: fn(),
		defaultIterations: 10000,
	},
	argTypes: { value: { control: 'object' } },
	decorators: [
		Story => (
			<div className="w-64">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof DpsCurvePrecisionOptions>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Iterations: Story = {};

export const TargetPrecision: Story = {
	args: { value: { precisionMode: 'target', iterations: 0, targetDpsCi: 1, maxIterationMultiplier: 0 } },
};

export const Disabled: Story = {
	args: { value: { precisionMode: 'target', iterations: 0, targetDpsCi: 0.5, maxIterationMultiplier: 8 }, disabled: true },
};
