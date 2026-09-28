import i18n from '@i18n/config';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { EnumPicker } from '@ui-kit/EnumPicker';

import { StatsType } from './types';

export interface EpWeightsOptionsProps {
	statsType: StatsType;
	showAllStats: boolean;
	onStatsTypeChange: (statsType: StatsType) => void;
	onShowAllStatsChange: (showAllStats: boolean) => void;
}

const STATS_TYPES: StatsType[] = [StatsType.Ep, StatsType.Weight];

export const EpWeightsOptions = ({ statsType, showAllStats, onStatsTypeChange, onShowAllStatsChange }: EpWeightsOptionsProps) => (
	<div className="-mx-3 mb-3 flex flex-wrap">
		<div className="w-full px-3 sm:w-1/4">
			<EnumPicker
				id="ep-type-select"
				ariaLabel={i18n.t('sidebar.buttons.stat_weights.modal.stats_type')}
				values={[
					{ name: i18n.t('sidebar.buttons.stat_weights.modal.ep'), value: 0 },
					{ name: i18n.t('sidebar.buttons.stat_weights.modal.weights'), value: 1 },
				]}
				value={STATS_TYPES.indexOf(statsType)}
				onChange={next => onStatsTypeChange(STATS_TYPES[next])}
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
