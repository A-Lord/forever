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
	<div className="mb-2 grid gap-x-4 sm:grid-cols-4">
		<ChoiceField
			id="ep-type-select"
			ariaLabel={i18n.t('sidebar.buttons.stat_weights.modal.stats_type')}
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
