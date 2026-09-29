import { DropdownPicker } from '@ui-kit/DropdownPicker';

import { useBoundField } from '../../hooks/useBoundField';
import type { BoundDropdownProps } from './types';

const identical = <V,>(a: V | undefined, b: V | undefined) => a === b;

export const BoundDropdown = <M, T, V = T>({
	modObject,
	config,
	options,
	equals = identical,
	defaultLabel,
	hideLabelWhenDefault,
	triggerClassName,
	testId,
}: BoundDropdownProps<M, T, V>) => {
	const { value, setValue, hidden, disabled, chrome } = useBoundField<M, T, V>(modObject, config);
	if (hidden) return null;
	return (
		<DropdownPicker<V>
			{...chrome}
			testId={testId}
			value={value}
			onChange={setValue}
			disabled={disabled}
			options={options}
			equals={equals}
			defaultLabel={defaultLabel}
			hideLabelWhenDefault={hideLabelWhenDefault}
			triggerClassName={triggerClassName}
		/>
	);
};
