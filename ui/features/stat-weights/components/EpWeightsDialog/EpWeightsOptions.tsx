import i18n from '@i18n/config';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { ChoiceField } from '@ui-kit/ChoiceField';

import { StatsType } from './types';

export interface EpWeightsOptionsProps {
	statsType: StatsType;
	showAllStats: boolean;
	onStatsTypeChange: (statsType: StatsType) => void;
	onShowAllStatsChange: (showAllStats: boolean) => void;
}

export const EpWeightsOptions = ({ statsType, showAllStats, onStatsTypeChange, onShowAllStatsChange }: EpWeightsOptionsProps) => (
	<div className="-mx-3 mb-3 flex flex-wrap">
		<div className="w-full px-3 sm:w-1/4">
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
		</div>
		<div className="w-full px-3 sm:w-1/4">
			<BooleanPicker
				id="ep-show-all-stats"
				label={i18n.t('sidebar.buttons.stat_weights.modal.show_all_stats')}
				layout="inline"
				value={showAllStats}
				onChange={onShowAllStatsChange}
			/>
		</div>
	</div>
);
