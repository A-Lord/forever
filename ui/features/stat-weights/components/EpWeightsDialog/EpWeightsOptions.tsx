import i18n from '@i18n/config';
import { ChoiceField } from '@ui-kit/ChoiceField';
import type { ReactNode } from 'react';

import { StatsType } from './types';

export interface EpWeightsOptionsProps {
	statsType: StatsType;
	onStatsTypeChange: (statsType: StatsType) => void;
	children: ReactNode;
}

export const EpWeightsOptions = ({ statsType, onStatsTypeChange, children }: EpWeightsOptionsProps) => (
	<div className="flex items-end gap-2 [&_.ui-field-label]:font-bold">
		<ChoiceField
			id="ep-type-select"
			className="min-w-0 flex-1"
			label={i18n.t('sidebar.buttons.stat_weights.modal.show_as')}
			options={[
				{ value: StatsType.Ep, label: i18n.t('sidebar.buttons.stat_weights.modal.ep') },
				{ value: StatsType.Weight, label: i18n.t('sidebar.buttons.stat_weights.modal.weights') },
			]}
			value={statsType}
			onChange={onStatsTypeChange}
		/>
		<div className="flex items-center">{children}</div>
	</div>
);
