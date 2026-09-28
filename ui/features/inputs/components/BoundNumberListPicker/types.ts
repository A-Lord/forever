import type { NumberListPickerProps } from '@ui-kit/NumberListPicker';

import type { InputConfig } from '../../input';

export type NumberListPickerConfig<M> = InputConfig<M, Array<number>> & Pick<NumberListPickerProps, 'placeholder'> & { id: string };
