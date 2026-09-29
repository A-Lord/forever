import { useContext } from 'react';

import { FieldContext } from '../context/FieldContext';

export const useField = () => {
	const value = useContext(FieldContext);
	if (value === null) throw new Error('A Field part must be rendered inside a Field.Root.');
	return value;
};
