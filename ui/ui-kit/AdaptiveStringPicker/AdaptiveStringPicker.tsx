import { Field } from '@ui-kit/Field';
import { Input } from '@ui-kit/FormControl';
import { useCommitChange } from '@ui-kit/hooks/useCommitChange';
import { useDraft } from '@ui-kit/hooks/useDraft';
import { useCallback, useState } from 'react';

import type { AdaptiveStringPickerProps } from './types';

/** A text field that grows with its text, reporting the raw text on commit (blur after an edit, or Enter). */
export const AdaptiveStringPicker = ({ value, onChange, disabled, className, testId, ...chrome }: AdaptiveStringPickerProps) => {
	const [input, setInput] = useState<HTMLInputElement | null>(null);
	const attachInput = useCallback((element: HTMLElement | null) => setInput(element instanceof HTMLInputElement ? element : null), []);
	const { text, setDraft, clear } = useDraft(value);

	useCommitChange(input, committed => {
		onChange(committed.value);
		clear();
	});

	return (
		<Field {...chrome} className={['ui-adaptive-string-picker-root', className]} testId={testId ?? 'adaptive-string-picker-root'} disabled={disabled}>
			<Input
				type="text"
				ref={attachInput}
				id={chrome.id}
				disabled={disabled}
				value={text}
				onChange={event => setDraft(event.target.value)}
				size={Math.max(3, text.length)}
			/>
		</Field>
	);
};
