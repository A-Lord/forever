import { createSimStore } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { SettingsDialog } from './SettingsDialog';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const toasts = vi.hoisted(() => [] as Array<{ variant: string; body: unknown }>);
vi.mock('@ui-kit/Toast', async importOriginal => ({
	...(await importOriginal<typeof import('@ui-kit/Toast')>()),
	toastManager: {
		add: (options: { variant: string; body: unknown }) => {
			toasts.push(options);
			return '';
		},
		close: () => {},
	},
}));
const trackEvent = vi.hoisted(() => vi.fn());
vi.mock('../../tracking/utils', () => ({ trackEvent }));

class FakeSim {
	readonly store = createSimStore();
	wasm = Promise.resolve(true);

	isWasm() {
		return this.wasm;
	}
	setFixedRngSeed(value: number) {
		this.store.setState(state => ({ sim: { ...state.sim, fixedRngSeed: value } }));
	}
	recordLastUsedRngSeed(value: number) {
		this.store.setState(state => ({
			sim: { ...state.sim, lastUsedRngSeed: value, lastUsedRngSeedVersion: state.sim.lastUsedRngSeedVersion + 1 },
		}));
	}
	setLanguage(value: string) {
		this.store.setState(state => ({ ui: { ...state.ui, language: value } }));
	}
	getShowThreatMetrics() {
		return this.store.getState().ui.showThreatMetrics;
	}
	setShowThreatMetrics(value: boolean) {
		this.store.setState(state => ({ ui: { ...state.ui, showThreatMetrics: value } }));
	}
	setShowExperimental(value: boolean) {
		this.store.setState(state => ({ ui: { ...state.ui, showExperimental: value } }));
	}
	setShowQuickSwap(value: boolean) {
		this.store.setState(state => ({ ui: { ...state.ui, showQuickSwap: value } }));
	}
	setWasmConcurrency(value: number) {
		this.store.setState(state => ({ ui: { ...state.ui, wasmConcurrency: value } }));
	}
}

const mount = (sim: FakeSim, applyDefaults = vi.fn()) => {
	const host = { applyDefaults } as never;
	const rendered = renderWithHost(<SettingsDialog open onOpenChange={() => {}} host={host} />, { sim });
	return { ...rendered, applyDefaults };
};

const container = () => document.querySelector('[data-testid="use-concurrency-container"]') as HTMLElement;

describe('SettingsDialog', () => {
	it('keeps the concurrency row mounted and hides it with the attribute when the sim is not wasm', async () => {
		const sim = new FakeSim();
		sim.wasm = Promise.resolve(false);
		mount(sim);

		expect(container().hidden).toBe(false);
		await waitFor(() => expect(container().hidden).toBe(true));
		expect(container().isConnected).toBe(true);
	});

	it('writes a toggle through to the sim', () => {
		const sim = new FakeSim();
		mount(sim);

		act(() => {
			fireEvent.click(document.getElementById('simui-show-threat-metrics')!);
		});
		expect(sim.getShowThreatMetrics()).toBe(true);
	});

	it('follows the last used rng seed', async () => {
		const sim = new FakeSim();
		mount(sim);

		const lastUsedRngSeedText = () => screen.getByText('info.options.fixed_rng_seed.last_used').parentElement!.textContent;
		expect(lastUsedRngSeedText()).toContain('0');
		act(() => {
			sim.recordLastUsedRngSeed(1234);
		});
		await waitFor(() => expect(lastUsedRngSeedText()).toContain('1234'));
	});

	it('restores defaults, reports it and toasts', () => {
		toasts.length = 0;
		trackEvent.mockClear();
		const { applyDefaults } = mount(new FakeSim());

		act(() => {
			fireEvent.click(screen.getByRole('button', { name: 'info.options.restore_defaults.button' }));
		});

		expect(applyDefaults).toHaveBeenCalledTimes(1);
		expect(trackEvent).toHaveBeenCalledWith({ action: 'settings', category: 'restore-defaults', label: 'restore' });
		expect(toasts).toEqual([{ variant: 'success', body: 'info.options.restore_defaults.success_message' }]);
	});
});
