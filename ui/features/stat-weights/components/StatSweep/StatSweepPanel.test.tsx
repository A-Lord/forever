import i18n from '@i18n/config';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { StatSweepPanel, type StatSweepPanelProps } from './StatSweepPanel';

const t = (key: string) => i18n.t(`sidebar.buttons.stat_weights.sweep.${key}`);

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
	...overrides,
});

describe('StatSweepPanel', () => {
	it('locks the sweep inputs while running', () => {
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
});
