import i18n from '@i18n/config';
import { createSimStore, patchSlice } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { act, fireEvent, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { DpsCurveAction } from './DpsCurveAction';

vi.mock('./DpsCurveDialog/DpsCurveDialog', () => ({ DpsCurveDialog: ({ open }: { open: boolean }) => (open ? <div data-testid="dps-curve-dialog" /> : null) }));
vi.mock('../../../tracking/utils', () => ({ trackPageView: vi.fn() }));

const renderAction = (simDisabled = false) => {
	const store = createSimStore();
	renderWithHost(<DpsCurveAction settings={{} as never} />, { sim: { store, waitForInit: () => Promise.resolve() }, simDisabled });
	return store;
};
const action = () => screen.queryByRole('button', { name: i18n.t('sidebar.buttons.dps_curve.title') });

describe('DpsCurveAction', () => {
	it('shows only while Show advanced options is on, as an icon button named for the tool', () => {
		const store = renderAction();
		expect(action()).toBeNull();

		act(() => patchSlice(store, 'ui', { showAdvancedOptions: true }));
		expect(action()).not.toBeNull();
		expect(action()!.textContent).toBe('');
	});

	it('opens the DPS curve dialog', async () => {
		const store = renderAction();
		act(() => patchSlice(store, 'ui', { showAdvancedOptions: true }));
		await act(async () => {});

		fireEvent.click(action()!);

		expect(screen.getByTestId('dps-curve-dialog')).toBeTruthy();
	});

	it('stays hidden in a gear planner, which runs no sims', () => {
		const store = renderAction(true);
		act(() => patchSlice(store, 'ui', { showAdvancedOptions: true }));

		expect(action()).toBeNull();
	});
});
