import { supportedLanguages } from '@i18n/locale_service';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
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
const checked = (testId: string) => within(screen.getByTestId(testId)).getByRole('checkbox').getAttribute('aria-checked');
const concurrencyRow = () => screen.getByTestId('use-concurrency-container');
const workerNote = () => concurrencyRow().querySelector('.ui-help-text') as HTMLElement;

describe('SettingsDialogView', () => {
	it('shows each value it is given', () => {
		render(<SettingsDialogView {...props({ fixedRngSeed: 42, lastUsedRngSeed: 1234, showThreatMetrics: true, wasmConcurrency: 3 })} />);

		expect((byId('simui-fixed-rng-seed') as HTMLInputElement).value).toBe('42');
		expect(screen.getByTestId('last-used-rng-seed').textContent).toBe('1234');
		expect((byId('simui-language-picker') as HTMLSelectElement).value).toBe(String(Object.keys(supportedLanguages).indexOf('en')));
		expect(checked('show-threat-metrics-picker')).toBe('true');
		expect(checked('show-experimental-picker')).toBe('false');
		expect(checked('show-quick-swap-picker')).toBe('true');
		expect((byId('simui-concurrent-workers-picker') as HTMLSelectElement).value).toBe('3');
	});

	it('offers Off and two up to the maximum worker count', () => {
		render(<SettingsDialogView {...props({ maxWorkers: 4 })} />);

		const options = [...(byId('simui-concurrent-workers-picker') as HTMLSelectElement).options].map(option => option.textContent);
		expect(options).toEqual(['Off', '2', '3', '4']);
	});

	it('reports each change through its handler', () => {
		const handlers = props();
		render(<SettingsDialogView {...handlers} />);

		act(() => {
			fireEvent.click(byId('simui-show-threat-metrics'));
			fireEvent.click(byId('simui-show-experimental'));
			fireEvent.click(byId('simui-show-quick-swap'));
			fireEvent.change(byId('simui-concurrent-workers-picker'), { target: { value: '2' } });
		});

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

	it('reports the chosen language by its code', () => {
		const handlers = props();
		render(<SettingsDialogView {...handlers} />);
		const codes = Object.keys(supportedLanguages);

		act(() => {
			fireEvent.change(byId('simui-language-picker'), { target: { value: String(codes.length - 1) } });
		});

		expect(handlers.onLanguageChange).toHaveBeenCalledWith(codes[codes.length - 1]);
	});

	it('selects English for a language it does not know', () => {
		render(<SettingsDialogView {...props({ language: '' })} />);

		expect((byId('simui-language-picker') as HTMLSelectElement).value).toBe(String(Object.keys(supportedLanguages).indexOf('en')));
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
