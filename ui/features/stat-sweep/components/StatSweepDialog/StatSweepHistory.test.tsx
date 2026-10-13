import i18n from '@i18n/config';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { StatSweepHistory, type StatSweepHistoryEntry } from './StatSweepHistory';

const RUNS: StatSweepHistoryEntry[] = [
	{ id: 'hit', label: 'Hit Rating ±100', createdAt: Date.UTC(2026, 9, 12, 14, 3), overlaid: true, colorSlot: 0, metric: 'dps' },
	{ id: 'crit', label: 'Crit Rating ±50', createdAt: Date.UTC(2026, 9, 10, 18, 15), overlaid: false, metric: 'tps' },
];

const exportButton = () => screen.getByRole<HTMLButtonElement>('button', { name: i18n.t('sidebar.buttons.stat_sweep.export_csv') });

describe('StatSweepHistory', () => {
	it('ticks each run by its overlay state and reports a toggle with the run id', async () => {
		const onOverlayChange = vi.fn();
		render(<StatSweepHistory runs={RUNS} onOverlayChange={onOverlayChange} onDelete={vi.fn()} />);

		const [hit, crit] = screen.getAllByRole('checkbox');
		expect([hit.getAttribute('aria-checked'), crit.getAttribute('aria-checked')]).toEqual(['true', 'false']);

		await userEvent.click(crit);
		expect(onOverlayChange).toHaveBeenCalledWith('crit', true);
	});

	it('tags a run that another metric than DPS drove with that metric', () => {
		render(<StatSweepHistory runs={RUNS} onOverlayChange={vi.fn()} onDelete={vi.fn()} />);

		expect(screen.getAllByTestId('stat-sweep-history-metric').map(tag => tag.textContent)).toEqual([i18n.t('sidebar.results.metrics.tps.label')]);
	});

	it('deletes the run whose button is pressed', async () => {
		const onDelete = vi.fn();
		render(<StatSweepHistory runs={RUNS} onOverlayChange={vi.fn()} onDelete={onDelete} />);

		const buttons = screen.getAllByRole('button', { name: i18n.t('sidebar.buttons.stat_sweep.history.delete', { label: RUNS[1].label }) });
		await userEvent.click(buttons[buttons.length - 1]);
		expect(onDelete).toHaveBeenCalledWith('crit');
	});

	it('exports on Export CSV, and disables it with nothing to export', async () => {
		const onExportCsv = vi.fn();
		const { rerender } = render(<StatSweepHistory runs={RUNS} onOverlayChange={vi.fn()} onDelete={vi.fn()} onExportCsv={onExportCsv} />);

		await userEvent.click(exportButton());
		expect(onExportCsv).toHaveBeenCalledTimes(1);

		rerender(<StatSweepHistory runs={RUNS} onOverlayChange={vi.fn()} onDelete={vi.fn()} />);
		expect(exportButton().disabled).toBe(true);
	});
});
