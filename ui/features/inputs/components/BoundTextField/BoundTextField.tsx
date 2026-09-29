import { NumberField, NumberListField, StringField } from '@ui-kit/TextField';
import clsx from 'clsx';

import { useBoundField } from '../../hooks/useBoundField';
import type { NumberFieldConfig, NumberListFieldConfig, StringFieldConfig } from './types';

export interface BoundNumberFieldProps<ModObject> {
	modObject: ModObject;
	config: NumberFieldConfig<ModObject>;
	inputClassName?: string;
	testId?: string;
}

export const BoundNumberField = <ModObject,>({ modObject, config, inputClassName, testId }: BoundNumberFieldProps<ModObject>) => {
	const { value, setValue, hidden, disabled, chrome } = useBoundField(modObject, config);
	if (hidden) return null;
	return (
		<NumberField
			{...chrome}
			testId={testId}
			value={value}
			onChange={setValue}
			disabled={disabled}
			float={config.float}
			positive={config.positive}
			showZeroes={config.showZeroes}
			maxDecimalDigits={config.maxDecimalDigits}
			inputClassName={clsx(config.inputClassName, inputClassName)}
		/>
	);
};

export interface BoundStringFieldProps<M> {
	modObject: M;
	config: StringFieldConfig<M>;
	testId?: string;
}

export const BoundStringField = <M,>({ modObject, config, testId }: BoundStringFieldProps<M>) => {
	const { value, setValue, hidden, disabled, chrome } = useBoundField(modObject, config);
	if (hidden) return null;
	return <StringField {...chrome} testId={testId} value={value} onChange={setValue} disabled={disabled} />;
};

export interface BoundNumberListFieldProps<M> {
	modObject: M;
	config: NumberListFieldConfig<M>;
	testId?: string;
}

export const BoundNumberListField = <M,>({ modObject, config, testId }: BoundNumberListFieldProps<M>) => {
	const { value, setValue, hidden, disabled, chrome } = useBoundField(modObject, config);
	if (hidden) return null;
	return <NumberListField {...chrome} testId={testId} value={value} onChange={setValue} disabled={disabled} placeholder={config.placeholder} />;
};
