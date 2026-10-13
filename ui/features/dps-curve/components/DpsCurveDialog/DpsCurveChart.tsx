import { formatAverage, formatStdev } from '@features/results/components/ResultMetricList/utils';
import { AXIS_GRID_COLOR, cssVarColor } from '@features/results/components/Timeline/chart/colors';
import i18n from '@i18n/config';
import type { DpsCurvePoint } from '@sim/cache/dps_curve_history';
import { distinct } from '@sim/utils/collections';
import type { ChartData, ChartDataset, ChartOptions, Plugin } from 'chart.js';
import { useMemo } from 'react';
import { Line } from 'react-chartjs-2';

import { curveChange } from '../../model/dps_curve';
import { seriesColor } from './colors';

export interface DpsCurveSeries {
	id: string;
	label: string;
	colorSlot: number;
	statLabel: string;
	/** The swept stat's total when the run started; the run plots each point at its change from it. */
	baseStatValue: number;
	points: ReadonlyArray<DpsCurvePoint>;
}

export interface DpsCurveChartProps {
	series: ReadonlyArray<DpsCurveSeries>;
}

type CurveDatum = { x: number; y: number; point: DpsCurvePoint; statLabel: string };

// Each run draws the band's upper edge, its lower edge, then the DPS line. Upper fills down to the next dataset, so the order is load-bearing.
// The band is one standard deviation of a single fight's DPS either side of the average, the spread the sim card's ± reports.
const DATASETS_PER_RUN = 3;
const LINE_INDEX = 2;
const isLine = (datasetIndex: number) => datasetIndex % DATASETS_PER_RUN === LINE_INDEX;
// The x range reaches this share of its span past the outermost value, so the end points draw whole inside the plot.
const X_RANGE_PAD = 0.03;
const BAND_ALPHA = '1a';
const LEGEND_LINE_WIDTH = 3;
// The y scale fits its data, with this share of the span as room past the extremes.
const Y_GRACE = '5%';

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`sidebar.buttons.dps_curve.${key}`, options);

// The sim card's own formatting of its DPS and its ±.
const formatDps = (average: number) => formatAverage({ metric: 'dps', name: '', average }, 'list');
const formatDpsError = (stdev: number) => formatStdev({ metric: 'dps', name: '', average: 0, stdev }, 'list');

function datasets(series: ReadonlyArray<DpsCurveSeries>): Array<ChartDataset<'line', CurveDatum[]>> {
	const surface = cssVarColor('--color-background');
	return series.flatMap(run => {
		const color = seriesColor(run.colorSlot);
		const at = (y: (point: DpsCurvePoint) => number) =>
			run.points.map(point => ({ x: curveChange(point.statValue, run.baseStatValue), y: y(point), point, statLabel: run.statLabel }));
		const edge = { pointRadius: 0, pointHoverRadius: 0, borderWidth: 0, label: run.label };
		return [
			{ ...edge, data: at(point => point.dps + point.dpsStdev), backgroundColor: color + BAND_ALPHA, fill: '+1' },
			{ ...edge, data: at(point => point.dps - point.dpsStdev), fill: false },
			{
				label: run.label,
				data: at(point => point.dps),
				borderColor: color,
				borderWidth: 2,
				borderJoinStyle: 'round',
				borderCapStyle: 'round',
				fill: false,
				backgroundColor: color,
				pointRadius: 4,
				pointHoverRadius: 6,
				pointBorderColor: surface,
				pointBorderWidth: 2,
			},
		];
	});
}

// The current value sits at 0, so the range always reaches it.
function xRange(series: ReadonlyArray<DpsCurveSeries>) {
	const changes = series.flatMap(run => run.points.map(point => curveChange(point.statValue, run.baseStatValue)));
	const min = Math.min(0, ...changes);
	const max = Math.max(0, ...changes);
	const pad = (max - min) * X_RANGE_PAD || 1;
	return { min: min - pad, max: max + pad };
}

const signed = (change: number) => (change > 0 ? `+${change}` : String(change));

const markerPlugin: Plugin<'line'> = {
	id: 'dpsCurveMarker',
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
				position: 'bottom',
				align: 'start',
				labels: {
					usePointStyle: true,
					pointStyle: 'line',
					// The default items take their stroke from the point border, which is the surface colour, so the swatch would vanish.
					generateLabels: chart =>
						chart.data.datasets.flatMap((dataset, datasetIndex) => {
							if (!isLine(datasetIndex)) return [];
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
								const { x, point, statLabel } = item.raw as CurveDatum;
								return t('tooltip_total', { stat: statLabel, total: point.statValue, change: signed(x) });
							}),
						),
					label: item => {
						const { point, y } = item.raw as CurveDatum;
						return `${item.dataset.label}: ${t('tooltip_value', { value: formatDps(y), stdev: formatDpsError(point.dpsStdev) })}`;
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
			y: {
				type: 'linear',
				beginAtZero: false,
				grace: Y_GRACE,
				title: { display: true, text: t('dps_axis'), color: cssVarColor('--color-foreground') },
				grid: { color: AXIS_GRID_COLOR },
			},
		},
	};
}

export const DpsCurveChart = ({ series }: DpsCurveChartProps) => {
	const data = useMemo(() => ({ datasets: datasets(series) }) satisfies ChartData<'line', CurveDatum[]>, [series]);
	const chartOptions = useMemo(() => options(xRange(series), series.length >= 2), [series]);
	const hasPoints = series.some(run => run.points.length);

	// The empty text sits in the chart's box, so the first point does not resize the dialog.
	return (
		<figure className="m-0 flex min-h-0 min-w-0 flex-1 flex-col gap-2">
			<div className="relative flex min-h-112 w-full min-w-0 flex-1 items-center justify-center lg:min-h-64">
				{hasPoints ? <Line data={data} options={chartOptions} plugins={PLUGINS} /> : <p className="mb-0 text-muted">{t('empty')}</p>}
			</div>
			<figcaption className="flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-muted">
				<span className="flex items-center gap-2">
					<span className="h-4 border-l border-dashed border-muted" />
					{t('current_value')}
				</span>
				<span className="flex items-center gap-2">
					<span className="h-4 w-4 rounded-sm bg-muted/30" />
					{t('band_key')}
				</span>
			</figcaption>
		</figure>
	);
};
