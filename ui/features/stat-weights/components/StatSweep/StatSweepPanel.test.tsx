import i18n from '@i18n/config';
import { fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { StatSweepPanel, type StatSweepPanelProps } from './StatSweepPanel';

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`sidebar.buttons.stat_weights.sweep.${key}`, options);

const props = (overrides: Partial<StatSweepPanelProps> = {}): StatSweepPanelProps => ({
	statOptions: [{ value: 'hit', label: 'Hit Rating' }],
	stat: 'hit',
	onStatChange: vi.fn(),
	range: 100,
	onRangeChange: vi.fn(),
	step: 10,
	onStepChange: vi.fn(),
	precision: null,
	chart: null,
	history: null,
	running: false,
	onRun: vi.fn(),
	onCancel: vi.fn(),
	...overrides,
});

describe('StatSweepPanel', () => {
	it('runs a sweep while idle', async () => {
		const onRun = vi.fn();
		render(<StatSweepPanel {...props({ onRun })} />);

		expect(screen.queryByRole('button', { name: t('cancel') })).toBeNull();
		await userEvent.click(screen.getByRole('button', { name: t('run') }));
		expect(onRun).toHaveBeenCalledOnce();
	});

	it('shows the step and a cancel button while running, and locks the sweep inputs', async () => {
		const onCancel = vi.fn();
		render(<StatSweepPanel {...props({ running: true, progress: { step: 7, steps: 21 }, onCancel })} />);

		expect(screen.getByText(t('progress', { current: 7, total: 21 }))).toBeTruthy();
		expect(screen.queryByRole('button', { name: t('run') })).toBeNull();
		expect(screen.getByLabelText<HTMLInputElement>(t('range')).disabled).toBe(true);
		await userEvent.click(screen.getByRole('button', { name: t('cancel') }));
		expect(onCancel).toHaveBeenCalledOnce();
	});

	it('reports an entered range and step', () => {
		const onRangeChange = vi.fn();
		const onStepChange = vi.fn();
		render(<StatSweepPanel {...props({ onRangeChange, onStepChange })} />);

		fireEvent.change(screen.getByLabelText(t('range')), { target: { value: '50' } });
		fireEvent.change(screen.getByLabelText(t('step')), { target: { value: '5' } });
		expect(onRangeChange).toHaveBeenCalledWith(50);
		expect(onStepChange).toHaveBeenCalledWith(5);
	});
});
