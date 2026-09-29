import { type ChoiceEntry, ChoiceField } from '@ui-kit/ChoiceField';
import type { EnumValueConfig } from '@ui-kit/MultiComboBox';
import { useMemo } from 'react';

import { useConnectedField } from '../../hooks/useConnectedField';
import type { ConnectedChoiceFieldProps } from './types';

const NO_VALUES: ReadonlyArray<EnumValueConfig> = [];

export const ConnectedChoiceField = <M, T, V = T>({ modObject, config, options, className, testId, ...presentation }: ConnectedChoiceFieldProps<M, T, V>) => {
	const { value, setValue, hidden, disabled, chrome } = useConnectedField<M, T, V>(modObject, config);
	const values = config.values ?? NO_VALUES;
	// A config that lists `values` is the number-valued `enum` kind, so its value is its `V`.
	const choices = useMemo(
		(): ReadonlyArray<ChoiceEntry<V>> => options ?? values.map(entry => ({ value: entry.value as V, label: entry.name, tooltip: entry.tooltip })),
		[options, values],
	);
	if (hidden) return null;
	return (
		<ChoiceField<V>
			{...chrome}
			{...presentation}
			className={[chrome.className, className]}
			testId={testId}
			value={value}
			onChange={setValue}
			disabled={disabled}
			options={choices}
		/>
	);
};
