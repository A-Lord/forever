import { type ControlledProps, Field, type FieldChromeProps } from '@ui-kit/Field';
import { Select } from '@ui-kit/FormControl';
import clsx from 'clsx';

import type { EnumPickerOptions } from './types';

export type EnumPickerProps = FieldChromeProps & ControlledProps<number> & EnumPickerOptions;

export const EnumPicker = ({ value, onChange, disabled, values, ariaLabel, selectClassName, testId, ...chrome }: EnumPickerProps) => (
	<Field {...chrome} testId={testId ?? 'enum-picker-root'} disabled={disabled}>
		<Select
			id={chrome.id}
			className={clsx('w-auto max-w-full', selectClassName)}
			data-testid="enum-picker-selector"
			aria-label={ariaLabel}
			disabled={disabled}
			value={String(value)}
			onChange={event => onChange(Number(event.currentTarget.value))}>
			{values.map(entry => (
				<option key={entry.value} value={String(entry.value)} title={entry.tooltip}>
					{entry.name}
				</option>
			))}
		</Select>
	</Field>
);
