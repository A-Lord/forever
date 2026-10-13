import i18n from '@i18n/config';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { StatSweepPanel, type StatSweepPanelProps } from './StatSweepPanel';

const t = (key: string) => i18n.t(`sidebar.buttons.stat_sweep.${key}`);

const props = (overrides: Partial<StatSweepPanelProps> = {}): StatSweepPanelProps => ({
	statOptions: [
		{ value: 'hit', label: 'Hit Rating' },
		{ value: 'crit', label: 'Crit Rating' },
		{ value: 'haste', label: 'Haste Rating' },
	],
	stats: ['crit'],
	onStatsChange: vi.fn(),
	range: 100,
	onRangeChange: vi.fn(),
	step: 10,
	onStepChange: vi.fn(),
	precision: null,
	chart: null,
	history: null,
	running: false,
	...overrides,
});

describe('StatSweepPanel', () => {
	it('locks the setup inputs while running', () => {
		render(<StatSweepPanel {...props({ running: true })} />);

		expect(screen.getByLabelText<HTMLInputElement>(t('range')).disabled).toBe(true);
		expect(screen.getByLabelText<HTMLInputElement>(t('step')).disabled).toBe(true);
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

	it('shows the picked stats as chips and reports a pick and a removal by stat key', () => {
		const onStatsChange = vi.fn();
		render(<StatSweepPanel {...props({ stats: ['haste', 'crit'], onStatsChange })} />);
		expect(screen.getAllByTestId('multi-combo-box-chip').map(chip => chip.textContent)).toEqual(['Crit Rating', 'Haste Rating']);

		fireEvent.click(screen.getAllByRole('button', { name: t('stats.remove') })[0]);
		expect(onStatsChange).toHaveBeenLastCalledWith(['haste']);

		fireEvent.click(screen.getByTestId('multi-combo-box-trigger'));
		fireEvent.click(within(screen.getByTestId('multi-combo-box-list')).getAllByRole('option')[0]);
		expect(onStatsChange).toHaveBeenLastCalledWith(['crit', 'haste', 'hit']);
	});
});
