import type { FieldChromeProps } from '@ui-kit/Field/types';

import type { InputConfig } from '../input';
import { useInput } from './useInput';

export const useBoundField = <M, T, V = T>(modObject: M, config: InputConfig<M, T, V>) => {
	const state = useInput(modObject, config);
	const chrome: FieldChromeProps = {
		id: config.id,
		label: config.label,
		labelTooltip: config.labelTooltip,
		description: config.description,
		layout: config.layout,
		className: config.extraClassNames,
	};
	return { ...state, chrome };
};
