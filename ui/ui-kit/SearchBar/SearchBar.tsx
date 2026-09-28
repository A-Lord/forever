import { Field } from '@ui-kit/Field';
import { Input } from '@ui-kit/FormControl';
import { type ReactNode, useEffect, useState } from 'react';
import { useDebounce } from 'react-use';

export interface SearchBarProps {
	value: string;
	onChange: (value: string) => void;
	placeholder?: string;
	label?: string;
	id?: string;
	debounceMs?: number;
	clearable?: boolean;
	clearLabel?: string;
	/** Lands on the clear button, for a caller whose stylesheet already names it. */
	clearClassName?: string;
	autoFocus?: boolean;
	autoComplete?: 'on' | 'off';
	className?: string;
	grow?: boolean;
	/** Rendered after the input group, inside the field — for a results list positioned against the same box. */
	children?: ReactNode;
	inputTestId?: string;
}

export const SearchBar = ({
	value,
	onChange,
	placeholder,
	label,
	id,
	debounceMs = 0,
	clearable = false,
	clearLabel,
	clearClassName,
	autoFocus,
	autoComplete,
	className,
	grow = true,
	children,
	inputTestId,
}: SearchBarProps) => {
	const [draft, setDraft] = useState(value);
	useEffect(() => setDraft(value), [value]);

	useDebounce(
		() => {
			if (debounceMs > 0) onChange(draft);
		},
		debounceMs,
		[draft],
	);

	const handleInput = (next: string) => {
		setDraft(next);
		if (debounceMs <= 0) onChange(next);
	};

	const handleClear = () => {
		setDraft('');
		onChange('');
	};

	return (
		<Field
			label={label}
			id={id}
			grow={grow}
			testId="search-bar-root"
			clear={{
				show: clearable && draft.length > 0,
				label: clearLabel,
				className: clearClassName,
				testId: 'search-bar-clear-btn',
				onClear: handleClear,
			}}
			after={children}>
			<Input
				id={id}
				type="text"
				data-testid={inputTestId}
				className={className}
				placeholder={placeholder}
				autoFocus={autoFocus}
				autoComplete={autoComplete}
				value={draft}
				onChange={event => handleInput(event.target.value)}
			/>
		</Field>
	);
};
