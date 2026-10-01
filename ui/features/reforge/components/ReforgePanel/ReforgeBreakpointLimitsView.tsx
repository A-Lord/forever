import i18n from '@i18n/config';
import { Table, TableCell, TableRow } from '@ui-kit/Table';
import { TooltipButton } from '@ui-kit/TooltipButton';
import type { ReactNode } from 'react';

export interface ReforgeBreakpointLimitRow {
	key: string;
	statName: string;
	picker: ReactNode;
}

export interface ReforgeBreakpointLimitsViewProps {
	rows: ReadonlyArray<ReforgeBreakpointLimitRow>;
}

export const ReforgeBreakpointLimitsView = ({ rows }: ReforgeBreakpointLimitsViewProps) => {
	return (
		<>
			<div className="flex pb-4">
				<h6 className="mr-1 mb-0">{i18n.t('sidebar.buttons.suggest_reforges.breakpoint_limit')}</h6>
				<TooltipButton
					tooltip={i18n.t('sidebar.buttons.suggest_reforges.breakpoint_limit_tooltip')}
					label={i18n.t('common.help')}
					testId="reforge-breakpoint-limit-help"
				/>
			</div>
			<Table dense className="mb-2">
				<tbody>
					{rows.map(row => (
						<TableRow key={row.key}>
							<TableCell className="w-0 whitespace-nowrap">{row.statName}</TableCell>
							<TableCell className="w-18">{row.picker}</TableCell>
						</TableRow>
					))}
				</tbody>
			</Table>
		</>
	);
};
