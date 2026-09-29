import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { SettingsDialogView } from './SettingsDialogView';

const meta = {
	title: 'App/SettingsDialog/SettingsDialogView',
	component: SettingsDialogView,
	tags: ['autodocs'],
	args: {
		open: true,
		onOpenChange: fn(),
		fixedRngSeed: 0,
		onFixedRngSeedChange: fn(),
		lastUsedRngSeed: 1839204711,
		language: 'en',
		onLanguageChange: fn(),
		showThreatMetrics: false,
		onShowThreatMetricsChange: fn(),
		showExperimental: false,
		onShowExperimentalChange: fn(),
		showQuickSwap: true,
		onShowQuickSwapChange: fn(),
		wasmConcurrency: 4,
		onWasmConcurrencyChange: fn(),
		showConcurrency: true,
		maxWorkers: 8,
		firefox: false,
		onApplyDefaults: fn(),
	},
} satisfies Meta<typeof SettingsDialogView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		const frame = within(canvasElement.ownerDocument.body);
		await userEvent.click(await frame.findByTestId('restore-defaults'));
		await expect(args.onApplyDefaults).toHaveBeenCalledOnce();
	},
};

export const FixedSeed: Story = { args: { fixedRngSeed: 12345, showThreatMetrics: true, showExperimental: true } };

export const Firefox: Story = { args: { firefox: true } };

export const NativeSim: Story = { args: { showConcurrency: false } };
