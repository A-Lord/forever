import { EnumPicker } from '@ui-kit/EnumPicker';

import { useBoundField } from '../../hooks/useBoundField';
import type { EnumPickerConfig } from './types';

export interface BoundEnumPickerProps<ModObject> {
	modObject: ModObject;
	config: EnumPickerConfig<ModObject>;
	ariaLabel?: string;
	selectClassName?: string;
	testId?: string;
}

export const BoundEnumPicker = <ModObject,>({ modObject, config, ariaLabel, selectClassName, testId }: BoundEnumPickerProps<ModObject>) => {
	const { value, setValue, hidden, disabled, chrome } = useBoundField(modObject, config);
	if (hidden) return null;
	return (
		<EnumPicker
			{...chrome}
			testId={testId}
			value={value}
			onChange={setValue}
			disabled={disabled}
			values={config.values}
			ariaLabel={ariaLabel}
			selectClassName={selectClassName}
		/>
	);
};
