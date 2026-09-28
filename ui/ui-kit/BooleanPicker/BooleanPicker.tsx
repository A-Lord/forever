import { Checkbox } from '@base-ui/react/checkbox';
import { type ControlledProps, Field, type FieldChromeProps } from '@ui-kit/Field';
import { Icon } from '@ui-kit/Icon';

import type { BooleanPickerOptions } from './types';

export type BooleanPickerProps = FieldChromeProps & ControlledProps<boolean> & BooleanPickerOptions;

export const BooleanPicker = ({ value, onChange, disabled, reverse, className, testId, ...chrome }: BooleanPickerProps) => {
	const input = (
		<Checkbox.Root
			id={chrome.id}
			className="ui-boolean-picker-input inline-flex items-center justify-center p-0"
			data-testid="boolean-picker-input"
			checked={value}
			disabled={disabled}
			onCheckedChange={checked => onChange(checked)}>
			<Checkbox.Indicator>
				<Icon name="check" className="text-primary-foreground" />
			</Checkbox.Indicator>
		</Checkbox.Root>
	);

	return (
		<Field
			{...chrome}
			className={['ui-boolean-picker', reverse && 'ui-boolean-picker-reverse', className]}
			testId={testId ?? 'boolean-picker-root'}
			disabled={disabled}
			leading={reverse ? undefined : input}>
			{reverse ? input : undefined}
		</Field>
	);
};
