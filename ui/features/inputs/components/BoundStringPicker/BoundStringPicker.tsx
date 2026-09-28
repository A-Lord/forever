import { AdaptiveStringPicker } from '@ui-kit/AdaptiveStringPicker';

import { useBoundField } from '../../hooks/useBoundField';
import type { StringPickerConfig } from './types';

export interface BoundStringPickerProps<M> {
	modObject: M;
	config: StringPickerConfig<M>;
	testId?: string;
}

export const BoundStringPicker = <M,>({ modObject, config, testId }: BoundStringPickerProps<M>) => {
	const { value, setValue, hidden, disabled, chrome } = useBoundField(modObject, config);
	if (hidden) return null;
	return <AdaptiveStringPicker {...chrome} testId={testId} value={value} onChange={setValue} disabled={disabled} />;
};
