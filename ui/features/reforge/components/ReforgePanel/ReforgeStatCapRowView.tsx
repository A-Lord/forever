import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { TableCell, TableRow } from '@ui-kit/Table';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { type ReactNode, useId } from 'react';

export interface ReforgeStatCapRowViewProps {
	statName: string;
	tooltip?: ReactNode;
	percentage: ReactNode;
	undershoot: ReactNode;
	/** The spec's preset caps for this stat, on a row of their own. */
	presets?: ReactNode;
}

export const ReforgeStatCapRowView = ({ statName, tooltip, percentage, undershoot, presets }: ReforgeStatCapRowViewProps) => {
	const tooltipId = useId();
	return (
		<>
			<TableRow>
				<TableCell className="w-0 whitespace-nowrap">
					{statName}{' '}
					{!!tooltip && (
						<>
							<Button variant="unstyled" className="inline" {...tooltipAnchorProps(tooltipId)}>
								<Icon name="circle-question" style="regular" />
							</Button>
							<Tooltip id={tooltipId} content={tooltip} />
						</>
					)}
				</TableCell>
				<TableCell className="w-18.75">{percentage}</TableCell>
				<TableCell className="w-15 text-right">{undershoot}</TableCell>
			</TableRow>
			{presets && (
				<TableRow>
					<TableCell />
					<TableCell className="w-18.75">{presets}</TableCell>
					<TableCell />
				</TableRow>
			)}
		</>
	);
};
