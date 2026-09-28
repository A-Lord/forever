export interface NumberPickerOptions {
	// Whether the picker represents a float value. Default `false`
	float?: boolean;
	maxDecimalDigits?: number;
	// Whether to only allow positive values. Default `false`
	positive?: boolean;
	// Whether to show values of zero within the input. Default `true`
	showZeroes?: boolean;
	inputClassName?: string;
}
