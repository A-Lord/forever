import type { ClassValue } from 'clsx';
import type { AriaRole, ReactElement, ReactNode, Ref } from 'react';

export type FieldLayout = 'inline' | 'split';

/** What every field renders around its control. */
export interface FieldChromeProps {
	id?: string;
	label?: string;
	labelTooltip?: string | Element | ReactElement;
	description?: string | Element;
	layout?: FieldLayout;
	/** `compact` tightens this field and every field nested inside it. */
	density?: 'compact';
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
