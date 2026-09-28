import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { useId } from 'react';

export interface TalentsHeaderViewProps {
	pointsRemaining: number;
	copied: boolean;
	onCopy: () => void;
}

export const TalentsHeaderView = ({ pointsRemaining, copied, onCopy }: TalentsHeaderViewProps) => {
	const copyTooltipId = useId();

	return (
		<div className="mb-1 flex w-full items-center">
			<label>
				{i18n.t('talents_tab.points_remaining', { defaultValue: 'Points Remaining' })}:{' '}
				<span data-testid="talents-picker-points-remaining">{pointsRemaining}</span>
			</label>
			<div className="ml-auto" data-testid="talents-picker-actions">
				<Button variant="outline-primary" size="sm" className="w-24" data-testid="copy-button" onClick={onCopy} {...tooltipAnchorProps(copyTooltipId)}>
					<Icon name={copied ? 'check' : 'copy'} className="mr-1" />
					{copied ? i18n.t('common.copy_button.copied') : i18n.t('talents_tab.copy_button.label')}
				</Button>
				<Tooltip id={copyTooltipId} content={i18n.t('talents_tab.copy_button.tooltip')} />
			</div>
		</div>
	);
};
