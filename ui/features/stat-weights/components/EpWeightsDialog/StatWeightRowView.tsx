import type { StatWeightValues } from '@generated/proto/api';
import type { UnitStat } from '@sim/proto/stats';
import { sanitizeId } from '@sim/utils/format';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { NumberPicker } from '@ui-kit/NumberPicker';
import { TableCell, TableRow } from '@ui-kit/Table';
import clsx from 'clsx';

import { StatWeightCells } from './StatWeightCells';

/** One shown metric's pair of cells: the weights it computed (none before a run, or for an excluded stat) and its EP ratio. */
export interface StatWeightRowMetric {
	ratioIndex: number;
	statWeights?: StatWeightValues;
	epRatio: number;
}

export interface StatWeightRowInclude {
	value: boolean;
	disabled: boolean;
	onChange: (included: boolean) => void;
}

export interface StatWeightRowViewProps {
	stat: UnitStat;
	fullName: string;
	shortName: string;
	metrics: StatWeightRowMetric[];
	iterations: number;
	/** The row's weighted EP minus its current EP weight, which colours the EP cells. */
	epDelta: number;
	currentEp: number;
	onCurrentEpChange: (value: number) => void;
	/** Absent for a stat outside the spec's EP set: the Update cell stays empty. */
	include?: StatWeightRowInclude;
	showThreatMetrics: boolean;
	showUpdateColumn?: boolean;
}

export const StatWeightRowView = ({
	stat,
	fullName,
	shortName,
	metrics,
	iterations,
	epDelta,
	currentEp,
	onCurrentEpChange,
	include,
	showThreatMetrics,
	showUpdateColumn = true,
}: StatWeightRowViewProps) => {
	const cellClassName = clsx('ui-ep-weights-table-cell', showThreatMetrics && 'max-lg:pl-0');

	return (
		<TableRow className="odd:bg-(--table-row-odd-bg) even:bg-(--table-row-even-bg)">
			<TableCell className={cellClassName}>{fullName}</TableCell>
			{showUpdateColumn && (
				<TableCell testId="swcalc-include-toggle" className={cellClassName}>
					{include && (
						<BooleanPicker
							id={`sw-stat-toggle-${sanitizeId(fullName)}`}
							value={include.value}
							disabled={include.disabled}
							onChange={include.onChange}
						/>
					)}
				</TableCell>
			)}
			{metrics.map(({ statWeights, ratioIndex, epRatio }) => (
				<StatWeightCells
					key={ratioIndex}
					stat={stat}
					statWeights={statWeights}
					iterations={iterations}
					epRatio={epRatio}
					epDelta={epDelta}
					cellClassName={cellClassName}
				/>
			))}
			<TableCell testId="current-ep" className={cellClassName}>
				<NumberPicker
					id={`ep-weight-stat-${sanitizeId(shortName)}`}
					className="mb-0"
					float
					value={currentEp}
					onChange={onCurrentEpChange}
					inputClassName={showThreatMetrics ? 'ui-ep-weights-compact-input' : undefined}
				/>
			</TableCell>
		</TableRow>
	);
};
