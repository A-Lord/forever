import { useSimStore } from '@sim/hooks/useSimStore';
import type { StatWeightActionSettings } from '@sim/settings/stat_weight_settings';

import type { StatSweepControls } from '../../hooks/useStatSweepControls';
import { EpPrecisionOptions } from '../EpPrecisionOptions/EpPrecisionOptions';
import { StatSweepChart } from './StatSweepChart';
import { StatSweepHistory } from './StatSweepHistory';
import { StatSweepPanel } from './StatSweepPanel';

export interface StatSweepProps {
	settings: StatWeightActionSettings;
	controls: StatSweepControls;
}

export const StatSweep = ({ settings, controls }: StatSweepProps) => {
	const { sweep, unitStats, unitStat, statOptions, range, step, precision } = controls;
	const defaultIterations = useSimStore('iterations');

	return (
		<StatSweepPanel
			statOptions={statOptions}
			stat={unitStat?.getKey() ?? ''}
			onStatChange={key => settings.setSweep({ sweepStat: unitStats.find(stat => stat.getKey() === key) ?? null })}
			range={range}
			onRangeChange={sweepRange => settings.setSweep({ sweepRange })}
			step={step}
			onStepChange={sweepStep => settings.setSweep({ sweepStep })}
			fractional={unitStat?.isPseudoStat()}
			precision={
				<EpPrecisionOptions
					value={precision}
					onChange={patch => settings.setPrecision(patch)}
					defaultIterations={defaultIterations}
					disabled={sweep.isRunning}
					idPrefix="stat-sweep-precision"
				/>
			}
			chart={<StatSweepChart series={sweep.series} />}
			history={<StatSweepHistory runs={sweep.historyEntries} onOverlayChange={sweep.setOverlaid} onDelete={sweep.remove} />}
			running={sweep.isRunning}
		/>
	);
};
