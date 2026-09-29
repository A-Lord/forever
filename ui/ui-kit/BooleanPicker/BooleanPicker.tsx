import { Checkbox } from '@base-ui/react/checkbox';
import { type ControlledProps, Field, type FieldChromeProps } from '@ui-kit/Field';
import { Icon } from '@ui-kit/Icon';
import clsx from 'clsx';

import type { BooleanPickerOptions } from './types';

export type BooleanPickerProps = FieldChromeProps & ControlledProps<boolean> & BooleanPickerOptions;

export const BooleanPicker = ({ value, onChange, disabled, reverse, className, testId, ...chrome }: BooleanPickerProps) => {
	const hasChrome = Boolean(chrome.label || chrome.description || chrome.layout);

	const input = (
		<Checkbox.Root
			id={chrome.id}
			className={clsx('ui-boolean-picker-input inline-flex items-center justify-center p-0', !hasChrome && className)}
			data-testid={hasChrome ? 'boolean-picker-input' : (testId ?? 'boolean-picker-input')}
			checked={value}
			disabled={disabled}
			onCheckedChange={checked => onChange(checked)}>
			<Checkbox.Indicator>
				<Icon name="check" className="text-primary-foreground" />
			</Checkbox.Indicator>
		</Checkbox.Root>
	);

	if (!hasChrome) return input;

	return (
		<Field
			{...chrome}
			className={['ui-boolean-picker', className]}
			testId={testId ?? 'boolean-picker-root'}
			disabled={disabled}
			reverse={reverse}
			leading={reverse ? undefined : input}>
			{reverse ? input : undefined}
		</Field>
	);
};
