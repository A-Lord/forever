import type { SpellSchool } from '@generated/proto/common';
import i18n from '@i18n/config';
import { formatToCompactNumber } from '@sim/utils/format';
import { Table, TableCell, TableHeaderCell, TableRow } from '@ui-kit/Table';
import { Fragment } from 'react';

import { MetricsTotalBar } from '../MetricsTotalBar';

export interface MetricsCombinedTooltipEntry {
	name: string;
	value: number;
	percentage: number;
	average?: number;
	min?: number;
	max?: number;
}

export interface MetricsCombinedTooltipGroup {
	name?: string;
	className?: string;
	totalPercentage: number;
	spellSchool?: SpellSchool | null;
	data: Array<MetricsCombinedTooltipEntry>;
}

export interface MetricsCombinedTooltipProps {
	groups: Array<MetricsCombinedTooltipGroup>;
	/** Overrides for the Type / Count / Average / Min / Max headers, by position. */
	headerValues?: Array<string | undefined>;
}

/** The tooltip *body* only — a nested breakdown table. It parameterises the rows and the bars; the anchor, the open/close and the veto belong to the column's one `<Tooltip>`, which is what keeps a 19-row table at seven tooltip instances instead of 133. */
export const MetricsCombinedTooltip = ({ groups, headerValues }: MetricsCombinedTooltipProps) => {
	const displayGroups = groups
		.filter(group => group.data.some(entry => entry.value))
		.map(group => ({ ...group, data: group.data.filter(entry => entry.value) }));
	const hasColumn = (field: 'average' | 'min' | 'max') => displayGroups.some(group => group.data.some(entry => typeof entry[field] === 'number'));
	const hasAverageColumn = hasColumn('average');
	const hasMinColumn = hasColumn('min');
	const hasMaxColumn = hasColumn('max');
	const columnCount = 2 + Number(hasAverageColumn) + Number(hasMinColumn) + Number(hasMaxColumn);

	return (
		<Table striped hover className="ui-metrics-table" testId="metrics-table">
			<thead data-testid="metrics-table-header">
				<TableRow testId="metrics-table-header-row">
					<TableHeaderCell testId="metrics-table-header-cell" className="ui-metrics-header-cell">
						{headerValues?.[0] || i18n.t('results_tab.details.tooltip_table.type')}
					</TableHeaderCell>
					<TableHeaderCell testId="metrics-table-header-cell" className="ui-metrics-header-cell">
						{headerValues?.[1] || i18n.t('results_tab.details.tooltip_table.count')}
					</TableHeaderCell>
					{hasMinColumn && (
						<TableHeaderCell testId="metrics-table-header-cell" className="ui-metrics-header-cell">
							{headerValues?.[3] || i18n.t('results_tab.details.tooltip_table.min')}
						</TableHeaderCell>
					)}
					{hasAverageColumn && (
						<TableHeaderCell testId="metrics-table-header-cell" className="ui-metrics-header-cell">
							{headerValues?.[2] || i18n.t('results_tab.details.tooltip_table.average')}
						</TableHeaderCell>
					)}
					{hasMaxColumn && (
						<TableHeaderCell testId="metrics-table-header-cell" className="ui-metrics-header-cell">
							{headerValues?.[4] || i18n.t('results_tab.details.tooltip_table.max')}
						</TableHeaderCell>
					)}
				</TableRow>
			</thead>
			<tbody data-testid="metrics-table-body">
				{displayGroups.map(({ name: groupName, className, data, spellSchool, totalPercentage }, groupIndex) => {
					const maxValue = Math.max(...data.map(entry => entry.value));
					return (
						<Fragment key={groupName ?? groupIndex}>
							{groupName && displayGroups.length > 1 && (
								<TableRow testId="metrics-table-group-header" className={className}>
									<TableHeaderCell className="p-0 text-left font-normal" colSpan={columnCount}>
										{groupName}
									</TableHeaderCell>
								</TableRow>
							)}
							{[...data]
								.sort((a, b) => b.value - a.value)
								.map(({ name, value, percentage, average, min, max }) => (
									<TableRow className={className} key={name}>
										<TableCell className="ui-metrics-cell">{name}</TableCell>
										<TableCell className="ui-metrics-cell">
											<MetricsTotalBar
												spellSchool={spellSchool}
												percentage={(percentage / totalPercentage) * 100}
												max={maxValue}
												total={value}
												value={value}
											/>
										</TableCell>
										{hasMinColumn && (
											<TableCell className="ui-metrics-cell">{typeof min === 'number' ? formatToCompactNumber(min) : ''}</TableCell>
										)}
										{hasAverageColumn && (
											<TableCell className="ui-metrics-cell">
												{typeof average === 'number' ? formatToCompactNumber(average) : ''}
											</TableCell>
										)}
										{hasMaxColumn && (
											<TableCell className="ui-metrics-cell">{typeof max === 'number' ? formatToCompactNumber(max) : ''}</TableCell>
										)}
									</TableRow>
								))}
						</Fragment>
					);
				})}
			</tbody>
		</Table>
	);
};
