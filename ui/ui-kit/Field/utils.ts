import type { FieldChromeProps } from './types';

/** A field renders a `Field` root only for chrome: a label, a description, a layout or a clear slot. Without any, the control renders bare. */
export const hasFieldChrome = ({ label, description, layout }: Pick<FieldChromeProps, 'label' | 'description' | 'layout'>, clear?: object): boolean =>
	!!(label || description || layout || clear);
