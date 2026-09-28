import { NumberListPicker } from '@ui-kit/NumberListPicker';

import { useBoundField } from '../../hooks/useBoundField';
import type { NumberListPickerConfig } from './types';

export interface BoundNumberListPickerProps<M> {
	modObject: M;
	config: NumberListPickerConfig<M>;
	testId?: string;
}

export const BoundNumberListPicker = <M,>({ modObject, config, testId }: BoundNumberListPickerProps<M>) => {
	const { value, setValue, hidden, disabled, chrome } = useBoundField(modObject, config);
	if (hidden) return null;
	return <NumberListPicker {...chrome} testId={testId} value={value} onChange={setValue} disabled={disabled} placeholder={config.placeholder} />;
};
