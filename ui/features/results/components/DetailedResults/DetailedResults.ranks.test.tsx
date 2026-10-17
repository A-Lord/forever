import { Emitter } from '@sim/state/events';
import { createSimStore } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { metricRows, spellActionsRun, stubSpellIcons } from '@testing/results-metrics';
import { act, fireEvent, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ResultChannel } from '../../model/result_channel';

vi.mock('@i18n/localization', async () => (await import('@testing/sim/localization')).localizationStub);
vi.mock('../MetricsTable/MetricsActionCell', () => ({ MetricsActionCell: ({ name }: { name: string }) => <span>{name}</span> }));
vi.mock('../ResultsFilter', async importOriginal => ({
	...(await importOriginal<typeof import('../ResultsFilter')>()),
	ResultsFilter: () => <span data-testid="results-filter-root" />,
}));
vi.mock('../LogRunner', () => ({ LogRunner: () => <div /> }));
vi.mock('../Timeline', () => ({ Timeline: () => <div /> }));
vi.mock('../ToplineResults', () => ({ ToplineResults: () => <div /> }));
vi.mock('../DpsHistogram', () => ({ DpsHistogram: () => <div /> }));
vi.mock('../CastMetricsTable', () => ({ CastMetricsTable: () => <div /> }));
vi.mock('../ResourceMetricsTable', () => ({ ResourceMetricsTable: () => <div /> }));
vi.mock('../AuraMetricsTable', () => ({ AuraMetricsTable: () => <div /> }));

import { DetailedResults } from './DetailedResults';

const FIREBALL_1 = 133;
const FIREBALL_2 = 143;
const FROSTBOLT = 116;

const RANK = 'common.action_tags.rank';

let resultChannel: ResultChannel;
let currentChangeEmitter: Emitter<void>;

const renderPane = async () => {
	const run = spellActionsRun([
		{ spellId: FIREBALL_1, damage: 100 },
		{ spellId: FIREBALL_2, damage: 300 },
		{ spellId: FROSTBOLT, damage: 50 },
	]);
	const view = renderWithHost(<DetailedResults resultsManager={{ currentChangeEmitter, getRunData: () => ({ run }) } as never} />, {
		sim: {
			store: createSimStore(),
			getShowDamageMetrics: () => true,
			getShowThreatMetrics: () => false,
			getShowHealingMetrics: () => false,
			getShowExperimental: () => false,
			getFixedRngSeed: () => 0,
			setFixedRngSeed: () => {},
		},
		resultChannel,
		disabled: false,
		runSingleIteration: () => Promise.resolve(undefined),
	});
	await act(async () => currentChangeEmitter.emit());
	return view;
};

const damageRows = (container: HTMLElement) =>
	metricRows(container.querySelector<HTMLElement>('[data-testid="damage-metrics-root"]')!).map(row => [
		row.cells[0].textContent,
		row.hasAttribute('data-child'),
	]);

const toggle = (container: HTMLElement) => within(container.querySelector<HTMLElement>('[data-testid="results-filter"]')!).getByRole('checkbox');

beforeEach(() => {
	window.localStorage.clear();
	resultChannel = new ResultChannel();
	currentChangeEmitter = new Emitter<void>();
	stubSpellIcons({
		[FIREBALL_1]: { name: 'Fireball', rank: 1 },
		[FIREBALL_2]: { name: 'Fireball', rank: 2 },
		[FROSTBOLT]: { name: 'Frostbolt' },
	});
});

describe('DetailedResults rank grouping', () => {
	it('opens with the ranks of a spell under one row named without a rank, and lists every rank as its own row once the toggle is cleared', async () => {
		const { container } = await renderPane();

		expect(toggle(container).getAttribute('aria-checked')).toBe('true');
		await waitFor(() =>
			expect(damageRows(container)).toEqual([
				['Fireball', false],
				[`Fireball (${RANK})`, true],
				[`Fireball (${RANK})`, true],
				['Frostbolt', false],
			]),
		);

		await act(async () => {
			fireEvent.click(toggle(container));
		});

		expect(damageRows(container)).toEqual([
			[`Fireball (${RANK})`, false],
			[`Fireball (${RANK})`, false],
			['Frostbolt', false],
		]);
	});

	it('remembers a cleared toggle for the next time the pane opens', async () => {
		const first = await renderPane();
		await waitFor(() => expect(damageRows(first.container)).toHaveLength(4));
		await act(async () => {
			fireEvent.click(toggle(first.container));
		});
		first.unmount();

		const { container } = await renderPane();
		expect(toggle(container).getAttribute('aria-checked')).toBe('false');
		await waitFor(() => expect(damageRows(container)).toHaveLength(3));
		expect(damageRows(container).every(([, child]) => !child)).toBe(true);
	});
});
