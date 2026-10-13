import { AXIS_GRID_COLOR, cssVarColor } from '@features/results/components/Timeline/chart/colors';
import i18n from '@i18n/config';
import type { StatSweepPoint } from '@sim/cache/stat_sweep_history';
import { distinct } from '@sim/utils/collections';
import type { ChartData, ChartDataset, ChartOptions, Plugin } from 'chart.js';
import { useMemo } from 'react';
import { Line } from 'react-chartjs-2';

import { sweepChange } from '../../model/stat_sweep';
import { seriesColor } from './colors';

export interface StatSweepSeries {
	id: string;
	label: string;
	colorSlot: number;
	statLabel: string;
	/** The swept stat's total when the run started; the run plots each point at its change from it. */
	baseStatValue: number;
	points: ReadonlyArray<StatSweepPoint>;
}

export interface StatSweepChartProps {
	series: ReadonlyArray<StatSweepSeries>;
}

type Metric = 'ep' | 'dps';
type SweepDatum = { x: number; y: number; point: StatSweepPoint; statLabel: string };

// Each run draws the band's upper edge, its lower edge, then the line. Upper fills down to the next dataset, so the order is load-bearing.
const DATASETS_PER_RUN = 3;
const LINE_INDEX = 2;
const isLine = (datasetIndex: number) => datasetIndex % DATASETS_PER_RUN === LINE_INDEX;
// The x range reaches this share of its span past the outermost value, so the end points draw whole inside the plot.
const X_RANGE_PAD = 0.03;
const BAND_ALPHA = '1a';
const LEGEND_LINE_WIDTH = 3;
// Each panel's scale fits its data, with this share of the span as room past the extremes.
const Y_GRACE = '5%';

const METRICS = {
	ep: { value: (point: StatSweepPoint) => point.ep, ci: (point: StatSweepPoint) => point.epCi, digits: 3, title: 'ep_chart' },
	dps: { value: (point: StatSweepPoint) => point.dps, ci: (point: StatSweepPoint) => point.dpsCi, digits: 1, title: 'dps_chart' },
} as const;

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`sidebar.buttons.stat_weights.sweep.${key}`, options);

function datasets(series: ReadonlyArray<StatSweepSeries>, metric: Metric): Array<ChartDataset<'line', SweepDatum[]>> {
	const { value, ci } = METRICS[metric];
	const surface = cssVarColor('--color-background');
	return series.flatMap(run => {
		const color = seriesColor(run.colorSlot);
		const at = (y: (point: StatSweepPoint) => number) =>
			run.points.map(point => ({ x: sweepChange(point.statValue, run.baseStatValue), y: y(point), point, statLabel: run.statLabel }));
		const edge = { yAxisID: metric, pointRadius: 0, pointHoverRadius: 0, borderWidth: 0, label: run.label };
		return [
			{ ...edge, data: at(point => value(point) + ci(point)), backgroundColor: color + BAND_ALPHA, fill: '+1' },
			{ ...edge, data: at(point => value(point) - ci(point)), fill: false },
			{
				yAxisID: metric,
				label: run.label,
				data: at(value),
				borderColor: color,
				backgroundColor: color,
				borderWidth: 2,
				borderJoinStyle: 'round',
				borderCapStyle: 'round',
				pointRadius: 4,
				pointHoverRadius: 6,
				pointBorderColor: surface,
				pointBorderWidth: 2,
				fill: false,
			},
		];
	});
}

// The current value sits at 0, so the range always reaches it.
function xRange(series: ReadonlyArray<StatSweepSeries>) {
	const changes = series.flatMap(run => run.points.map(point => sweepChange(point.statValue, run.baseStatValue)));
	const min = Math.min(0, ...changes);
	const max = Math.max(0, ...changes);
	const pad = (max - min) * X_RANGE_PAD || 1;
	return { min: min - pad, max: max + pad };
}

const signed = (change: number) => (change > 0 ? `+${change}` : String(change));

const markerPlugin: Plugin<'line'> = {
	id: 'statSweepMarker',
	afterDatasetsDraw: chart => {
		const { ctx, chartArea, scales } = chart;
		const line = (x: number, dash: number[]) => {
			if (x < chartArea.left || x > chartArea.right) return;
			ctx.save();
			ctx.strokeStyle = cssVarColor('--color-muted');
			ctx.lineWidth = 1;
			ctx.setLineDash(dash);
			ctx.beginPath();
			ctx.moveTo(x, chartArea.top);
			ctx.lineTo(x, chartArea.bottom);
			ctx.stroke();
			ctx.restore();
		};
		line(scales.x.getPixelForValue(0), [4, 4]);
		const hovered = chart.tooltip?.getActiveElements()[0];
		if (hovered) line(hovered.element.x, []);
	},
};

const PLUGINS = [markerPlugin];

// The EP and DPS panels stack on one left edge over the shared x scale. Chart.js puts the stacked scale with the higher
// weight on top, and `offset` keeps each panel's end ticks half a tick step inside it, so the labels where the panels
// meet do not collide.
const panelScale = (metric: Metric, weight: number) =>
	({
		type: 'linear',
		position: 'left',
		stack: 'sweep',
		stackWeight: 1,
		weight,
		offset: true,
		beginAtZero: false,
		grace: Y_GRACE,
		title: { display: true, text: t(METRICS[metric].title) },
		grid: { color: AXIS_GRID_COLOR },
	}) as const;

function options(range: { min: number; max: number }, showLegend: boolean): ChartOptions<'line'> {
	return {
		animation: false,
		responsive: true,
		maintainAspectRatio: false,
		// Runs of different ranges or steps hold different changes at one index, so the tooltip lists every run at the hovered change, not at the hovered index.
		interaction: { mode: 'nearest', axis: 'x', intersect: false },
		plugins: {
			legend: {
				display: showLegend,
				align: 'start',
				labels: {
					usePointStyle: true,
					pointStyle: 'line',
					// The default items take their stroke from the point border, which is the surface colour, so the swatch would vanish.
					generateLabels: chart =>
						chart.data.datasets.flatMap((dataset, datasetIndex) => {
							if (!isLine(datasetIndex) || (dataset as ChartDataset<'line'>).yAxisID !== 'ep') return [];
							const color = (dataset as ChartDataset<'line'>).borderColor as string;
							return [
								{
									text: dataset.label ?? '',
									datasetIndex,
									fillStyle: color,
									strokeStyle: color,
									lineWidth: LEGEND_LINE_WIDTH,
									pointStyle: 'line',
									fontColor: cssVarColor('--color-foreground'),
									hidden: !chart.isDatasetVisible(datasetIndex),
								},
							];
						}),
				},
			},
			tooltip: {
				filter: item => isLine(item.datasetIndex),
				callbacks: {
					// One line per distinct stat total, so runs of one stat from one base value share a line.
					title: items =>
						distinct(
							items.map(item => {
								const { x, point, statLabel } = item.raw as SweepDatum;
								return t('tooltip_total', { stat: statLabel, total: point.statValue, change: signed(x) });
							}),
						),
					label: item => {
						const { digits, title, ci } = METRICS[item.dataset.yAxisID as Metric];
						const { point, y } = item.raw as SweepDatum;
						return `${item.dataset.label} (${t(title)}): ${t('tooltip_value', { value: y.toFixed(digits), ci: ci(point).toFixed(digits) })}`;
					},
				},
			},
		},
		scales: {
			x: {
				type: 'linear',
				min: range.min,
				max: range.max,
				title: { display: true, text: t('change_axis') },
				ticks: { includeBounds: false },
				grid: { color: AXIS_GRID_COLOR },
			},
			ep: panelScale('ep', 1),
			dps: panelScale('dps', 0),
		},
	};
}

export const StatSweepChart = ({ series }: StatSweepChartProps) => {
	const data = useMemo(() => ({ datasets: [...datasets(series, 'ep'), ...datasets(series, 'dps')] }) satisfies ChartData<'line', SweepDatum[]>, [series]);
	const chartOptions = useMemo(() => options(xRange(series), series.length >= 2), [series]);

	if (!series.some(run => run.points.length)) {
		return <p className="text-muted">{t('empty')}</p>;
	}

	return (
		<figure className="m-0 grid gap-2">
			<div className="h-112 w-full">
				<Line data={data} options={chartOptions} plugins={PLUGINS} />
			</div>
			<figcaption className="flex items-center gap-2 text-sm text-muted">
				<span className="h-4 border-l border-dashed border-muted" />
				{t('current_value')}
			</figcaption>
		</figure>
	);
};
