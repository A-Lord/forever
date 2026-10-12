import { AXIS_GRID_COLOR, cssVarColor } from '@features/results/components/Timeline/chart/colors';
import i18n from '@i18n/config';
import type { StatSweepPoint } from '@sim/cache/stat_sweep_history';
import type { ChartData, ChartDataset, ChartOptions, Plugin } from 'chart.js';
import { type RefObject, useMemo, useRef, useState } from 'react';
import { Line } from 'react-chartjs-2';

import { seriesColor } from './colors';

export interface StatSweepSeries {
	id: string;
	label: string;
	colorSlot: number;
	points: ReadonlyArray<StatSweepPoint>;
}

export interface StatSweepChartProps {
	series: ReadonlyArray<StatSweepSeries>;
	statLabel: string;
	/** The swept stat's current total, marked across both panels. */
	currentValue: number;
}

type Metric = 'ep' | 'dps';
type SweepDatum = { x: number; y: number; point: StatSweepPoint };

// Each run draws the band's upper edge, its lower edge, then the line. Upper fills down to the next dataset, so the order is load-bearing.
const DATASETS_PER_RUN = 3;
const LINE_INDEX = 2;
const isLine = (datasetIndex: number) => datasetIndex % DATASETS_PER_RUN === LINE_INDEX;
// The x range reaches this share of its span past the outermost value, so the end points draw whole inside the plot.
const X_RANGE_PAD = 0.03;
const BAND_ALPHA = '1a';

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
		const at = (y: (point: StatSweepPoint) => number) => run.points.map(point => ({ x: point.statValue, y: y(point), point }));
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

function xRange(series: ReadonlyArray<StatSweepSeries>, currentValue: number) {
	const values = series.flatMap(run => run.points.map(point => point.statValue));
	const min = Math.min(currentValue, ...values);
	const max = Math.max(currentValue, ...values);
	const pad = (max - min) * X_RANGE_PAD || 1;
	return { min: min - pad, max: max + pad };
}

function markerPlugin(current: RefObject<number>): Plugin<'line'> {
	return {
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
			line(scales.x.getPixelForValue(current.current), [4, 4]);
			const hovered = chart.tooltip?.getActiveElements()[0];
			if (hovered) line(hovered.element.x, []);
		},
	};
}

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
		title: { display: true, text: t(METRICS[metric].title) },
		grid: { color: AXIS_GRID_COLOR },
	}) as const;

function options(range: { min: number; max: number }, statLabel: string, showLegend: boolean): ChartOptions<'line'> {
	return {
		animation: false,
		responsive: true,
		maintainAspectRatio: false,
		// Runs swept from different base values share few x positions, so the tooltip lists every run at the hovered stat value, not at the hovered index.
		interaction: { mode: 'nearest', axis: 'x', intersect: false },
		plugins: {
			legend: {
				display: showLegend,
				align: 'start',
				labels: {
					usePointStyle: true,
					pointStyle: 'line',
					filter: (item, data) => {
						const datasetIndex = item.datasetIndex ?? 0;
						return isLine(datasetIndex) && (data.datasets[datasetIndex] as ChartDataset<'line'>).yAxisID === 'ep';
					},
				},
			},
			tooltip: {
				filter: item => isLine(item.datasetIndex),
				callbacks: {
					title: items => (items.length ? `${statLabel} ${items[0].parsed.x}` : ''),
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
				title: { display: true, text: statLabel },
				ticks: { includeBounds: false },
				grid: { color: AXIS_GRID_COLOR },
			},
			ep: panelScale('ep', 1),
			dps: panelScale('dps', 0),
		},
	};
}

export const StatSweepChart = ({ series, statLabel, currentValue }: StatSweepChartProps) => {
	const current = useRef(currentValue);
	current.current = currentValue;
	const [plugins] = useState(() => [markerPlugin(current)]);

	const data = useMemo(() => ({ datasets: [...datasets(series, 'ep'), ...datasets(series, 'dps')] }) satisfies ChartData<'line', SweepDatum[]>, [series]);
	const chartOptions = useMemo(() => options(xRange(series, currentValue), statLabel, series.length >= 2), [series, statLabel, currentValue]);

	if (!series.some(run => run.points.length)) {
		return <p className="text-muted">{t('empty')}</p>;
	}

	return (
		<figure className="m-0 grid gap-2">
			<div className="h-112 w-full">
				<Line data={data} options={chartOptions} plugins={plugins} />
			</div>
			<figcaption className="flex items-center gap-2 text-sm text-muted">
				<span className="h-4 border-l border-dashed border-muted" />
				{t('current_value')}: {currentValue}
			</figcaption>
		</figure>
	);
};
