import { ConnectedChoiceField } from '@features/inputs/components/ConnectedChoiceField';
import { ConnectedNumberField } from '@features/inputs/components/ConnectedTextField';
import { useSimHost } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { useMemo, useState } from 'react';

import { AdvancedEncounterModal } from '../AdvancedEncounterModal';
import { TargetInputsPicker } from '../TargetsPicker';
import { AreaTypesPicker } from './AreaTypesPicker';
import { EncounterPickerView } from './EncounterPickerView';
import { durationConfigs, executeConfigs, minBaseDamageConfig, numAlliesConfig, presetEncounterConfig, zoneConfig } from './utils/configs';

export interface EncounterPickerProps {
	showExecuteProportion: boolean;
}

export const EncounterPicker = ({ showExecuteProportion }: EncounterPickerProps) => {
	const host = useSimHost();
	const encounter = host.sim.encounter;
	const player = host.player;
	const { damage: showDamage } = useDisplayMetrics(host.sim);

	const duration = useMemo(() => durationConfigs(encounter), [encounter]);
	const execute = useMemo(() => (showExecuteProportion ? executeConfigs(encounter) : []), [encounter, showExecuteProportion]);
	const preset = useMemo(() => presetEncounterConfig(encounter), [encounter]);
	const allies = useMemo(() => numAlliesConfig(player), [player]);
	const minBaseDamage = useMemo(() => minBaseDamageConfig(), []);
	const zone = useMemo(() => zoneConfig(encounter), [encounter]);

	const [advancedOpen, setAdvancedOpen] = useState(false);

	return (
		<>
			<EncounterPickerView
				duration={duration.map(config => (
					<ConnectedNumberField key={config.id} modObject={encounter} config={config} />
				))}
				execute={showExecuteProportion && execute.map(config => <ConnectedNumberField key={config.id} modObject={encounter} config={config} />)}
				zone={<ConnectedChoiceField modObject={encounter} config={zone} />}
				areaTypes={<AreaTypesPicker encounter={encounter} />}
				preset={showDamage && <ConnectedChoiceField modObject={encounter} config={preset} />}
				allies={player.canEnableTargetDummies() && <ConnectedNumberField modObject={host.sim.raid} config={allies} />}
				minBaseDamage={player.getPlayerSpec().isTankSpec && <ConnectedNumberField modObject={encounter} config={minBaseDamage} />}
				targetInputs={<TargetInputsPicker encounter={encounter} targetIndex={0} />}
				onAdvancedClick={() => setAdvancedOpen(true)}
			/>
			<AdvancedEncounterModal open={advancedOpen} onOpenChange={setAdvancedOpen} />
		</>
	);
};
