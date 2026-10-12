import { AXIS_GRID_COLOR, cssVarColor } from '@features/results/components/Timeline/chart/colors';
import i18n from '@i18n/config';
import type { ChartData, ChartDataset, ChartOptions, Plugin, Scale } from 'chart.js';
import { type RefObject, useMemo, useRef } from 'react';
import { Line } from 'react-chartjs-2';

import type { SweepPoint } from '../../model/stat_sweep';
import { seriesColor } from './colors';

export interface StatSweepSeries {
	id: string;
	label: string;
	colorSlot: number;
	points: ReadonlyArray<SweepPoint>;
}

export interface StatSweepChartProps {
	series: ReadonlyArray<StatSweepSeries>;
	statLabel: string;
	/** The swept stat's current total, marked on both charts. */
	currentValue: number;
}

type Metric = 'ep' | 'dps';
type SweepDatum = { x: number; y: number; point: SweepPoint };

// Each run draws the band's upper edge, its lower edge, then the line. Upper fills down to the next dataset, so the order is load-bearing.
const DATASETS_PER_RUN = 3;
const LINE_INDEX = 2;
const isLine = (datasetIndex: number) => datasetIndex % DATASETS_PER_RUN === LINE_INDEX;
// Both y axes take the same width, so a stat value sits at the same pixel in the EP and the DPS chart.
const Y_AXIS_WIDTH = 64;
// The x range reaches this share of its span past the outermost value, so the end points draw whole inside the plot.
const X_RANGE_PAD = 0.03;
const BAND_ALPHA = '1a';

const METRICS = {
	ep: { value: (point: SweepPoint) => point.ep, ci: (point: SweepPoint) => point.epCi, digits: 3, title: 'ep_chart' },
	dps: { value: (point: SweepPoint) => point.dps, ci: (point: SweepPoint) => point.dpsCi, digits: 1, title: 'dps_chart' },
} as const;

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`sidebar.buttons.stat_weights.sweep.${key}`, options);

function datasets(series: ReadonlyArray<StatSweepSeries>, metric: Metric): Array<ChartDataset<'line', SweepDatum[]>> {
	const { value, ci } = METRICS[metric];
	const surface = cssVarColor('--color-background');
	return series.flatMap(run => {
		const color = seriesColor(run.colorSlot);
		const at = (y: (point: SweepPoint) => number) => run.points.map(point => ({ x: point.statValue, y: y(point), point }));
		const edge = { pointRadius: 0, pointHoverRadius: 0, borderWidth: 0, label: run.label };
		return [
			{ ...edge, data: at(point => value(point) + ci(point)), backgroundColor: color + BAND_ALPHA, fill: '+1' },
			{ ...edge, data: at(point => value(point) - ci(point)), fill: false },
			{
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

function options(metric: Metric, range: { min: number; max: number }, statLabel: string, showLegend: boolean, showX: boolean): ChartOptions<'line'> {
	const { digits, title, ci } = METRICS[metric];
	const fitYAxis = (scale: Scale) => {
		scale.width = Y_AXIS_WIDTH;
	};
	return {
		animation: false,
		responsive: true,
		maintainAspectRatio: false,
		// Runs swept from different base values share few x positions, so the tooltip lists every run at the hovered stat value, not at the hovered index.
		interaction: { mode: 'nearest', axis: 'x', intersect: false },
		plugins: {
			title: { display: true, text: t(title), align: 'start' },
			legend: {
				display: showLegend,
				align: 'start',
				labels: { usePointStyle: true, pointStyle: 'line', filter: item => isLine(item.datasetIndex ?? 0) },
			},
			tooltip: {
				filter: item => isLine(item.datasetIndex),
				callbacks: {
					title: items => (items.length ? `${statLabel} ${items[0].parsed.x}` : ''),
					label: item => {
						const { point, y } = item.raw as SweepDatum;
						return `${item.dataset.label}: ${t('tooltip_value', { value: y.toFixed(digits), ci: ci(point).toFixed(digits) })}`;
					},
				},
			},
		},
		scales: {
			// Both charts lay out the same tick labels, see-through on the EP chart, so the edge labels pad both plots alike and the x positions line up.
			x: {
				type: 'linear',
				min: range.min,
				max: range.max,
				title: { display: showX, text: statLabel },
				ticks: { includeBounds: false, ...(showX ? {} : { color: 'transparent' }) },
				grid: { color: AXIS_GRID_COLOR },
			},
			y: { afterFit: fitYAxis, grid: { color: AXIS_GRID_COLOR } },
		},
	};
}

export const StatSweepChart = ({ series, statLabel, currentValue }: StatSweepChartProps) => {
	const current = useRef(currentValue);
	current.current = currentValue;
	const plugins = useRef<Array<Plugin<'line'>>>(undefined as never);
	if (!plugins.current) plugins.current = [markerPlugin(current)];

	const charts = useMemo(() => {
		const range = xRange(series, currentValue);
		const showLegend = series.length >= 2;
		return (['ep', 'dps'] as const).map(metric => ({
			metric,
			data: { datasets: datasets(series, metric) } satisfies ChartData<'line', SweepDatum[]>,
			options: options(metric, range, statLabel, showLegend && metric === 'ep', metric === 'dps'),
		}));
	}, [series, statLabel, currentValue]);

	if (!series.some(run => run.points.length)) {
		return <p className="text-muted">{t('empty')}</p>;
	}

	return (
		<figure className="m-0 grid gap-2">
			{charts.map(chart => (
				<div key={chart.metric} className="h-56 w-full" data-metric={chart.metric}>
					<Line data={chart.data} options={chart.options} plugins={plugins.current} />
				</div>
			))}
			<figcaption className="flex items-center gap-2 text-sm text-muted">
				<span className="h-4 border-l border-dashed border-muted" />
				{t('current_value')}: {currentValue}
			</figcaption>
		</figure>
	);
};
