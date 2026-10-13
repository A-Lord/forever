import i18n from '@i18n/config';
import type { StatSweepPoint } from '@sim/cache/stat_sweep_history';
import { fakeCharts } from '@testing/chart-js-mock';
import { makeCritSweepPoints, makeHitSweepPoints, sweepSeries } from '@testing/storybook/fixtures/stat-sweep';
import { render, screen } from '@testing-library/react';
import type { ChartOptions, Plugin, TooltipItem } from 'chart.js';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

import { seriesColor } from './colors';
import { StatSweepChart, type StatSweepSeries } from './StatSweepChart';

vi.mock('chart.js', async () => (await import('@testing/chart-js-mock')).chartJsMock);

type Datum = { x: number; y: number; point: StatSweepPoint };
type Dataset = { label: string; yAxisID: string; data: Datum[] };
const charts = fakeCharts as Array<{ config: { data: { datasets: Dataset[] }; options: ChartOptions<'line'>; plugins: Array<Plugin<'line'>> } }>;

const run = (id: string, base: number, colorSlot: number): StatSweepSeries => sweepSeries(id, 'Hit', base, makeHitSweepPoints(base), colorSlot, id);
const sweepT = (key: string) => i18n.t(`sidebar.buttons.stat_weights.sweep.${key}`);

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
		render(<StatSweepChart series={[{ ...run('a', 120, 0), points: [] }]} />);

		expect(screen.getByText(sweepT('empty'))).toBeTruthy();
		expect(charts).toHaveLength(0);
	});

	it('plots EP on the left axis and DPS on the right one over one plot, over the change from the current value', () => {
		render(<StatSweepChart series={[run('a', 120, 0), run('b', 40, 1)]} />);

		expect(charts).toHaveLength(1);
		const { x, ep, dps } = charts[0].config.options.scales! as Record<string, any>;
		expect([ep.position, ep.title.text]).toEqual(['left', sweepT('ep_chart')]);
		expect([dps.position, dps.title.text]).toEqual(['right', sweepT('dps_chart')]);
		expect(ep.stack).toBeUndefined();
		expect(ep.grid.drawOnChartArea).not.toBe(false);
		expect(dps.grid.drawOnChartArea).toBe(false);
		expect(x.title.text).toBe(sweepT('change_axis'));
		expect(x.min).toBeLessThan(-100);
		expect(x.max).toBeGreaterThan(100);
	});

	it('plots every run at its change from its own current value, so runs of other stats and bases line up around 0', () => {
		const hit = run('hit', 40, 0);
		const crit = sweepSeries('crit', 'Crit', 210, makeCritSweepPoints(210), 1);
		render(<StatSweepChart series={[hit, crit]} />);

		const lines = charts[0].config.data.datasets.filter((_, index) => index % 4 === 2);
		const [hitEp, critEp] = lines;
		expect(hitEp.data.map(datum => datum.x)).toEqual(hit.points.map(point => point.statValue - 40));
		expect(critEp.data.map(datum => datum.x)).toEqual(crit.points.map(point => point.statValue - 210));
		expect(hitEp.data[0].x).toBe(-40);
		expect(critEp.data[0].x).toBe(-100);
		expect(hitEp.data.some(datum => datum.x === 0) && critEp.data.some(datum => datum.x === 0)).toBe(true);
	});

	it('marks the current value at 0', () => {
		render(<StatSweepChart series={[run('a', 120, 0)]} />);
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

	it('draws each run as a solid EP line in a band of EP ± ci and a dashed DPS line with hollow points', () => {
		const single = run('a', 120, 0);
		render(<StatSweepChart series={[single]} />);

		const datasets = charts[0].config.data.datasets;
		const point = single.points[3];
		expect(datasets.map(dataset => [dataset.yAxisID, dataset.data[3].y])).toEqual([
			['ep', point.ep + point.epCi],
			['ep', point.ep - point.epCi],
			['ep', point.ep],
			['dps', point.dps],
		]);
		const [, , ep, dps] = datasets as Array<Dataset & Record<string, unknown>>;
		expect(ep.borderDash).toBeUndefined();
		expect(dps.borderDash).toEqual(expect.any(Array));
		expect([ep.borderColor, dps.borderColor, dps.pointBorderColor]).toEqual([seriesColor(0), seriesColor(0), seriesColor(0)]);
		expect(dps.backgroundColor).toBe(ep.pointBorderColor);
		expect(dps.backgroundColor).not.toBe(seriesColor(0));
		expect(dps.pointRadius).toBeLessThan(ep.pointRadius as number);
	});

	it('keys solid as EP, dashed as DPS and the current value under the chart, even for one run', () => {
		render(<StatSweepChart series={[run('a', 120, 0)]} />);

		const caption = screen.getByText(sweepT('current_value')).closest('figcaption')!;
		expect(caption.textContent).toContain(sweepT('ep_chart'));
		expect(caption.textContent).toContain(sweepT('dps_chart'));
	});

	it('lists each run once in the legend, both its lines in the tooltip, and shows the legend only for two or more runs', () => {
		render(<StatSweepChart series={[run('a', 120, 0)]} />);
		expect(charts[0].config.options.plugins!.legend!.display).toBe(false);

		charts.length = 0;
		render(<StatSweepChart series={[run('a', 120, 0), run('b', 90, 1)]} />);
		const { data, options } = charts[0].config;
		expect(options.plugins!.legend!.display).toBe(true);

		const shown = data.datasets.map((_, datasetIndex) => datasetIndex);
		const generateLabels = options.plugins!.legend!.labels!.generateLabels!;
		const tooltipFilter = options.plugins!.tooltip!.filter!;
		const items = generateLabels({ data, isDatasetVisible: () => true } as never);
		expect(items.map(item => item.datasetIndex)).toEqual([2, 6]);
		expect(items.map(item => [item.text, item.strokeStyle, item.fillStyle])).toEqual([
			['a', seriesColor(0), seriesColor(0)],
			['b', seriesColor(1), seriesColor(1)],
		]);
		expect(shown.filter(datasetIndex => tooltipFilter({ datasetIndex } as TooltipItem<'line'>, 0, [], {} as never))).toEqual([2, 3, 6, 7]);
	});

	it('titles the tooltip with each hovered stat total and its change, once per total', () => {
		render(<StatSweepChart series={[run('a', 120, 0), run('b', 90, 1)]} />);
		const { data, options } = charts[0].config;
		const title = options.plugins!.tooltip!.callbacks!.title! as unknown as (items: object[]) => string[];
		const at = (datasetIndex: number, change: number) => {
			const raw = data.datasets[datasetIndex].data.find(datum => datum.x === change)!;
			return { raw, dataset: data.datasets[datasetIndex] };
		};
		const t = vi
			.spyOn(i18n, 't')
			.mockImplementation(
				((_key: string, options: { stat: string; total: number; change: string }) => `${options.stat} ${options.total} (${options.change})`) as never,
			);

		expect(title([at(2, 20), at(3, 20)])).toEqual(['Hit 140 (+20)']);
		expect(title([at(2, -20), at(3, -20), at(6, -20), at(7, -20)])).toEqual(['Hit 100 (-20)', 'Hit 70 (-20)']);
		expect(t).toHaveBeenCalledWith('sidebar.buttons.stat_weights.sweep.tooltip_total', { stat: 'Hit', total: 140, change: '+20' });
		t.mockRestore();
	});

	it('reads the hovered value with its ± and measure, EP to three decimals and DPS to one', () => {
		const single = run('a', 120, 0);
		render(<StatSweepChart series={[single]} />);
		const { data, options } = charts[0].config;
		const label = options.plugins!.tooltip!.callbacks!.label! as unknown as (item: object) => string;
		const t = vi.spyOn(i18n, 't');
		const point = single.points[0];

		const [ep, dps] = [data.datasets[2], data.datasets[3]].map(dataset => label({ dataset, raw: dataset.data[0] }));

		expect(ep.startsWith(`a (${sweepT('ep_chart')}): `)).toBe(true);
		expect(dps.startsWith(`a (${sweepT('dps_chart')}): `)).toBe(true);
		expect(t).toHaveBeenCalledWith('sidebar.buttons.stat_weights.sweep.tooltip_value', { value: point.ep.toFixed(3), ci: point.epCi.toFixed(3) });
		expect(t).toHaveBeenCalledWith('sidebar.buttons.stat_weights.sweep.tooltip_value', { value: point.dps.toFixed(1), ci: point.dpsCi.toFixed(1) });
		t.mockRestore();
	});
});
