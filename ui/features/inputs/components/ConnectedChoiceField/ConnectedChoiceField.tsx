import { type ChoiceEntry, ChoiceField } from '@ui-kit/ChoiceField';

import { useConnectedField } from '../../hooks/useConnectedField';
import type { ConnectedChoiceFieldProps } from './types';

export const ConnectedChoiceField = <M, T, V = T>({ modObject, config, options, className, testId, ...presentation }: ConnectedChoiceFieldProps<M, T, V>) => {
	const { value, setValue, hidden, disabled, chrome } = useConnectedField<M, T, V>(modObject, config);
	if (hidden) return null;
	// A config that lists `values` is the number-valued `enum` kind, so its value is its `V`.
	const choices: ReadonlyArray<ChoiceEntry<V>> =
		options ?? (config.values ?? []).map(entry => ({ value: entry.value as V, label: entry.name, tooltip: entry.tooltip }));
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
