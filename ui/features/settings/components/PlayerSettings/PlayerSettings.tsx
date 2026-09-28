import { BoundEnumPicker } from '@features/inputs/components/BoundEnumPicker';
import { BoundIconInput } from '@features/inputs/components/BoundIconInput';
import { BoundInput } from '@features/inputs/components/BoundInput';
import type { IconInputConfig } from '@features/inputs/icon_inputs';
import { usePlayer } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import type { InputConfig } from '@sim/spec_config';
import { PickerGroup } from '@ui-kit/PickerGroup';
import { useMemo } from 'react';

import { professionInput, raceInput } from './utils';

export interface PlayerSettingsProps {
	iconInputs: ReadonlyArray<IconInputConfig<Player<any>, any>>;
	inputs: ReadonlyArray<InputConfig<Player<any>>>;
}

export const PlayerSettings = ({ iconInputs, inputs }: PlayerSettingsProps) => {
	const player = usePlayer() as Player<any>;
	const race = useMemo(() => raceInput(player), [player]);
	const professions = useMemo(() => [professionInput(1), professionInput(2)] as const, []);

	return (
		<>
			{iconInputs.length > 0 && (
				<PickerGroup variant="icons" className="ui-player-icon-group" data-testid="player-icon-group">
					{iconInputs.map((config, index) => (
						<BoundIconInput key={index} modObject={player} config={config} />
					))}
				</PickerGroup>
			)}
			<BoundEnumPicker modObject={player} config={race} />
			{inputs.map(config => (
				<BoundInput key={config.id} config={config} />
			))}
			<PickerGroup>
				{professions.map(config => (
					<BoundEnumPicker key={config.id} modObject={player} config={config} />
				))}
			</PickerGroup>
		</>
	);
};
