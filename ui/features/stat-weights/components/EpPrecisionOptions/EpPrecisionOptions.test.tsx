import i18n from '@i18n/config';
import type { EpPrecisionSettings } from '@sim/settings/stat_weight_settings';
import { fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { EpPrecisionOptions } from './EpPrecisionOptions';

const t = (key: string) => i18n.t(`sidebar.buttons.stat_weights.precision.${key}`);
const ITERATIONS: EpPrecisionSettings = { precisionMode: 'iterations', iterations: 0, targetEpCi: 0.01, maxIterationMultiplier: 0 };

describe('EpPrecisionOptions', () => {
	it('shows the iteration count in iterations mode and the target fields alone in target mode, and reports a mode switch', async () => {
		const onChange = vi.fn();
		const { rerender } = render(<EpPrecisionOptions value={ITERATIONS} onChange={onChange} defaultIterations={10000} />);
		expect(screen.queryByLabelText(t('target_ep_ci'))).toBeNull();
		expect(screen.getByLabelText(t('iterations'))).toBeTruthy();

		await userEvent.click(screen.getByRole('button', { name: t('mode_target') }));
		expect(onChange).toHaveBeenCalledWith({ precisionMode: 'target' });

		rerender(<EpPrecisionOptions value={{ ...ITERATIONS, precisionMode: 'target' }} onChange={onChange} defaultIterations={10000} />);
		expect(screen.getByLabelText(t('target_ep_ci'))).toBeTruthy();
		expect(screen.getByLabelText(t('max_multiplier'))).toBeTruthy();
		expect(screen.queryByLabelText(t('iterations'))).toBeNull();
	});

	it('leaves an unset iteration count empty with the sim setting as placeholder, and reports an entered count', () => {
		const onChange = vi.fn();
		render(<EpPrecisionOptions value={ITERATIONS} onChange={onChange} defaultIterations={10000} />);
		const input = screen.getByLabelText<HTMLInputElement>(t('iterations'));
		expect([input.value, input.placeholder]).toEqual(['', '10000']);

		fireEvent.change(input, { target: { value: '4000' } });
		expect(onChange).toHaveBeenCalledWith({ iterations: 4000 });
	});

	it('leaves an unset multiplier empty as no cap, and reports an entered one', () => {
		const onChange = vi.fn();
		render(<EpPrecisionOptions value={{ ...ITERATIONS, precisionMode: 'target' }} onChange={onChange} defaultIterations={10000} />);
		const input = screen.getByLabelText<HTMLInputElement>(t('max_multiplier'));
		expect([input.value, input.placeholder]).toEqual(['', t('max_multiplier_none')]);

		fireEvent.change(input, { target: { value: '4' } });
		expect(onChange).toHaveBeenCalledWith({ maxIterationMultiplier: 4 });
	});

	it('keeps four decimals of the target', () => {
		const onChange = vi.fn();
		render(<EpPrecisionOptions value={{ ...ITERATIONS, precisionMode: 'target', targetEpCi: 0.0025 }} onChange={onChange} defaultIterations={10000} />);
		const input = screen.getByLabelText<HTMLInputElement>(t('target_ep_ci'));
		expect(input.value).toBe('0.0025');

		fireEvent.change(input, { target: { value: '0.0075' } });
		expect(onChange).toHaveBeenCalledWith({ targetEpCi: 0.0075 });
	});
});
