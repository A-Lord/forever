export interface EnumValueConfig {
	name: string;
	value: number;
	tooltip?: string;
}

export interface EnumPickerOptions {
	values: Array<EnumValueConfig>;
	/** For a select with no visible label. `label` renders one; this names it without adding markup. */
	ariaLabel?: string;
	selectClassName?: string;
}
