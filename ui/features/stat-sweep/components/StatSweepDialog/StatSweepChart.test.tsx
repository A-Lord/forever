import i18n from '@i18n/config';
import type { StatSweepPoint } from '@sim/cache/stat_sweep_history';
import type { StatSweepMetric } from '@sim/state/sim_store';
import { fakeCharts } from '@testing/chart-js-mock';
import { makeCritSweepPoints, makeDefenseSweepPoints, makeHitSweepPoints, sweepSeries } from '@testing/storybook/fixtures/stat-sweep';
import { fireEvent, render, screen, within } from '@testing-library/react';
import type { ChartOptions, Plugin, TooltipItem } from 'chart.js';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

import { seriesColor } from './colors';
import { StatSweepChart, type StatSweepChartProps, type StatSweepSeries } from './StatSweepChart';

vi.mock('chart.js', async () => (await import('@testing/chart-js-mock')).chartJsMock);

type Datum = { x: number; y: number; point: StatSweepPoint };
type Dataset = { label: string; data: Datum[] } & Record<string, unknown>;
const charts = fakeCharts as Array<{ config: { data: { datasets: Dataset[] }; options: ChartOptions<'line'>; plugins: Array<Plugin<'line'>> } }>;

const run = (id: string, base: number, colorSlot: number): StatSweepSeries => sweepSeries(id, 'Hit', base, makeHitSweepPoints(base), colorSlot, id);
const sweepT = (key: string, options?: Record<string, unknown>) => i18n.t(`sidebar.buttons.stat_sweep.${key}`, options);
const metricLabel = (metric: StatSweepMetric) => i18n.t(`sidebar.results.metrics.${metric}.label`);

const renderChart = (series: StatSweepSeries[], props: Partial<StatSweepChartProps> = {}) =>
	render(
		<StatSweepChart series={series} metric="dps" metrics={['dps', 'tps']} onMetricChange={vi.fn()} showSpread onShowSpreadChange={vi.fn()} {...props} />,
	);

// jsdom loads no theme, and the chart caches each colour on first read, so the colours are set before any render.
beforeAll(() => {
	const root = document.documentElement.style;
	root.setProperty('--color-chart-series-1', '#111111');
	root.setProperty('--color-chart-series-2', '#222222');
	root.setProperty('--color-background', '#000000');
});

beforeEach(() => {
	charts.length = 0;
});

describe('StatSweepChart', () => {
	it('shows the empty text and builds no chart before a run has points', () => {
		renderChart([{ ...run('a', 120, 0), points: [] }]);

		expect(screen.getByText(sweepT('empty', { metric: metricLabel('dps') }))).toBeTruthy();
		expect(charts).toHaveLength(0);
	});

	it('plots the metric on one axis over the stat added', () => {
		renderChart([run('a', 120, 0), run('b', 40, 1)]);

		expect(charts).toHaveLength(1);
		const scales = charts[0].config.options.scales! as Record<string, any>;
		expect(Object.keys(scales)).toEqual(['x', 'y']);
		expect(scales.y.title.text).toBe(metricLabel('dps'));
		expect(scales.x.title.text).toBe(sweepT('added_axis'));
		expect(scales.x.min).toBeLessThan(-100);
		expect(scales.x.max).toBeGreaterThan(100);
	});

	it('plots every run at the stat added, so runs of other stats and bases line up around 0', () => {
		const hit = run('hit', 40, 0);
		const crit = sweepSeries('crit', 'Crit', 210, makeCritSweepPoints(210), 1);
		renderChart([hit, crit]);

		const [hitLine, critLine] = charts[0].config.data.datasets.filter((_, index) => index % 3 === 2);
		expect(hitLine.data.map(datum => datum.x)).toEqual(hit.points.map(point => point.offset));
		expect(critLine.data.map(datum => datum.x)).toEqual(crit.points.map(point => point.offset));
		expect(hitLine.data.some(datum => datum.x === 0) && critLine.data.some(datum => datum.x === 0)).toBe(true);
	});

	it('plots a stat with a multiplier at the points added, not at the change of its total', () => {
		const base = run('a', 100, 0);
		const multiplied = { ...base, points: base.points.map(point => ({ ...point, statValue: 100 + point.offset * 1.1 })) };
		renderChart([multiplied]);

		const line = charts[0].config.data.datasets[2];
		expect(line.data.map(datum => datum.x)).toEqual(multiplied.points.map(point => point.offset));
		expect(line.data.find(datum => datum.x === 10)!.point.statValue).toBeCloseTo(111);
	});

	it('marks the current value at 0', () => {
		renderChart([run('a', 120, 0)]);
		const marker = charts[0].config.plugins.find(plugin => plugin.id === 'statSweepMarker')!;
		const getPixelForValue = vi.fn(() => 40);
		const moveTo = vi.fn();
		const ctx = { save: vi.fn(), restore: vi.fn(), setLineDash: vi.fn(), beginPath: vi.fn(), moveTo, lineTo: vi.fn(), stroke: vi.fn() };

		(marker.afterDatasetsDraw as (chart: unknown) => void)({
			ctx,
			chartArea: { left: 0, right: 100, top: 0, bottom: 50 },
			scales: { x: { getPixelForValue } },
		});

		expect(getPixelForValue).toHaveBeenCalledWith(0);
		expect(moveTo).toHaveBeenCalledWith(40, 0);
	});

	it("draws each run as a line in a band of the metric ± one standard deviation in the run's colour", () => {
		const single = run('a', 120, 0);
		renderChart([single]);

		const datasets = charts[0].config.data.datasets;
		const { avg, stdev } = single.points[3].metrics.dps;
		expect(datasets.map(dataset => dataset.data[3].y)).toEqual([avg + stdev, avg - stdev, avg]);
		expect(datasets[2].borderColor).toBe(seriesColor(0));
	});

	it('draws only the lines, lists each run once in the legend and tooltip, and drops the band key without the spread', () => {
		const single = run('a', 120, 0);
		renderChart([single, run('b', 90, 1)], { showSpread: false });

		const { data, options } = charts[0].config;
		expect(data.datasets.map(dataset => dataset.data[3].y)).toEqual([single.points[3].metrics.dps.avg, expect.any(Number)]);
		expect(data.datasets.map(dataset => dataset.borderColor)).toEqual([seriesColor(0), seriesColor(1)]);
		const items = options.plugins!.legend!.labels!.generateLabels!({ data, isDatasetVisible: () => true } as never);
		expect(items.map(item => [item.datasetIndex, item.text])).toEqual([
			[0, 'a'],
			[1, 'b'],
		]);
		const tooltipFilter = options.plugins!.tooltip!.filter!;
		expect([0, 1].filter(datasetIndex => tooltipFilter({ datasetIndex } as TooltipItem<'line'>, 0, [], {} as never))).toEqual([0, 1]);
		expect(screen.queryByText(sweepT('band_key', { metric: metricLabel('dps') }))).toBeNull();
	});

	it('reports the spread checkbox and shows it as given', () => {
		const onShowSpreadChange = vi.fn();
		renderChart([run('a', 120, 0)], { onShowSpreadChange });

		const checkbox = within(screen.getByTestId('stat-sweep-show-spread')).getByRole('checkbox');
		expect(checkbox.getAttribute('aria-checked')).toBe('true');
		fireEvent.click(checkbox);
		expect(onShowSpreadChange).toHaveBeenCalledWith(false);
	});

	it('offers the given metrics in a toggle and reports a pick', () => {
		const onMetricChange = vi.fn();
		renderChart([run('a', 120, 0)], { metrics: ['dps', 'tps', 'dtps', 'tmi'], onMetricChange });

		const toggle = screen.getByTestId('stat-sweep-metric');
		expect(
			within(toggle)
				.getAllByRole('button')
				.map(button => [button.textContent, button.getAttribute('aria-pressed')]),
		).toEqual([
			[metricLabel('dps'), 'true'],
			[metricLabel('tps'), 'false'],
			[metricLabel('dtps'), 'false'],
			[metricLabel('tmi'), 'false'],
		]);

		fireEvent.click(within(toggle).getByRole('button', { name: metricLabel('dtps') }));
		expect(onMetricChange).toHaveBeenCalledWith('dtps');
	});

	it("redraws a picked metric's line, band, axis title and band key from the stored points", () => {
		const tank = sweepSeries('def', 'Defense', 300, makeDefenseSweepPoints(300));
		const { rerender } = renderChart([tank], { metrics: ['dps', 'tps', 'dtps', 'tmi'] });

		rerender(
			<StatSweepChart
				series={[tank]}
				metric="dtps"
				metrics={['dps', 'tps', 'dtps', 'tmi']}
				onMetricChange={vi.fn()}
				showSpread
				onShowSpreadChange={vi.fn()}
			/>,
		);

		const { data, options } = charts[charts.length - 1].config;
		const { avg, stdev } = tank.points[3].metrics.dtps;
		expect(data.datasets.map(dataset => dataset.data[3].y)).toEqual([avg + stdev, avg - stdev, avg]);
		expect((options.scales as Record<string, any>).y.title.text).toBe(metricLabel('dtps'));
		expect(screen.getByText(sweepT('band_key', { metric: metricLabel('dtps') }))).toBeTruthy();
	});

	it('lists each run once in the legend and in the tooltip, and shows the legend only for two or more runs', () => {
		renderChart([run('a', 120, 0)]);
		expect(charts[0].config.options.plugins!.legend!.display).toBe(false);

		charts.length = 0;
		renderChart([run('a', 120, 0), run('b', 90, 1)]);
		const { data, options } = charts[0].config;
		expect(options.plugins!.legend!.display).toBe(true);

		const generateLabels = options.plugins!.legend!.labels!.generateLabels!;
		const tooltipFilter = options.plugins!.tooltip!.filter!;
		const items = generateLabels({ data, isDatasetVisible: () => true } as never);
		expect(items.map(item => [item.datasetIndex, item.text, item.strokeStyle])).toEqual([
			[2, 'a', seriesColor(0)],
			[5, 'b', seriesColor(1)],
		]);
		const shown = data.datasets.map((_, datasetIndex) => datasetIndex);
		expect(shown.filter(datasetIndex => tooltipFilter({ datasetIndex } as TooltipItem<'line'>, 0, [], {} as never))).toEqual([2, 5]);
	});

	it('titles the tooltip with each hovered stat total and the stat added, once per total', () => {
		renderChart([run('a', 120, 0), run('b', 90, 1)]);
		const { data, options } = charts[0].config;
		const title = options.plugins!.tooltip!.callbacks!.title! as unknown as (items: object[]) => string[];
		const at = (datasetIndex: number, change: number) => ({ raw: data.datasets[datasetIndex].data.find(datum => datum.x === change)! });
		const t = vi
			.spyOn(i18n, 't')
			.mockImplementation(
				((_key: string, options: { stat: string; total: number; added: string }) => `${options.stat} ${options.total} (${options.added})`) as never,
			);

		expect(title([at(2, 20)])).toEqual(['Hit 140 (+20)']);
		expect(title([at(2, -20), at(5, -20)])).toEqual(['Hit 100 (-20)', 'Hit 70 (-20)']);
		t.mockRestore();
	});

	it('reads the hovered value and its ± rounded as the sim card does for the plotted metric', () => {
		const tank = sweepSeries('a', 'Defense', 300, makeDefenseSweepPoints(300), 0, 'a');
		const tooltipText = (metric: StatSweepMetric) => {
			charts.length = 0;
			renderChart([tank], { metric, metrics: ['dps', 'tps', 'dtps', 'tmi'] });
			const { data, options } = charts[0].config;
			const label = options.plugins!.tooltip!.callbacks!.label! as unknown as (item: object) => string;
			const t = vi.spyOn(i18n, 't');
			const text = label({ dataset: data.datasets[2], raw: data.datasets[2].data[0] });
			const call = t.mock.calls.find(([key]) => key === 'sidebar.buttons.stat_sweep.tooltip_value');
			t.mockRestore();
			return { text, values: call?.[1] };
		};
		const { dps, tmi } = tank.points[0].metrics;

		const byDps = tooltipText('dps');
		expect(byDps.text.startsWith('a: ')).toBe(true);
		expect(byDps.values).toEqual({ value: dps.avg.toFixed(2), stdev: dps.stdev.toFixed(0) });
		expect(tooltipText('tmi').values).toEqual({ value: tmi.avg.toFixed(2), stdev: tmi.stdev.toFixed(2) });
	});

	it('names where a capped run capped in the legend, and marks its points from there on as capped in the tooltip', () => {
		renderChart([{ ...run('a', 120, 0), cappedAt: 20 }, run('b', 90, 1)]);
		const { data, options } = charts[0].config;
		const t = vi.spyOn(i18n, 't');

		options.plugins!.legend!.labels!.generateLabels!({ data, isDatasetVisible: () => true } as never);
		expect(t).toHaveBeenCalledWith('sidebar.buttons.stat_sweep.capped_label', { label: 'a', added: '+20' });

		const label = options.plugins!.tooltip!.callbacks!.label! as unknown as (item: object) => string;
		const keyAt = (datasetIndex: number, change: number) => {
			t.mockClear();
			label({ dataset: data.datasets[datasetIndex], raw: data.datasets[datasetIndex].data.find(datum => datum.x === change) });
			return t.mock.calls.find(([key]) => String(key).startsWith('sidebar.buttons.stat_sweep.'))![0];
		};
		expect([keyAt(2, 10), keyAt(2, 20), keyAt(2, 30), keyAt(5, 30)]).toEqual([
			'sidebar.buttons.stat_sweep.tooltip_value',
			'sidebar.buttons.stat_sweep.tooltip_value_capped',
			'sidebar.buttons.stat_sweep.tooltip_value_capped',
			'sidebar.buttons.stat_sweep.tooltip_value',
		]);
		t.mockRestore();
	});
});
