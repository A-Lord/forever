import { SimHostProvider } from '@sim/context/SimHostContext';
import { fakeHost } from '@sim/testing';
import { act, fireEvent, render } from '@testing-library/react';
import { PortalContainerContext } from '@ui-kit/hooks/usePortalContainer';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const trackPageView = vi.fn();
const runState = { isRunning: false, isAborting: false, abort: vi.fn() };

vi.mock('../../../../tracking/utils', () => ({ trackEvent: vi.fn(), trackPageView: (...args: unknown[]) => trackPageView(...args) }));
vi.mock('@sim/hooks/useSimRun', () => ({ useSimRun: () => runState }));
vi.mock('./ReforgeSettingsPanel', () => ({ ReforgeSettingsPanel: () => <div data-testid="settings-body" /> }));

const { ReforgePanel } = await import('./ReforgePanel');

const model = {
	settings: {},
	softCapsConfigWithLimits: [],
	previousGear: null,
} as any;

const host = fakeHost({ rootElem: document.createElement('div'), sim: { runs: { isRunning: () => false } } });

const renderPanel = () =>
	render(
		<PortalContainerContext value={host.rootElem}>
			<SimHostProvider host={host}>
				<ReforgePanel model={model} />
			</SimHostProvider>
		</PortalContainerContext>,
	);

beforeEach(() => {
	trackPageView.mockClear();
	runState.isRunning = false;
});

describe('ReforgePanel', () => {
	it('renders the run button and the settings trigger as one sidebar action pair', () => {
		const { container } = renderPanel();
		const [run, settings] = [...container.querySelectorAll('button')];

		expect([...run.classList]).toEqual(expect.arrayContaining(['ui-button', 'ui-button-primary']));
		expect(run.getAttribute('data-testid')).toBe('suggest-reforges-action-button');
		expect(run.hasAttribute('data-sidebar-action')).toBe(true);
		expect([...settings.classList]).toEqual(expect.arrayContaining(['ui-button', 'ui-button-primary']));
		expect(settings.getAttribute('data-testid')).toBe('suggest-reforges-button-settings');
		expect(settings.hasAttribute('data-sidebar-action')).toBe(true);
		expect(settings.querySelector('.fa-gear')).not.toBeNull();
	});

	it('shows no spinner on the run button at rest', () => {
		const { container } = renderPanel();
		const run = container.querySelector('button')!;

		expect(run.hasAttribute('aria-busy')).toBe(false);
		expect(container.querySelector('[data-testid="loader"]')).toBeNull();
	});

	it('disables the run button and shows its spinner while a gem optimization run is in flight', () => {
		runState.isRunning = true;
		const { container } = renderPanel();
		const run = container.querySelector('button')!;

		expect(run.disabled).toBe(true);
		expect(run.getAttribute('aria-busy')).toBe('true');
		expect(run.querySelector('[data-testid="loader"]')).not.toBeNull();
	});

	it('builds no popover body and tracks no page view until the trigger is used', () => {
		const { container } = renderPanel();

		expect(container.querySelector('[data-testid="settings-body"]')).toBeNull();
		expect(document.querySelector('[data-testid="settings-body"]')).toBeNull();
		expect(trackPageView).not.toHaveBeenCalled();
	});

	it('mounts no progress dialog until a run starts', () => {
		renderPanel();

		expect(document.querySelector('[data-testid="progress-tracker-dialog"]')).toBeNull();
	});

	it('builds the body and tracks the page view when the trigger opens the popover', async () => {
		const { container } = renderPanel();
		const settings = [...container.querySelectorAll('button')][1];

		await act(async () => void fireEvent.click(settings));

		expect(host.rootElem.querySelector('[data-testid="settings-body"]')).not.toBeNull();
		expect(trackPageView).toHaveBeenCalledWith('Reforge Settings', 'reforge-settings');
	});
});
