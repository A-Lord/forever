import i18n from '@i18n/config';
import type { DpsCurvePoint } from '@sim/cache/dps_curve_history';
import { fakeCharts } from '@testing/chart-js-mock';
import { curveSeries, makeCritCurvePoints, makeHitCurvePoints } from '@testing/storybook/fixtures/dps-curve';
import { render, screen } from '@testing-library/react';
import type { ChartOptions, Plugin, TooltipItem } from 'chart.js';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

import { pointCi } from '../../model/dps_curve';
import { seriesColor } from './colors';
import { DpsCurveChart, type DpsCurveSeries } from './DpsCurveChart';

vi.mock('chart.js', async () => (await import('@testing/chart-js-mock')).chartJsMock);

type Datum = { x: number; y: number; point: DpsCurvePoint };
type Dataset = { label: string; data: Datum[] } & Record<string, unknown>;
const charts = fakeCharts as Array<{ config: { data: { datasets: Dataset[] }; options: ChartOptions<'line'>; plugins: Array<Plugin<'line'>> } }>;

const run = (id: string, base: number, colorSlot: number): DpsCurveSeries => curveSeries(id, 'Hit', base, makeHitCurvePoints(base), colorSlot, id);
const curveT = (key: string) => i18n.t(`sidebar.buttons.dps_curve.${key}`);

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

describe('DpsCurveChart', () => {
	it('shows the empty text and builds no chart before a run has points', () => {
		render(<DpsCurveChart series={[{ ...run('a', 120, 0), points: [] }]} />);

		expect(screen.getByText(curveT('empty'))).toBeTruthy();
		expect(charts).toHaveLength(0);
	});

	it('plots DPS on one axis over the change from the current value', () => {
		render(<DpsCurveChart series={[run('a', 120, 0), run('b', 40, 1)]} />);

		expect(charts).toHaveLength(1);
		const scales = charts[0].config.options.scales! as Record<string, any>;
		expect(Object.keys(scales)).toEqual(['x', 'y']);
		expect(scales.y.title.text).toBe(curveT('dps_axis'));
		expect(scales.x.title.text).toBe(curveT('change_axis'));
		expect(scales.x.min).toBeLessThan(-100);
		expect(scales.x.max).toBeGreaterThan(100);
	});

	it('plots every run at its change from its own current value, so runs of other stats and bases line up around 0', () => {
		const hit = run('hit', 40, 0);
		const crit = curveSeries('crit', 'Crit', 210, makeCritCurvePoints(210), 1);
		render(<DpsCurveChart series={[hit, crit]} />);

		const [hitLine, critLine] = charts[0].config.data.datasets.filter((_, index) => index % 3 === 2);
		expect(hitLine.data.map(datum => datum.x)).toEqual(hit.points.map(point => point.statValue - 40));
		expect(critLine.data.map(datum => datum.x)).toEqual(crit.points.map(point => point.statValue - 210));
		expect(hitLine.data.some(datum => datum.x === 0) && critLine.data.some(datum => datum.x === 0)).toBe(true);
	});

	it('marks the current value at 0', () => {
		render(<DpsCurveChart series={[run('a', 120, 0)]} />);
		const marker = charts[0].config.plugins.find(plugin => plugin.id === 'dpsCurveMarker')!;
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

	it("draws each run as a DPS line in a band of DPS ± its 90% ± in the run's colour", () => {
		const single = run('a', 120, 0);
		render(<DpsCurveChart series={[single]} />);

		const datasets = charts[0].config.data.datasets;
		const point = single.points[3];
		expect(datasets.map(dataset => dataset.data[3].y)).toEqual([point.dps + pointCi(point), point.dps - pointCi(point), point.dps]);
		expect(datasets[2].borderColor).toBe(seriesColor(0));
	});

	it('lists each run once in the legend and in the tooltip, and shows the legend only for two or more runs', () => {
		render(<DpsCurveChart series={[run('a', 120, 0)]} />);
		expect(charts[0].config.options.plugins!.legend!.display).toBe(false);

		charts.length = 0;
		render(<DpsCurveChart series={[run('a', 120, 0), run('b', 90, 1)]} />);
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

	it('titles the tooltip with each hovered stat total and its change, once per total', () => {
		render(<DpsCurveChart series={[run('a', 120, 0), run('b', 90, 1)]} />);
		const { data, options } = charts[0].config;
		const title = options.plugins!.tooltip!.callbacks!.title! as unknown as (items: object[]) => string[];
		const at = (datasetIndex: number, change: number) => ({ raw: data.datasets[datasetIndex].data.find(datum => datum.x === change)! });
		const t = vi
			.spyOn(i18n, 't')
			.mockImplementation(
				((_key: string, options: { stat: string; total: number; change: string }) => `${options.stat} ${options.total} (${options.change})`) as never,
			);

		expect(title([at(2, 20)])).toEqual(['Hit 140 (+20)']);
		expect(title([at(2, -20), at(5, -20)])).toEqual(['Hit 100 (-20)', 'Hit 70 (-20)']);
		t.mockRestore();
	});

	it('reads the hovered DPS and its ± rounded as the sim card does', () => {
		const single = run('a', 120, 0);
		render(<DpsCurveChart series={[single]} />);
		const { data, options } = charts[0].config;
		const label = options.plugins!.tooltip!.callbacks!.label! as unknown as (item: object) => string;
		const t = vi.spyOn(i18n, 't');
		const point = single.points[0];

		const text = label({ dataset: data.datasets[2], raw: data.datasets[2].data[0] });

		expect(text.startsWith('a: ')).toBe(true);
		expect(t).toHaveBeenCalledWith('sidebar.buttons.dps_curve.tooltip_value', { value: point.dps.toFixed(2), ci: pointCi(point).toFixed(0) });
		t.mockRestore();
	});
});
