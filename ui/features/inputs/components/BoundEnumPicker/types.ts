import type { EnumPickerProps } from '@ui-kit/EnumPicker';

import type { InputConfig } from '../../input';

export type EnumPickerConfig<ModObject> = InputConfig<ModObject, number> & Pick<EnumPickerProps, 'values'> & { id: string };
