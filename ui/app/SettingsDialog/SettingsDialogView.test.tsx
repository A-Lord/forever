import { supportedLanguages } from '@i18n/locale_service';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { SettingsDialogView, type SettingsDialogViewProps } from './SettingsDialogView';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const props = (overrides: Partial<SettingsDialogViewProps> = {}): SettingsDialogViewProps => ({
	open: true,
	onOpenChange: vi.fn(),
	fixedRngSeed: 0,
	onFixedRngSeedChange: vi.fn(),
	lastUsedRngSeed: 0,
	language: 'en',
	onLanguageChange: vi.fn(),
	showThreatMetrics: false,
	onShowThreatMetricsChange: vi.fn(),
	showExperimental: false,
	onShowExperimentalChange: vi.fn(),
	showQuickSwap: true,
	onShowQuickSwapChange: vi.fn(),
	wasmConcurrency: 0,
	onWasmConcurrencyChange: vi.fn(),
	showConcurrency: true,
	maxWorkers: 4,
	firefox: false,
	onApplyDefaults: vi.fn(),
	...overrides,
});

const byId = (id: string) => document.getElementById(id) as HTMLElement;
const pick = async (id: string, option: string) => {
	await userEvent.click(byId(id));
	await userEvent.click(screen.getByRole('option', { name: option }));
};
const checked = (testId: string) => within(screen.getByTestId(testId)).getByRole('checkbox').getAttribute('aria-checked');
const concurrencyRow = () => screen.getByTestId('use-concurrency-container');
const workerNote = () => concurrencyRow().querySelector('.ui-help-text') as HTMLElement;

describe('SettingsDialogView', () => {
	it('shows each value it is given', () => {
		render(<SettingsDialogView {...props({ fixedRngSeed: 42, lastUsedRngSeed: 1234, showThreatMetrics: true, wasmConcurrency: 3 })} />);

		expect((byId('simui-fixed-rng-seed') as HTMLInputElement).value).toBe('42');
		expect(screen.getByText('info.options.fixed_rng_seed.last_used').parentElement?.textContent).toContain('1234');
		expect(byId('simui-language-picker').textContent).toBe(supportedLanguages.en);
		expect(checked('show-threat-metrics-picker')).toBe('true');
		expect(checked('show-experimental-picker')).toBe('false');
		expect(checked('show-quick-swap-picker')).toBe('true');
		expect(byId('simui-concurrent-workers-picker').textContent).toBe('3');
	});

	it('offers Off and two up to the maximum worker count', async () => {
		render(<SettingsDialogView {...props({ maxWorkers: 4 })} />);

		await userEvent.click(byId('simui-concurrent-workers-picker'));
		expect(screen.getAllByRole('option').map(option => option.textContent)).toEqual(['Off', '2', '3', '4']);
	});

	it('reports each change through its handler', async () => {
		const handlers = props();
		render(<SettingsDialogView {...handlers} />);

		act(() => {
			fireEvent.click(byId('simui-show-threat-metrics'));
			fireEvent.click(byId('simui-show-experimental'));
			fireEvent.click(byId('simui-show-quick-swap'));
		});
		await pick('simui-concurrent-workers-picker', '2');

		expect(handlers.onShowThreatMetricsChange).toHaveBeenCalledWith(true);
		expect(handlers.onShowExperimentalChange).toHaveBeenCalledWith(true);
		expect(handlers.onShowQuickSwapChange).toHaveBeenCalledWith(false);
		expect(handlers.onWasmConcurrencyChange).toHaveBeenCalledWith(2);
	});

	it('commits the seed on change', () => {
		const handlers = props();
		render(<SettingsDialogView {...handlers} />);

		act(() => {
			fireEvent.change(byId('simui-fixed-rng-seed'), { target: { value: '99' } });
		});

		expect(handlers.onFixedRngSeedChange).toHaveBeenCalledWith(99);
	});

	it('reports the chosen language by its code', async () => {
		const handlers = props();
		render(<SettingsDialogView {...handlers} />);
		const [code, name] = Object.entries(supportedLanguages).at(-1)!;

		await pick('simui-language-picker', name);

		expect(handlers.onLanguageChange).toHaveBeenCalledWith(code);
	});

	it('selects English for a language it does not know', () => {
		render(<SettingsDialogView {...props({ language: '' })} />);

		expect(byId('simui-language-picker').textContent).toBe(supportedLanguages.en);
	});

	it('keeps the concurrency row mounted and hidden when it is not shown', () => {
		render(<SettingsDialogView {...props({ showConcurrency: false })} />);

		expect(concurrencyRow().hidden).toBe(true);
	});

	it('shows the worker note on firefox only', () => {
		const { rerender } = render(<SettingsDialogView {...props()} />);
		expect(workerNote().hidden).toBe(true);
		expect(workerNote().textContent).toBe('');

		rerender(<SettingsDialogView {...props({ firefox: true })} />);
		expect(workerNote().hidden).toBe(false);
		expect(workerNote().textContent).toContain('Too many workers');
	});

	it('applies the defaults from the footer button', () => {
		const handlers = props();
		render(<SettingsDialogView {...handlers} />);

		act(() => {
			fireEvent.click(screen.getByRole('button', { name: 'info.options.restore_defaults.button' }));
		});

		expect(handlers.onApplyDefaults).toHaveBeenCalledTimes(1);
	});
});
