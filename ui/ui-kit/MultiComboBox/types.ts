import type { EnumValueConfig } from '@ui-kit/EnumPicker/types';
import type { ControlledProps, FieldChromeProps } from '@ui-kit/Field';

export interface MultiComboBoxProps extends FieldChromeProps, ControlledProps<Array<number>> {
	values: Array<EnumValueConfig>;
	placeholder?: string;
	emptyText?: string;
	removeLabel?: string;
	openLabel?: string;
}
