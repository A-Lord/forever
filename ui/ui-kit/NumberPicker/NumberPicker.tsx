import { formatToNumber } from '@sim/utils/format';
import { type ControlledProps, Field, type FieldChromeProps } from '@ui-kit/Field';
import { Input } from '@ui-kit/FormControl';
import { useCommitChange } from '@ui-kit/hooks/useCommitChange';
import { useDraft } from '@ui-kit/hooks/useDraft';
import clsx from 'clsx';
import { useCallback, useState } from 'react';

import type { NumberPickerOptions } from './types';

export type NumberPickerProps = FieldChromeProps & ControlledProps<number> & NumberPickerOptions;

const formatSourceValue = (value: number, float: boolean, showZeroes: boolean, maxDecimalDigits: number): string => {
	if (value === 0 && !showZeroes) return '';
	if (float) return formatToNumber(value, { useGrouping: false, minimumFractionDigits: 2, maximumFractionDigits: maxDecimalDigits });
	return String(value);
};

const applyPositive = (text: string, float: boolean, maxDecimalDigits: number): string => {
	if (float) return formatToNumber(Math.abs(Number(text)), { useGrouping: false, minimumFractionDigits: 2, maximumFractionDigits: maxDecimalDigits });
	return Math.abs(parseInt(text)).toString();
};

const parseValue = (text: string, float: boolean): number => {
	return float ? Number(text || '') || 0 : parseInt(text || '') || 0;
};

export const NumberPicker = ({
	value,
	onChange,
	disabled,
	float = false,
	positive = false,
	showZeroes = true,
	maxDecimalDigits = 2,
	inputClassName,
	className,
	testId,
	...chrome
}: NumberPickerProps) => {
	const [input, setInput] = useState<HTMLInputElement | null>(null);
	const attachInput = useCallback((element: HTMLElement | null) => setInput(element instanceof HTMLInputElement ? element : null), []);

	const { text, setDraft, clear } = useDraft(formatSourceValue(value, float, showZeroes, maxDecimalDigits));

	useCommitChange(input, committed => {
		const entered = positive ? applyPositive(committed.value, float, maxDecimalDigits) : committed.value;
		onChange(parseValue(entered, float));
		clear();
	});

	return (
		<Field {...chrome} className={['ui-number-picker-root', className]} testId={testId ?? 'number-picker-root'} disabled={disabled}>
			<Input
				type="text"
				ref={attachInput}
				id={chrome.id}
				className={clsx('ui-number-picker-input', inputClassName)}
				data-testid="number-picker-input"
				disabled={disabled}
				value={text}
				size={Math.max(3, text.length)}
				onChange={event => {
					// React also reports the native `change` event here, which is the commit above; only a keystroke is a draft.
					if (event.nativeEvent.type === 'input') setDraft(event.currentTarget.value);
				}}
			/>
		</Field>
	);
};
