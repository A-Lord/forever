import { NumberPicker } from '@ui-kit/NumberPicker';
import clsx from 'clsx';

import { useBoundField } from '../../hooks/useBoundField';
import type { NumberPickerConfig } from './types';

export interface BoundNumberPickerProps<ModObject> {
	modObject: ModObject;
	config: NumberPickerConfig<ModObject>;
	inputClassName?: string;
	testId?: string;
}

export const BoundNumberPicker = <ModObject,>({ modObject, config, inputClassName, testId }: BoundNumberPickerProps<ModObject>) => {
	const { value, setValue, hidden, disabled, chrome } = useBoundField(modObject, config);
	if (hidden) return null;
	return (
		<NumberPicker
			{...chrome}
			testId={testId}
			value={value}
			onChange={setValue}
			disabled={disabled}
			float={config.float}
			positive={config.positive}
			showZeroes={config.showZeroes}
			maxDecimalDigits={config.maxDecimalDigits}
			inputClassName={clsx(config.inputClassName, inputClassName)}
		/>
	);
};
