import { useInput } from '@features/inputs/hooks/useInput';
import type { InputConfig } from '@features/inputs/input';
import type { APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { useSimHost } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { useMemo } from 'react';

import { RotationTypePickerView } from './RotationTypePickerView';

export const RotationTypePicker = ({ testId }: { testId?: string }) => {
	const player = useSimHost().player;

	const config = useMemo(
		(): InputConfig<Player<any>, APLRotationType> => ({
			storeField: 'rotation',
			getValue: (subject: Player<any>) => subject.getRotationType(),
			setValue: (subject: Player<any>, newValue: APLRotationType) => {
				subject.modifyAplRotation(rotation => {
					rotation.type = newValue;
				});
			},
		}),
		[],
	);
	const { value, setValue } = useInput(player, config);

	return <RotationTypePickerView value={value} hasSimple={player.hasSimpleRotationGenerator()} onChange={setValue} testId={testId} />;
};
