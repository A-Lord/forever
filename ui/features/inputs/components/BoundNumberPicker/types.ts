import type { NumberPickerProps } from '@ui-kit/NumberPicker';

import type { InputConfig } from '../../input';

export type NumberPickerConfig<ModObject> = InputConfig<ModObject, number> &
	Pick<NumberPickerProps, 'float' | 'maxDecimalDigits' | 'positive' | 'showZeroes' | 'inputClassName'> & { id: string };
