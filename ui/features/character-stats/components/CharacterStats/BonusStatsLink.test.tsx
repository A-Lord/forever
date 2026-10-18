import { Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { Stats, UnitStat } from '@sim/proto/stats';
import { createSimStore } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { BonusStatsLink } from './BonusStatsLink';

const STRENGTH = UnitStat.fromStat(Stat.StatStrength);

describe('BonusStatsLink', () => {
	it('opens a picker labelled with its stat, showing the bonus on it, and writes the entry', async () => {
		let bonusStats = new Stats().withStat(Stat.StatStrength, 5);
		const setBonusStats = vi.fn((next: Stats) => (bonusStats = next));
		const player = { storeKey: 0, sim: { store: createSimStore() }, getBonusStats: () => bonusStats, setBonusStats };
		renderWithHost(<BonusStatsLink unitStat={STRENGTH} />, { player });

		const button = screen.getByTestId('add-bonus-stats');
		expect(button.getAttribute('aria-label')).toMatch(/^sidebar\.character_stats\.bonus_prefix /);
		await act(async () => {
			fireEvent.click(button);
		});

		const input = (await screen.findByLabelText(button.getAttribute('aria-label')!, { selector: 'input' })) as HTMLInputElement;
		expect(input.id).toBe(`character-bonus-stat-${STRENGTH.getKey()}`);
		expect(input.value).toBe('5');
		await act(async () => {
			input.value = '20';
			fireEvent.change(input);
		});
		expect(setBonusStats).toHaveBeenCalledTimes(1);
		expect(setBonusStats.mock.calls[0][0].getStat(Stat.StatStrength)).toBe(20);
	});

	it.each([
		[Stat.StatHitRating, 20, 2, 2.5, 25],
		[Stat.StatCritRating, 14, 1, 1.5, 21],
		[Stat.StatHasteRating, 10, 1, 2.5, 25],
		[Stat.StatExpertiseRating, 10, 1, 0.65, 6.5],
	] as const)('accepts a percent bonus for rating stat %s', async (stat, initialRating, initialPercent, percent, expectedRating) => {
		let bonusStats = new Stats().withStat(stat, initialRating);
		const setBonusStats = vi.fn((next: Stats) => (bonusStats = next));
		const unitStat = UnitStat.fromStat(stat);
		const player = { storeKey: 0, sim: { store: createSimStore() }, getBonusStats: () => bonusStats, setBonusStats };
		renderWithHost(<BonusStatsLink unitStat={unitStat} />, { player });

		await act(async () => {
			fireEvent.click(screen.getByTestId('add-bonus-stats'));
		});

		const percentButton = screen.getByRole('button', { name: i18n.t('sidebar.character_stats.percent_label') });
		expect(percentButton.getAttribute('aria-pressed')).toBe('true');
		const input = screen.getByRole('textbox') as HTMLInputElement;
		expect(Number(input.value)).toBeCloseTo(initialPercent);

		await act(async () => {
			fireEvent.click(screen.getByRole('button', { name: i18n.t('sidebar.character_stats.rating_suffix') }));
		});
		expect(screen.getByRole('button', { name: i18n.t('sidebar.character_stats.rating_suffix') }).getAttribute('aria-pressed')).toBe('true');
		expect((screen.getByRole('textbox') as HTMLInputElement).value).toBe(String(initialRating));

		await act(async () => {
			fireEvent.click(percentButton);
		});
		const percentInput = screen.getByRole('textbox') as HTMLInputElement;
		percentInput.value = String(percent);
		await act(async () => {
			fireEvent.change(percentInput);
		});

		expect(setBonusStats).toHaveBeenCalledTimes(1);
		expect(setBonusStats.mock.calls[0][0].getStat(stat)).toBeCloseTo(expectedRating);
	});

	it('keeps the picker open and commits the edited value when switching units', async () => {
		let bonusStats = new Stats().withStat(Stat.StatHitRating, 20);
		const setBonusStats = vi.fn((next: Stats) => (bonusStats = next));
		const unitStat = UnitStat.fromStat(Stat.StatHitRating);
		const player = { storeKey: 0, sim: { store: createSimStore() }, getBonusStats: () => bonusStats, setBonusStats };
		renderWithHost(<BonusStatsLink unitStat={unitStat} />, { player });

		await act(async () => {
			fireEvent.click(screen.getByTestId('add-bonus-stats'));
		});

		const percentInput = screen.getByRole('textbox');
		await userEvent.clear(percentInput);
		await userEvent.type(percentInput, '2.5');
		await userEvent.click(screen.getByRole('button', { name: i18n.t('sidebar.character_stats.rating_suffix') }));

		await waitFor(() => expect((screen.getByRole('textbox') as HTMLInputElement).value).toBe('25'));
		expect(setBonusStats).toHaveBeenCalledTimes(1);
		expect(bonusStats.getStat(Stat.StatHitRating)).toBeCloseTo(25);
	});
});
