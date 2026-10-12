import i18n from '@i18n/config';
import type { StatSweepPoint } from '@sim/cache/stat_sweep_history';
import { fakeCharts } from '@testing/chart-js-mock';
import { makeHitSweepPoints } from '@testing/storybook/fixtures/stat-sweep';
import { render, screen } from '@testing-library/react';
import type { ChartOptions, TooltipItem } from 'chart.js';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { seriesColor } from './colors';
import { StatSweepChart, type StatSweepSeries } from './StatSweepChart';

vi.mock('chart.js', async () => (await import('@testing/chart-js-mock')).chartJsMock);

type Datum = { x: number; y: number; point: StatSweepPoint };
type Dataset = { label: string; yAxisID: string; data: Datum[] };
const charts = fakeCharts as Array<{ config: { data: { datasets: Dataset[] }; options: ChartOptions<'line'> } }>;

const run = (id: string, base: number, colorSlot: number): StatSweepSeries => ({ id, label: id, colorSlot, points: makeHitSweepPoints(base) });
const sweepT = (key: string) => i18n.t(`sidebar.buttons.stat_weights.sweep.${key}`);

beforeEach(() => {
	charts.length = 0;
});

describe('StatSweepChart', () => {
	it('shows the empty text and builds no chart before a run has points', () => {
		render(<StatSweepChart series={[{ ...run('a', 120, 0), points: [] }]} statLabel="Hit" currentValue={120} />);

		expect(screen.getByText(sweepT('empty'))).toBeTruthy();
		expect(charts).toHaveLength(0);
	});

	it('stacks the EP panel over the DPS panel in one chart, over an x range reaching past every run and the current value', () => {
		render(<StatSweepChart series={[run('a', 120, 0), run('b', 40, 1)]} statLabel="Hit" currentValue={300} />);

		expect(charts).toHaveLength(1);
		const { x, ep, dps } = charts[0].config.options.scales! as Record<string, any>;
		expect(ep.stack).toBe(dps.stack);
		expect(ep.weight).toBeGreaterThan(dps.weight);
		expect(x.min).toBeLessThan(0);
		expect(x.max).toBeGreaterThan(300);
	});

	it('wraps each run line in a band of value ± ci on its own panel', () => {
		const single = run('a', 120, 0);
		render(<StatSweepChart series={[single]} statLabel="Hit" currentValue={120} />);

		const datasets = charts[0].config.data.datasets;
		const point = single.points[3];
		expect(datasets.map(dataset => [dataset.yAxisID, dataset.data[3].y])).toEqual([
			['ep', point.ep + point.epCi],
			['ep', point.ep - point.epCi],
			['ep', point.ep],
			['dps', point.dps + point.dpsCi],
			['dps', point.dps - point.dpsCi],
			['dps', point.dps],
		]);
	});

	it('lists each run once in the legend, both its lines in the tooltip, and shows the legend only for two or more runs', () => {
		render(<StatSweepChart series={[run('a', 120, 0)]} statLabel="Hit" currentValue={120} />);
		expect(charts[0].config.options.plugins!.legend!.display).toBe(false);

		charts.length = 0;
		render(<StatSweepChart series={[run('a', 120, 0), run('b', 90, 1)]} statLabel="Hit" currentValue={120} />);
		const { data, options } = charts[0].config;
		expect(options.plugins!.legend!.display).toBe(true);

		const shown = data.datasets.map((_, datasetIndex) => datasetIndex);
		const generateLabels = options.plugins!.legend!.labels!.generateLabels!;
		const tooltipFilter = options.plugins!.tooltip!.filter!;
		const items = generateLabels({ data, isDatasetVisible: () => true } as never);
		expect(items.map(item => item.datasetIndex)).toEqual([2, 5]);
		expect(items.map(item => [item.text, item.strokeStyle, item.fillStyle])).toEqual([
			['a', seriesColor(0), seriesColor(0)],
			['b', seriesColor(1), seriesColor(1)],
		]);
		expect(shown.filter(datasetIndex => tooltipFilter({ datasetIndex } as TooltipItem<'line'>, 0, [], {} as never))).toEqual([2, 5, 8, 11]);
	});

	it('reads the hovered value with its ± and panel, EP to three decimals and DPS to one', () => {
		const single = run('a', 120, 0);
		render(<StatSweepChart series={[single]} statLabel="Hit" currentValue={120} />);
		const { data, options } = charts[0].config;
		const label = options.plugins!.tooltip!.callbacks!.label! as unknown as (item: object) => string;
		const t = vi.spyOn(i18n, 't');
		const point = single.points[0];

		const [ep, dps] = [data.datasets[2], data.datasets[5]].map(dataset => label({ dataset, raw: dataset.data[0] }));

		expect(ep.startsWith(`a (${sweepT('ep_chart')}): `)).toBe(true);
		expect(dps.startsWith(`a (${sweepT('dps_chart')}): `)).toBe(true);
		expect(t).toHaveBeenCalledWith('sidebar.buttons.stat_weights.sweep.tooltip_value', { value: point.ep.toFixed(3), ci: point.epCi.toFixed(3) });
		expect(t).toHaveBeenCalledWith('sidebar.buttons.stat_weights.sweep.tooltip_value', { value: point.dps.toFixed(1), ci: point.dpsCi.toFixed(1) });
		t.mockRestore();
	});
});
