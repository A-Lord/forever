import i18n from '@i18n/config';
import { fakeCharts } from '@testing/chart-js-mock';
import { makeHitSweepPoints } from '@testing/storybook/fixtures/stat-sweep';
import { render, screen } from '@testing-library/react';
import type { ChartOptions, LegendItem, TooltipItem } from 'chart.js';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { SweepPoint } from '../../model/stat_sweep';
import { StatSweepChart, type StatSweepSeries } from './StatSweepChart';

vi.mock('chart.js', async () => (await import('@testing/chart-js-mock')).chartJsMock);

type Datum = { x: number; y: number; point: SweepPoint };
const charts = fakeCharts as Array<{ config: { data: { datasets: Array<{ label: string; data: Datum[] }> }; options: ChartOptions<'line'> } }>;

const run = (id: string, base: number, colorSlot: number): StatSweepSeries => ({ id, label: id, colorSlot, points: makeHitSweepPoints(base) });

beforeEach(() => {
	charts.length = 0;
});

describe('StatSweepChart', () => {
	it('shows the empty text and builds no chart before a run has points', () => {
		render(<StatSweepChart series={[{ ...run('a', 120, 0), points: [] }]} statLabel="Hit" currentValue={120} />);

		expect(screen.getByText(i18n.t('sidebar.buttons.stat_weights.sweep.empty'))).toBeTruthy();
		expect(charts).toHaveLength(0);
	});

	it('draws EP and DPS as two charts over one x range reaching past every run and the current value', () => {
		render(<StatSweepChart series={[run('a', 120, 0), run('b', 40, 1)]} statLabel="Hit" currentValue={300} />);

		expect(charts).toHaveLength(2);
		const [ep, dps] = charts.map(chart => chart.config.options.scales!.x!);
		expect([ep.min, ep.max]).toEqual([dps.min, dps.max]);
		expect(ep.min).toBeLessThan(0);
		expect(ep.max).toBeGreaterThan(300);
	});

	it('lays out the same x tick labels on both charts and shows them only under the DPS chart', () => {
		render(<StatSweepChart series={[run('a', 120, 0)]} statLabel="Hit" currentValue={120} />);

		const [ep, dps] = charts.map(chart => chart.config.options.scales!.x!.ticks!);
		expect(ep.color).toBe('transparent');
		expect({ ...ep, color: undefined }).toEqual(dps);
	});

	it('wraps each run line in a band of value ± ci', () => {
		const single = run('a', 120, 0);
		render(<StatSweepChart series={[single]} statLabel="Hit" currentValue={120} />);

		const [ep, dps] = charts.map(chart => chart.config.data.datasets);
		const point = single.points[3];
		expect(ep.map(dataset => dataset.data[3].y)).toEqual([point.ep + point.epCi, point.ep - point.epCi, point.ep]);
		expect(dps.map(dataset => dataset.data[3].y)).toEqual([point.dps + point.dpsCi, point.dps - point.dpsCi, point.dps]);
	});

	it('lists each run once in the legend and the tooltip, and shows the legend only for two or more runs', () => {
		render(<StatSweepChart series={[run('a', 120, 0)]} statLabel="Hit" currentValue={120} />);
		expect(charts[0].config.options.plugins!.legend!.display).toBe(false);

		charts.length = 0;
		render(<StatSweepChart series={[run('a', 120, 0), run('b', 90, 1)]} statLabel="Hit" currentValue={120} />);
		const [ep, dps] = charts.map(chart => chart.config.options.plugins!);
		expect([ep.legend!.display, dps.legend!.display]).toEqual([true, false]);

		const shown = charts[0].config.data.datasets.map((_, datasetIndex) => datasetIndex);
		const legendFilter = ep.legend!.labels!.filter!;
		const tooltipFilter = ep.tooltip!.filter!;
		expect(shown.filter(datasetIndex => legendFilter({ datasetIndex } as LegendItem, {} as never))).toEqual([2, 5]);
		expect(shown.filter(datasetIndex => tooltipFilter({ datasetIndex } as TooltipItem<'line'>, 0, [], {} as never))).toEqual([2, 5]);
	});

	it('reads the hovered value with its ±, EP to three decimals and DPS to one', () => {
		const single = run('a', 120, 0);
		render(<StatSweepChart series={[single]} statLabel="Hit" currentValue={120} />);
		const t = vi.spyOn(i18n, 't');
		const point = single.points[0];

		const labels = charts.map(chart => {
			const dataset = chart.config.data.datasets[2];
			const label = chart.config.options.plugins!.tooltip!.callbacks!.label! as unknown as (item: object) => string;
			return label({ dataset, raw: dataset.data[0] });
		});

		expect(labels.every(label => label.startsWith('a: '))).toBe(true);
		expect(t).toHaveBeenCalledWith('sidebar.buttons.stat_weights.sweep.tooltip_value', { value: point.ep.toFixed(3), ci: point.epCi.toFixed(3) });
		expect(t).toHaveBeenCalledWith('sidebar.buttons.stat_weights.sweep.tooltip_value', { value: point.dps.toFixed(1), ci: point.dpsCi.toFixed(1) });
		t.mockRestore();
	});
});
