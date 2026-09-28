import { Field } from '@ui-kit/Field';
import { Input } from '@ui-kit/FormControl';
import { useCommitChange } from '@ui-kit/hooks/useCommitChange';
import { useDraft } from '@ui-kit/hooks/useDraft';
import { useCallback, useState } from 'react';

import type { NumberListPickerProps } from './types';

const parseInputValue = (text: string): Array<number> => {
	if (!text) return [];
	return text
		.split(',')
		.map(parseFloat)
		.filter(val => !isNaN(val));
};

const formatValue = (value: Array<number>): string => value.map(v => String(v)).join(',');

export const NumberListPicker = ({ value, onChange, disabled, placeholder, testId, ...chrome }: NumberListPickerProps) => {
	const [input, setInput] = useState<HTMLInputElement | null>(null);
	const attachInput = useCallback((element: HTMLElement | null) => setInput(element instanceof HTMLInputElement ? element : null), []);
	const { text, setDraft, clear } = useDraft(formatValue(value));

	useCommitChange(input, committed => {
		onChange(parseInputValue(committed.value));
		clear();
	});

	return (
		<Field {...chrome} testId={testId ?? 'number-list-picker-root'} disabled={disabled}>
			<Input
				type="text"
				ref={attachInput}
				id={chrome.id}
				className="w-full"
				data-testid="number-list-picker-input"
				placeholder={placeholder || ''}
				disabled={disabled}
				value={text}
				onChange={event => setDraft(event.target.value)}
			/>
		</Field>
	);
};
