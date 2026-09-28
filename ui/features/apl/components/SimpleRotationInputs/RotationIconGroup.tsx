import { BoundIconInput } from '@features/inputs/components/BoundIconInput';
import type { IconInputConfig } from '@features/inputs/icon_inputs';
import { usePlayer } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { PickerGroup } from '@ui-kit/PickerGroup';

export interface RotationIconGroupProps {
	inputs: ReadonlyArray<IconInputConfig<Player<any>, any>>;
}

/**
 * The rotation tab's icon row.
 *
 * `configureIconSection`'s `hide` branch is dropped: it fired only for an empty picker list, and
 * the one caller built the list from a `rotationIconInputs` array it had already checked was
 * non-empty, so that branch could never run.
 */
export const RotationIconGroup = ({ inputs }: RotationIconGroupProps) => {
	const player = usePlayer();

	return (
		<PickerGroup variant="icons" data-testid="rotation-icon-group">
			{inputs.map((config, index) => (
				<BoundIconInput key={index} modObject={player} config={config} />
			))}
		</PickerGroup>
	);
};
