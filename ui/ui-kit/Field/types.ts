import type { ClassValue } from 'clsx';
import type { AriaRole, ReactElement, ReactNode, Ref } from 'react';

export type FieldLayout = 'inline' | 'split';
export type FieldDensity = 'compact';
export type FieldTooltip = string | Element | ReactElement;

/** What every field renders around its control. */
export interface FieldChromeProps {
	id?: string;
	label?: string;
	labelTooltip?: FieldTooltip;
	description?: string | Element;
	layout?: FieldLayout;
	/** `compact` tightens this field and every field nested inside it. */
	density?: FieldDensity;
	className?: ClassValue;
	testId?: string;
}

export interface ControlledProps<V> {
	value: V;
	onChange: (next: V) => void;
	disabled?: boolean;
}

/** An icon already resolved by the caller; the kit never resolves one. */
export interface ResolvedIcon {
	iconUrl: string;
	name: string;
	href?: string;
}

export interface FieldClearProps {
	show: boolean;
	label?: string;
	className?: string;
	testId?: string;
	onClear: () => void;
}

export interface FieldProps extends FieldChromeProps {
	disabled?: boolean;
	/** Icon fields skip the stacking classes. */
	iconField?: boolean;
	/** `false` renders the root `flex-none` (SearchBar, ComboBox). */
	grow?: boolean;
	/** Renders the input wrapper with an absolute clear button; `show: false` keeps the wrapper and hides the button. */
	clear?: FieldClearProps;
	leading?: ReactNode;
	children?: ReactNode;
	after?: ReactNode;
	ref?: Ref<HTMLDivElement>;
	/** For a field that names a group of controls rather than one input; a `group` with a `label` and no `aria-labelledby` is named by that label. */
	role?: AriaRole;
	'aria-labelledby'?: string;
	/** Hides the field and keeps its children mounted. */
	hidden?: boolean;
}

export interface FieldRootProps {
	/** The control's id: `Field.Label` points at it and `Field.Control` hands it to a native control. */
	id?: string;
	layout?: FieldLayout;
	density?: FieldDensity;
	disabled?: boolean;
	/** `false` renders the root and a `Field.Clear` inside it `flex-none`. */
	grow?: boolean;
	className?: ClassValue;
	testId?: string;
	ref?: Ref<HTMLDivElement>;
	role?: AriaRole;
	'aria-labelledby'?: string;
	hidden?: boolean;
	children?: ReactNode;
}

export interface FieldLabelProps {
	/** `span` for a caption that names a group of controls, which a `<label>` cannot. */
	as?: 'label' | 'span';
	id?: string;
	tooltip?: FieldTooltip;
	className?: ClassValue;
	testId?: string;
	children: string;
}

export interface FieldDescriptionProps {
	id?: string;
	className?: ClassValue;
	testId?: string;
	children: ReactNode | Element;
}

export interface FieldControlProps {
	className?: ClassValue;
	testId?: string;
	children: ReactElement;
}

export interface FieldSlotProps {
	className?: ClassValue;
	testId?: string;
	children?: ReactNode;
}

export interface FieldClearPartProps extends FieldClearProps {
	children?: ReactNode;
}
