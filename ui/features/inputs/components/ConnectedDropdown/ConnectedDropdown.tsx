import { DropdownPicker } from '@ui-kit/DropdownPicker';

import { useConnectedField } from '../../hooks/useConnectedField';
import type { ConnectedDropdownProps } from './types';

const identical = <V,>(a: V | undefined, b: V | undefined) => a === b;

export const ConnectedDropdown = <M, T, V = T>({
	modObject,
	config,
	options,
	equals = identical,
	defaultLabel,
	hideLabelWhenDefault,
	triggerClassName,
	testId,
}: ConnectedDropdownProps<M, T, V>) => {
	const { value, setValue, hidden, disabled, chrome } = useConnectedField<M, T, V>(modObject, config);
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
