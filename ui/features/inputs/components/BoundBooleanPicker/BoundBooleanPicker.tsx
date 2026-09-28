import { BooleanPicker } from '@ui-kit/BooleanPicker';

import { useBoundField } from '../../hooks/useBoundField';
import type { BooleanPickerConfig } from './types';

export interface BoundBooleanPickerProps<ModObject> {
	modObject: ModObject;
	config: BooleanPickerConfig<ModObject>;
	testId?: string;
}

export const BoundBooleanPicker = <ModObject,>({ modObject, config, testId }: BoundBooleanPickerProps<ModObject>) => {
	const { value, setValue, hidden, disabled, chrome } = useBoundField(modObject, config);
	if (hidden) return null;
	return <BooleanPicker {...chrome} testId={testId} value={value} onChange={setValue} disabled={disabled} reverse={config.reverse} />;
};
