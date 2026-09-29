import type { ControlledProps, FieldChromeProps } from '@ui-kit/Field';

export interface EnumValueConfig {
	name: string;
	value: number;
	tooltip?: string;
}

export interface MultiComboBoxProps extends FieldChromeProps, ControlledProps<Array<number>> {
	values: Array<EnumValueConfig>;
	placeholder?: string;
	emptyText?: string;
	removeLabel?: string;
	openLabel?: string;
}
