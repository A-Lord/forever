import { BoundBooleanPicker } from '@features/inputs/components/BoundBooleanPicker';
import { BoundNumberPicker } from '@features/inputs/components/BoundNumberPicker';
import type { TargetInput } from '@generated/proto/common';
import type { Encounter } from '@sim/raid/encounter';
import { EnumPicker } from '@ui-kit/EnumPicker';
import { useMemo } from 'react';

import { targetInputConfig } from './utils/configs';

export interface TargetInputPickerProps {
	encounter: Encounter;
	targetIndex: number;
	inputIndex: number;
	input: TargetInput | undefined;
}

export const TargetInputPicker = ({ encounter, targetIndex, inputIndex, input }: TargetInputPickerProps) => {
	const picker = useMemo(
		() => (input ? targetInputConfig(encounter, targetIndex, inputIndex, input) : { kind: 'none' as const }),
		[encounter, targetIndex, inputIndex, input],
	);

	return (
		<div className="ui-field" data-testid="target-input-picker-root">
			{picker.kind === 'number' && <BoundNumberPicker modObject={null} config={picker.config} />}
			{picker.kind === 'boolean' && <BoundBooleanPicker modObject={null} config={picker.config} />}
			{picker.kind === 'enum' && <EnumPicker modObject={null} config={picker.config} />}
		</div>
	);
};
