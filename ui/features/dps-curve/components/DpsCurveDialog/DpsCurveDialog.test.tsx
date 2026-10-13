// The sim and the history store are stubbed; the settings, the store and the run flags are real.
import { ErrorOutcomeType } from '@generated/proto/api';
import { Class, Spec, Stat } from '@generated/proto/common';
import type { DpsCurveRecord } from '@sim/cache/dps_curve_history';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { Stats, UnitStat } from '@sim/proto/stats';
import { DpsCurveSettings } from '@sim/settings/dps_curve_settings';
import type { DpsSimOptions } from '@sim/sim';
import { SimRuns } from '@sim/sim_runs';
import { createSimStore, PLAYER_FIELDS, seedKeyed, zeroVersions } from '@sim/state/sim_store';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { PortalContainerContext } from '@ui-kit/hooks/usePortalContainer';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { DPS_CURVE_PROBE_ITERATIONS } from '../../model/dps_curve';
import { DpsCurveDialog } from './DpsCurveDialog';

// happy-dom has no IndexedDB.
const history = vi.hoisted(() => ({
	listRuns: vi.fn(async () => [] as DpsCurveRecord[]),
	saveRun: vi.fn(async (_run: DpsCurveRecord) => {}),
	deleteRun: vi.fn(async () => {}),
}));
vi.mock('@sim/cache/dps_curve_history', () => history);
vi.mock('chart.js', async () => (await import('@testing/chart-js-mock')).chartJsMock);
const download = vi.hoisted(() => vi.fn());
vi.mock('@ui-kit/utils/dom', async importOriginal => ({ ...(await importOriginal<typeof import('@ui-kit/utils/dom')>()), downloadString: download }));

const STRENGTH = UnitStat.fromStat(Stat.StatStrength);
const AGILITY = UnitStat.fromStat(Stat.StatAgility);

const sample = (iterations: number, avg = 1000) => ({ avg, stdev: 100, n: iterations, sumSq: iterations * (100 ** 2 + avg ** 2) });
// One DPS per bonus point of either stat, so every step's DPS differs.
const dpsOf = ({ bonusStats }: DpsSimOptions) => 1000 + bonusStats.getStat(Stat.StatStrength) + bonusStats.getStat(Stat.StatAgility);

const rootElem = document.createElement('div');
let settings: DpsCurveSettings;
let simulateDps: ReturnType<typeof vi.fn<(options: DpsSimOptions) => Promise<object>>>;
let abortType: ReturnType<typeof vi.fn>;
let host: any;

const setup = () => {
	const store = createSimStore();
	const storage = new Map<string, string>();
	const env = { storage: { getItem: (key: string) => storage.get(key) ?? null, setItem: (key: string, value: string) => storage.set(key, value) } };
	seedKeyed(store, 'players', 0, { v: zeroVersions(PLAYER_FIELDS) } as never);
	abortType = vi.fn().mockResolvedValue(undefined);
	simulateDps = vi.fn(async (options: DpsSimOptions) => sample(options.iterations, dpsOf(options)));
	const sim = {
		store,
		env,
		runs: new SimRuns(store, { abortType: abortType as unknown as (mask: number) => Promise<void> }),
		getIterations: () => 1000,
		nextRngSeed: () => 7,
		// A player with no stats of its own: the final stats are the bonus.
		getFinalStatsWithBonus: async (bonusStats: Stats) => bonusStats,
		simulateDps,
	};
	const player = { storeKey: 0, sim, getClass: () => Class.ClassWarrior, getSpec: () => Spec.SpecUnknown };
	settings = new DpsCurveSettings(player as never, 'dps-curve-test');
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
				<DpsCurveDialog open onOpenChange={onOpenChange} settings={settings} />
			</SimHostProvider>
		</PortalContainerContext>,
	);
	return onOpenChange;
};

const runButton = () => screen.getByRole<HTMLButtonElement>('button', { name: 'sidebar.buttons.dps_curve.run' });
const tracker = () => rootElem.querySelector('[data-testid="dps-curve-progress-dialog"]');
const bonusOf = (call: [DpsSimOptions], stat: Stat) => call[0].bonusStats.getStat(stat);

beforeEach(() => {
	setup();
	history.saveRun.mockClear();
	download.mockClear();
});

describe('DpsCurveDialog', () => {
	it('runs one plain DPS sim per step at one seed with the step as bonus stats, and saves the run', async () => {
		settings.setSetup({ range: 20, step: 10 });
		settings.setPrecision({ iterations: 3000 });
		renderDialog();

		await act(async () => {
			fireEvent.click(runButton());
		});

		const calls = simulateDps.mock.calls;
		expect(calls.map(call => bonusOf(call, Stat.StatStrength))).toEqual([0, 10, 20]);
		expect(calls.map(call => [call[0].iterations, call[0].seed])).toEqual([
			[3000, 7],
			[3000, 7],
			[3000, 7],
		]);
		expect(history.saveRun).toHaveBeenCalledTimes(1);
		const saved = history.saveRun.mock.calls[0][0];
		expect(saved.points.map(point => [point.statValue, point.dps, point.samples])).toEqual([
			[0, 1000, 3000],
			[10, 1010, 3000],
			[20, 1020, 3000],
		]);
		expect(saved.cappedAt).toBeUndefined();
	});

	it('stops a stat going up once a step gives the DPS of the step below it, and saves where it capped', async () => {
		simulateDps.mockImplementation(async options => sample(options.iterations, Math.min(dpsOf(options), 1020)));
		settings.setSetup({ range: 50, step: 10 });
		renderDialog();

		await act(async () => {
			fireEvent.click(runButton());
		});

		expect(simulateDps.mock.calls.map(call => bonusOf(call, Stat.StatStrength))).toEqual([0, 10, 20, 30]);
		const saved = history.saveRun.mock.calls[0][0];
		expect(saved.points.map(point => point.statValue)).toEqual([0, 10, 20, 30]);
		expect(saved.cappedAt).toBe(20);
		expect(screen.getByText('sidebar.buttons.dps_curve.capped_label')).toBeTruthy();
	});

	it('runs each picked stat on its own in the order offered, saving each as its own run', async () => {
		settings.setSetup({ stats: [AGILITY, STRENGTH], range: 10, step: 10 });
		renderDialog();

		await act(async () => {
			fireEvent.click(runButton());
		});

		const calls = simulateDps.mock.calls;
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
			[...rootElem.querySelectorAll('[id^="dps-curve-overlay-"]')].map(
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
		settings.setPrecision({ precisionMode: 'target', targetDpsCi: 2 });
		renderDialog();

		await act(async () => {
			fireEvent.click(runButton());
		});

		const calls = simulateDps.mock.calls.map(call => [call[0].iterations, call[0].seed]);
		expect(calls[0]).toEqual([DPS_CURVE_PROBE_ITERATIONS, 7]);
		expect(calls[1][1]).toBe(7 + DPS_CURVE_PROBE_ITERATIONS);
		expect(history.saveRun.mock.calls[0][0].points[0].samples).toBe(DPS_CURVE_PROBE_ITERATIONS + calls[1][0]);
	});

	it('shows the run in the progress dialog and cancels it there, keeping the finished steps', async () => {
		let settleSecond!: (result: object) => void;
		simulateDps.mockResolvedValueOnce(sample(1000)).mockImplementationOnce(options => {
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
		expect(tracker()!.querySelector('[data-testid="progress-tracker-modal-progress-title"]')!.textContent).toBe('sidebar.buttons.dps_curve.progress');
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
		expect(simulateDps).toHaveBeenCalledTimes(2);
		expect(history.saveRun.mock.calls[0][0].points).toHaveLength(1);
	});

	it("lists the run's stats in the progress dialog, done up to the running one, and titles it with that stat's step", async () => {
		let settleThird!: (result: object) => void;
		simulateDps
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
		const states = () => [...tracker()!.querySelectorAll('[data-testid="dps-curve-stat"]')].map(row => row.getAttribute('data-state'));
		expect(states()).toEqual(['done', 'running']);
		expect(tracker()!.querySelector('[data-testid="progress-tracker-modal-progress-title"]')!.textContent).toBe('sidebar.buttons.dps_curve.progress');

		await act(async () => {
			fireEvent.click(tracker()!.querySelector('[data-testid="progress-tracker-modal-cancel-btn"]')!);
		});
		expect(states()).toEqual(['done', 'running']);
		await act(async () => {
			settleThird({ type: ErrorOutcomeType.ErrorOutcomeAborted, message: '' });
		});
		expect(tracker()).toBeNull();
	});

	it('exports the drawn runs as CSV', async () => {
		settings.setSetup({ range: 10, step: 10 });
		renderDialog();
		await act(async () => {
			fireEvent.click(runButton());
		});

		fireEvent.click(screen.getByRole('button', { name: 'sidebar.buttons.dps_curve.export_csv' }));

		expect(download).toHaveBeenCalledTimes(1);
		const [csv, fileName, mimeType] = download.mock.calls[0];
		expect(csv.trim().split('\n')).toHaveLength(3);
		expect([fileName, mimeType]).toEqual(['wowsims-dps-curve.csv', 'text/csv']);
	});

	it('disables Run with every stat removed', () => {
		settings.setSetup({ stats: [] });
		renderDialog();

		expect(runButton().disabled).toBe(true);
	});

	it('cancels a run when the dialog closes', async () => {
		simulateDps.mockImplementation(() => new Promise(() => {}));
		const onOpenChange = renderDialog();
		await act(async () => {
			fireEvent.click(runButton());
		});
		abortType.mockClear();

		await act(async () => {
			fireEvent.click(rootElem.querySelector('[data-testid="dps-curve-dialog"] [data-testid="sim-dialog-close"]')!);
		});

		expect(onOpenChange).toHaveBeenCalledWith(false);
		expect(abortType).toHaveBeenCalled();
	});
});
