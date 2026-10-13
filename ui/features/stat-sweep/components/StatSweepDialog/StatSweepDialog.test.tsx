// The sim and the history store are stubbed; the settings, the store and the run flags are real.
import { ErrorOutcomeType } from '@generated/proto/api';
import { Class, Spec, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import type { StatSweepRecord } from '@sim/cache/stat_sweep_history';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { Stats, UnitStat } from '@sim/proto/stats';
import { StatSweepSettings } from '@sim/settings/stat_sweep_settings';
import type { MetricSample, StatSweepSample, StatSweepStepOptions } from '@sim/sim';
import { SimRuns } from '@sim/sim_runs';
import { createSimStore, patchSlice, PLAYER_FIELDS, seedKeyed, type SimStore, zeroVersions } from '@sim/state/sim_store';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { PortalContainerContext } from '@ui-kit/hooks/usePortalContainer';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { iterationsForCi, STAT_SWEEP_PROBE_ITERATIONS } from '../../model/stat_sweep';
import { StatSweepDialog } from './StatSweepDialog';

// happy-dom has no IndexedDB.
const history = vi.hoisted(() => ({
	listRuns: vi.fn(async () => [] as StatSweepRecord[]),
	saveRun: vi.fn(async (_run: StatSweepRecord) => {}),
	deleteRun: vi.fn(async () => {}),
}));
vi.mock('@sim/cache/stat_sweep_history', () => history);
vi.mock('chart.js', async () => (await import('@testing/chart-js-mock')).chartJsMock);
const download = vi.hoisted(() => vi.fn());
vi.mock('@ui-kit/utils/dom', async importOriginal => ({ ...(await importOriginal<typeof import('@ui-kit/utils/dom')>()), downloadString: download }));

const STRENGTH = UnitStat.fromStat(Stat.StatStrength);
const AGILITY = UnitStat.fromStat(Stat.StatAgility);

const flat = (avg: number, stdev: number, n: number): MetricSample => ({ avg, stdev, n, sumSq: n * (stdev ** 2 + avg ** 2) });
// Each metric moves with `dps`, and TPS has the widest spread.
const sample = (iterations: number, dps = 1000, tmi = dps / 100): StatSweepSample => ({
	dps: flat(dps, 100, iterations),
	hps: flat(10, 5, iterations),
	tps: flat(2 * dps, 400, iterations),
	dtps: flat(3000 - dps, 300, iterations),
	tmi: flat(tmi, 2, iterations),
});
// One DPS per bonus point of either stat, so every step's DPS differs.
const dpsOf = ({ bonusStats }: StatSweepStepOptions) => 1000 + bonusStats.getStat(Stat.StatStrength) + bonusStats.getStat(Stat.StatAgility);

const rootElem = document.createElement('div');
let store: SimStore;
let settings: StatSweepSettings;
let simulateStatSweepStep: ReturnType<typeof vi.fn<(options: StatSweepStepOptions) => Promise<object>>>;
let abortType: ReturnType<typeof vi.fn>;
let host: any;

const setup = () => {
	store = createSimStore();
	const storage = new Map<string, string>();
	const env = { storage: { getItem: (key: string) => storage.get(key) ?? null, setItem: (key: string, value: string) => storage.set(key, value) } };
	seedKeyed(store, 'players', 0, { v: zeroVersions(PLAYER_FIELDS) } as never);
	abortType = vi.fn().mockResolvedValue(undefined);
	simulateStatSweepStep = vi.fn(async (options: StatSweepStepOptions) => sample(options.iterations, dpsOf(options)));
	const sim = {
		store,
		env,
		runs: new SimRuns(store, { abortType: abortType as unknown as (mask: number) => Promise<void> }),
		getIterations: () => 1000,
		nextRngSeed: () => 7,
		getShowDamageMetrics: () => store.getState().ui.showDamageMetrics,
		getShowHealingMetrics: () => store.getState().ui.showHealingMetrics,
		getShowThreatMetrics: () => store.getState().ui.showThreatMetrics,
		// A player with no stats of its own: the final stats are the bonus.
		getFinalStatsWithBonus: async (bonusStats: Stats) => bonusStats,
		simulateStatSweepStep,
	};
	const player = { storeKey: 0, sim, getClass: () => Class.ClassWarrior, getSpec: () => Spec.SpecUnknown };
	settings = new StatSweepSettings(player as never, 'stat-sweep-test');
	host = {
		player,
		sim,
		individualConfig: { epStats: [Stat.StatStrength, Stat.StatAgility], epPseudoStats: [], epReferenceStat: Stat.StatStrength },
	};
};

const renderDialog = () => {
	document.body.appendChild(rootElem);
	const onOpenChange = vi.fn();
	render(
		<PortalContainerContext value={rootElem}>
			<SimHostProvider host={host}>
				<StatSweepDialog open onOpenChange={onOpenChange} settings={settings} />
			</SimHostProvider>
		</PortalContainerContext>,
	);
	return onOpenChange;
};

const runButton = () => screen.getByRole<HTMLButtonElement>('button', { name: 'sidebar.buttons.stat_sweep.run' });
const tracker = () => rootElem.querySelector('[data-testid="stat-sweep-progress-dialog"]');
const bonusOf = (call: [StatSweepStepOptions], stat: Stat) => call[0].bonusStats.getStat(stat);
const metricLabel = (metric: string) => i18n.t(`sidebar.results.metrics.${metric}.label`);
const metricToggle = () => within(rootElem.querySelector<HTMLElement>('[data-testid="stat-sweep-metric"]')!);
const offeredMetrics = () =>
	metricToggle()
		.getAllByRole('button')
		.map(button => button.textContent);
const pressedMetric = () =>
	metricToggle()
		.getAllByRole('button')
		.find(button => button.getAttribute('aria-pressed') === 'true')?.textContent;

beforeEach(() => {
	setup();
	history.saveRun.mockClear();
	download.mockClear();
});

describe('StatSweepDialog', () => {
	it('runs one plain sim per step at one seed with the step as bonus stats, and saves every metric of each step', async () => {
		settings.setSetup({ range: 20, step: 10 });
		settings.setPrecision({ iterations: 3000 });
		renderDialog();

		await act(async () => {
			fireEvent.click(runButton());
		});

		const calls = simulateStatSweepStep.mock.calls;
		expect(calls.map(call => bonusOf(call, Stat.StatStrength))).toEqual([0, 10, 20]);
		expect(calls.map(call => [call[0].iterations, call[0].seed])).toEqual([
			[3000, 7],
			[3000, 7],
			[3000, 7],
		]);
		expect(history.saveRun).toHaveBeenCalledTimes(1);
		const saved = history.saveRun.mock.calls[0][0];
		expect(saved.metric).toBe('dps');
		expect(saved.points.map(point => [point.statValue, point.metrics.dps.avg, point.metrics.dps.n])).toEqual([
			[0, 1000, 3000],
			[10, 1010, 3000],
			[20, 1020, 3000],
		]);
		expect(saved.points[1].metrics).toEqual({
			dps: { avg: 1010, stdev: 100, n: 3000 },
			hps: { avg: 10, stdev: 5, n: 3000 },
			tps: { avg: 2020, stdev: 400, n: 3000 },
			dtps: { avg: 1990, stdev: 300, n: 3000 },
			tmi: { avg: 10.1, stdev: 2, n: 3000 },
		});
		expect(saved.cappedAt).toBeUndefined();
	});

	it('shows the band key while the spread is on, and hides it with the spread checkbox, which stores the pick', () => {
		renderDialog();
		const bandKey = () => screen.queryByText(i18n.t('sidebar.buttons.stat_sweep.band_key', { metric: metricLabel('dps') }));
		expect(bandKey()).toBeTruthy();

		fireEvent.click(within(screen.getByTestId('stat-sweep-show-spread')).getByRole('checkbox'));
		expect(settings.showSpread).toBe(false);
		expect(bandKey()).toBeNull();

		fireEvent.click(within(screen.getByTestId('stat-sweep-show-spread')).getByRole('checkbox'));
		expect(settings.showSpread).toBe(true);
		expect(bandKey()).toBeTruthy();
	});

	it('offers DPS and TPS, adds DTPS and TMI with the threat and tank options, and HPS with the healing metrics', () => {
		renderDialog();
		expect(offeredMetrics()).toEqual([metricLabel('dps'), metricLabel('tps')]);

		act(() => patchSlice(store, 'ui', { showThreatMetrics: true }));
		expect(offeredMetrics()).toEqual(['dps', 'tps', 'dtps', 'tmi'].map(metricLabel));

		act(() => patchSlice(store, 'ui', { showHealingMetrics: true }));
		expect(offeredMetrics()).toEqual(['dps', 'hps', 'tps', 'dtps', 'tmi'].map(metricLabel));
	});

	it('stores the picked metric, and shows DPS while the stored one is not offered', () => {
		act(() => patchSlice(store, 'ui', { showThreatMetrics: true }));
		renderDialog();

		fireEvent.click(metricToggle().getByRole('button', { name: metricLabel('tmi') }));
		expect(settings.metric).toBe('tmi');
		expect(pressedMetric()).toBe(metricLabel('tmi'));

		act(() => patchSlice(store, 'ui', { showThreatMetrics: false }));
		expect(pressedMetric()).toBe(metricLabel('dps'));
		expect(settings.metric).toBe('tmi');
	});

	it('stops a stat going up once a step gives every metric of the step below it, and saves where it capped', async () => {
		simulateStatSweepStep.mockImplementation(async options => sample(options.iterations, Math.min(dpsOf(options), 1020)));
		settings.setSetup({ range: 50, step: 10 });
		renderDialog();

		await act(async () => {
			fireEvent.click(runButton());
		});

		expect(simulateStatSweepStep.mock.calls.map(call => bonusOf(call, Stat.StatStrength))).toEqual([0, 10, 20, 30]);
		const saved = history.saveRun.mock.calls[0][0];
		expect(saved.points.map(point => point.statValue)).toEqual([0, 10, 20, 30]);
		expect(saved.cappedAt).toBe(20);
		expect(screen.getByText('sidebar.buttons.stat_sweep.capped_label')).toBeTruthy();
	});

	it('keeps a stat going up while it changes only TMI, whatever metric the run follows, and records that metric on the run', async () => {
		// Every metric but TMI stays put across the steps.
		simulateStatSweepStep.mockImplementation(async options => sample(options.iterations, 1000, 10 + dpsOf(options) / 100));
		settings.setSetup({ range: 40, step: 10 });
		settings.setMetric('tps');
		renderDialog();

		await act(async () => {
			fireEvent.click(runButton());
		});

		expect(simulateStatSweepStep.mock.calls.map(call => bonusOf(call, Stat.StatStrength))).toEqual([0, 10, 20, 30, 40]);
		const saved = history.saveRun.mock.calls[0][0];
		expect(saved.metric).toBe('tps');
		expect(saved.cappedAt).toBeUndefined();
		expect(screen.getByTestId('stat-sweep-history-metric').textContent).toBe(metricLabel('tps'));
	});

	it('runs each picked stat on its own in the order offered, saving each as its own run', async () => {
		settings.setSetup({ stats: [AGILITY, STRENGTH], range: 10, step: 10 });
		renderDialog();

		await act(async () => {
			fireEvent.click(runButton());
		});

		const calls = simulateStatSweepStep.mock.calls;
		expect(calls.map(call => [bonusOf(call, Stat.StatStrength), bonusOf(call, Stat.StatAgility)])).toEqual([
			[0, 0],
			[10, 0],
			[0, 0],
			[0, 10],
		]);
		const saved = history.saveRun.mock.calls.map(call => call[0]);
		expect(saved.map(run => UnitStat.fromJson(run.unitStat).getKey())).toEqual([STRENGTH.getKey(), AGILITY.getKey()]);
		expect(saved[0].colorSlot).not.toBe(saved[1].colorSlot);
	});

	it('unticks the earlier runs when a run starts, keeping every run of the new batch ticked', async () => {
		settings.setSetup({ stats: [STRENGTH], range: 10, step: 10 });
		renderDialog();
		const ticks = () =>
			[...rootElem.querySelectorAll('[id^="stat-sweep-overlay-"]')].map(
				box => box.getAttribute('aria-checked') ?? String((box as HTMLInputElement).checked),
			);

		await act(async () => {
			fireEvent.click(runButton());
		});
		expect(ticks()).toEqual(['true']);

		act(() => settings.setSetup({ stats: [AGILITY, STRENGTH] }));
		await act(async () => {
			fireEvent.click(runButton());
		});
		expect(ticks()).toEqual(['true', 'true', 'false']);
	});

	it('probes each step in Target ± mode and tops it up at the seed after the probe', async () => {
		settings.setSetup({ range: 5, step: 10 });
		settings.setPrecision({ precisionMode: 'target', targetCi: 2 });
		renderDialog();

		await act(async () => {
			fireEvent.click(runButton());
		});

		const calls = simulateStatSweepStep.mock.calls.map(call => [call[0].iterations, call[0].seed]);
		expect(calls[0]).toEqual([STAT_SWEEP_PROBE_ITERATIONS, 7]);
		expect(calls[1][1]).toBe(7 + STAT_SWEEP_PROBE_ITERATIONS);
		expect(history.saveRun.mock.calls[0][0].points[0].metrics.dps.n).toBe(STAT_SWEEP_PROBE_ITERATIONS + calls[1][0]);
	});

	it("sizes Target ± by the picked metric's spread, with the target in that metric's units", async () => {
		settings.setSetup({ range: 5, step: 10 });
		settings.setMetric('tps');
		settings.setPrecision({ precisionMode: 'target', targetCi: 20 });
		renderDialog();
		expect(rootElem.querySelector('label[for="stat-sweep-precision-target"]')!.textContent).toContain('sidebar.buttons.stat_sweep.precision.target_ci');

		await act(async () => {
			fireEvent.click(runButton());
		});

		// TPS spreads 400 a fight, where DPS spreads 100: a ± of 20 needs that many iterations of TPS.
		const total = iterationsForCi(400, 20);
		expect(simulateStatSweepStep.mock.calls.map(call => call[0].iterations)).toEqual([STAT_SWEEP_PROBE_ITERATIONS, total - STAT_SWEEP_PROBE_ITERATIONS]);
		expect(history.saveRun.mock.calls[0][0].points[0].metrics.tps.n).toBe(total);
	});

	it('shows the run in the progress dialog and cancels it there, keeping the finished steps', async () => {
		let settleSecond!: (result: object) => void;
		simulateStatSweepStep.mockResolvedValueOnce(sample(1000)).mockImplementationOnce(options => {
			options.onProgress({ completedIterations: 250, totalIterations: 1000 } as never);
			return new Promise(resolve => (settleSecond = resolve));
		});
		settings.setSetup({ range: 20, step: 10 });
		renderDialog();
		const run = runButton();

		await act(async () => {
			fireEvent.click(run);
		});
		// The second of three steps, a quarter done: the bar reads across the whole run.
		expect(tracker()!.querySelector('[data-testid="progress-tracker-modal-progress-title"]')!.textContent).toBe('sidebar.buttons.stat_sweep.progress');
		expect(tracker()!.querySelector('[data-testid="progress-tracker-modal-progress-text"]')!.textContent).toBe('1250/3000');
		expect(run.disabled).toBe(true);

		abortType.mockClear();
		await act(async () => {
			fireEvent.click(tracker()!.querySelector('[data-testid="progress-tracker-modal-cancel-btn"]')!);
		});
		expect(abortType).toHaveBeenCalledTimes(1);
		await act(async () => {
			settleSecond({ type: ErrorOutcomeType.ErrorOutcomeAborted, message: '' });
		});

		expect(tracker()).toBeNull();
		expect(simulateStatSweepStep).toHaveBeenCalledTimes(2);
		expect(history.saveRun.mock.calls[0][0].points).toHaveLength(1);
	});

	it("lists the run's stats in the progress dialog, done up to the running one, and titles it with that stat's step", async () => {
		let settleThird!: (result: object) => void;
		simulateStatSweepStep
			.mockResolvedValueOnce(sample(1000))
			.mockResolvedValueOnce(sample(1000))
			.mockImplementationOnce(options => {
				options.onProgress({ completedIterations: 250, totalIterations: 1000 } as never);
				return new Promise(resolve => (settleThird = resolve));
			});
		settings.setSetup({ stats: [AGILITY, STRENGTH], range: 10, step: 10 });
		renderDialog();

		await act(async () => {
			fireEvent.click(runButton());
		});
		const states = () => [...tracker()!.querySelectorAll('[data-testid="stat-sweep-stat"]')].map(row => row.getAttribute('data-state'));
		expect(states()).toEqual(['done', 'running']);
		expect(tracker()!.querySelector('[data-testid="progress-tracker-modal-progress-title"]')!.textContent).toBe('sidebar.buttons.stat_sweep.progress');

		await act(async () => {
			fireEvent.click(tracker()!.querySelector('[data-testid="progress-tracker-modal-cancel-btn"]')!);
		});
		expect(states()).toEqual(['done', 'running']);
		await act(async () => {
			settleThird({ type: ErrorOutcomeType.ErrorOutcomeAborted, message: '' });
		});
		expect(tracker()).toBeNull();
	});

	it('exports the drawn runs as CSV with every metric', async () => {
		settings.setSetup({ range: 10, step: 10 });
		renderDialog();
		await act(async () => {
			fireEvent.click(runButton());
		});

		fireEvent.click(screen.getByRole('button', { name: 'sidebar.buttons.stat_sweep.export_csv' }));

		expect(download).toHaveBeenCalledTimes(1);
		const [csv, fileName, mimeType] = download.mock.calls[0];
		const lines = csv.trim().split('\n');
		expect(lines).toHaveLength(3);
		expect(lines[0]).toBe('run,stat,base_total,added,total,dps,dps_stdev,hps,hps_stdev,tps,tps_stdev,dtps,dtps_stdev,tmi,tmi_stdev,samples');
		expect([fileName, mimeType]).toEqual(['wowsims-stat-sweep.csv', 'text/csv']);
	});

	it('disables Run with every stat removed', () => {
		settings.setSetup({ stats: [] });
		renderDialog();

		expect(runButton().disabled).toBe(true);
	});

	it('cancels a run when the dialog closes', async () => {
		simulateStatSweepStep.mockImplementation(() => new Promise(() => {}));
		const onOpenChange = renderDialog();
		await act(async () => {
			fireEvent.click(runButton());
		});
		abortType.mockClear();

		await act(async () => {
			fireEvent.click(rootElem.querySelector('[data-testid="stat-sweep-dialog"] [data-testid="sim-dialog-close"]')!);
		});

		expect(onOpenChange).toHaveBeenCalledWith(false);
		expect(abortType).toHaveBeenCalled();
	});
});
