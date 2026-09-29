import { EnumPicker } from '@ui-kit/EnumPicker';

import { useConnectedField } from '../../hooks/useConnectedField';
import type { EnumPickerConfig } from './types';

export interface ConnectedEnumPickerProps<ModObject> {
	modObject: ModObject;
	config: EnumPickerConfig<ModObject>;
	ariaLabel?: string;
	selectClassName?: string;
	testId?: string;
}

export const ConnectedEnumPicker = <ModObject,>({ modObject, config, ariaLabel, selectClassName, testId }: ConnectedEnumPickerProps<ModObject>) => {
	const { value, setValue, hidden, disabled, chrome } = useConnectedField(modObject, config);
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
