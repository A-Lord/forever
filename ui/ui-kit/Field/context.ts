import { type AriaRole, createContext, useContext } from 'react';

import type { FieldDensity, FieldLayout } from './types';

export interface FieldContextValue {
	id: string | undefined;
	role: AriaRole | undefined;
	layout: FieldLayout | undefined;
	density: FieldDensity | undefined;
	disabled: boolean;
	grow: boolean;
	/** The ids of the root's descriptions, space-separated; `undefined` without one. */
	describedBy: string | undefined;
	registerDescription: (id: string) => () => void;
}

export const FieldContext = createContext<FieldContextValue | null>(null);

export const useField = () => {
	const value = useContext(FieldContext);
	if (value === null) throw new Error('A Field part must be rendered inside a Field.Root.');
	return value;
};
