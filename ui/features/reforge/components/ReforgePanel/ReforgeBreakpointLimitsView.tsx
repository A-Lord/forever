import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Table, TableCell, TableRow } from '@ui-kit/Table';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { type ReactNode, useId } from 'react';

export interface ReforgeBreakpointLimitRow {
	key: string;
	statName: string;
	picker: ReactNode;
}

export interface ReforgeBreakpointLimitsViewProps {
	rows: ReadonlyArray<ReforgeBreakpointLimitRow>;
}

export const ReforgeBreakpointLimitsView = ({ rows }: ReforgeBreakpointLimitsViewProps) => {
	const tooltipId = useId();
	return (
		<>
			<div className="flex pb-4">
				<h6 className="mr-1 mb-0">{i18n.t('sidebar.buttons.suggest_reforges.breakpoint_limit')}</h6>
				<Button variant="unstyled" className="inline" {...tooltipAnchorProps(tooltipId)}>
					<Icon name="circle-question" style="regular" />
				</Button>
				<Tooltip id={tooltipId} content={i18n.t('sidebar.buttons.suggest_reforges.breakpoint_limit_tooltip')} />
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
