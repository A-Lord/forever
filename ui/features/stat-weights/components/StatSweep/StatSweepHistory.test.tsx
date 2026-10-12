import i18n from '@i18n/config';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { StatSweepHistory } from './StatSweepHistory';

const RUNS = [
	{ id: 'hit', label: 'Hit Rating ±100', createdAt: Date.UTC(2026, 9, 12, 14, 3), overlaid: true, colorSlot: 0 },
	{ id: 'crit', label: 'Crit Rating ±50', createdAt: Date.UTC(2026, 9, 10, 18, 15), overlaid: false },
];

describe('StatSweepHistory', () => {
	it('ticks each run by its overlay state and reports a toggle with the run id', async () => {
		const onOverlayChange = vi.fn();
		render(<StatSweepHistory runs={RUNS} onOverlayChange={onOverlayChange} onDelete={vi.fn()} />);

		const [hit, crit] = screen.getAllByRole('checkbox');
		expect([hit.getAttribute('aria-checked'), crit.getAttribute('aria-checked')]).toEqual(['true', 'false']);

		await userEvent.click(crit);
		expect(onOverlayChange).toHaveBeenCalledWith('crit', true);
	});

	it('deletes the run whose button is pressed', async () => {
		const onDelete = vi.fn();
		render(<StatSweepHistory runs={RUNS} onOverlayChange={vi.fn()} onDelete={onDelete} />);

		const buttons = screen.getAllByRole('button', { name: i18n.t('sidebar.buttons.stat_weights.sweep.history.delete', { label: RUNS[1].label }) });
		await userEvent.click(buttons[buttons.length - 1]);
		expect(onDelete).toHaveBeenCalledWith('crit');
	});
});
