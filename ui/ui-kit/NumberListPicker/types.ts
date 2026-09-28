import type { ControlledProps, FieldChromeProps } from '@ui-kit/Field';

export interface NumberListPickerProps extends FieldChromeProps, ControlledProps<Array<number>> {
	placeholder?: string;
}
