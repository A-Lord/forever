import i18n from '@i18n/config';
import { TableCell, TableRow } from '@ui-kit/Table';
import { TooltipButton } from '@ui-kit/TooltipButton';
import type { ReactNode } from 'react';

export interface ReforgeStatCapRowViewProps {
	statName: string;
	tooltip?: ReactNode;
	percentage: ReactNode;
	undershoot: ReactNode;
	/** The spec's preset caps for this stat, on a row of their own. */
	presets?: ReactNode;
}

export const ReforgeStatCapRowView = ({ statName, tooltip, percentage, undershoot, presets }: ReforgeStatCapRowViewProps) => {
	return (
		<>
			<TableRow>
				<TableCell className="w-0 whitespace-nowrap">
					{statName} {!!tooltip && <TooltipButton tooltip={tooltip} label={i18n.t('common.help')} testId="reforge-stat-cap-help" />}
				</TableCell>
				<TableCell className="w-18">{percentage}</TableCell>
				<TableCell className="w-14 text-right">{undershoot}</TableCell>
			</TableRow>
			{presets && (
				<TableRow>
					<TableCell />
					<TableCell className="w-18">{presets}</TableCell>
					<TableCell />
				</TableRow>
			)}
		</>
	);
};
