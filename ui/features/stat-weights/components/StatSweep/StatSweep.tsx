import { useSimStore } from '@sim/hooks/useSimStore';
import type { StatWeightActionSettings } from '@sim/settings/stat_weight_settings';
import { useMemo } from 'react';

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
	const { sweep, unitStats, selectedStats, statOptions, range, step, precision } = controls;
	const defaultIterations = useSimStore('iterations');
	const selectedKeys = useMemo(() => selectedStats.map(stat => stat.getKey()), [selectedStats]);

	return (
		<StatSweepPanel
			statOptions={statOptions}
			stats={selectedKeys}
			onStatsChange={keys => settings.setSweep({ sweepStats: unitStats.filter(stat => keys.includes(stat.getKey())) })}
			range={range}
			onRangeChange={sweepRange => settings.setSweep({ sweepRange })}
			step={step}
			onStepChange={sweepStep => settings.setSweep({ sweepStep })}
			fractional={selectedStats.some(stat => stat.isPseudoStat())}
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
