import i18n from '@i18n/config';
import { CopyButton } from '@ui-kit/CopyButton';

export interface TalentsHeaderViewProps {
	pointsRemaining: number;
	getTalentsString: () => string;
}

export const TalentsHeaderView = ({ pointsRemaining, getTalentsString }: TalentsHeaderViewProps) => (
	<div className="mb-1 flex w-full items-center">
		<label>
			{i18n.t('talents_tab.points_remaining', { defaultValue: 'Points Remaining' })}: {pointsRemaining}
		</label>
		<div className="ml-auto" data-testid="talents-picker-actions">
			<CopyButton
				getContent={getTalentsString}
				variant="outline-primary"
				size="sm"
				className="w-24"
				text={i18n.t('talents_tab.copy_button.label')}
				tooltip={i18n.t('talents_tab.copy_button.tooltip')}
			/>
		</div>
	</div>
);
