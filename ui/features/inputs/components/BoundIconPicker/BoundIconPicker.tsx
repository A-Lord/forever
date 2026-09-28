import { useResolvedIcon } from '@sim/hooks/useResolvedIcon';
import { IconPicker } from '@ui-kit/IconPicker';

import { useBoundField } from '../../hooks/useBoundField';
import { useHideZeroRestore } from '../../hooks/useHideZeroRestore';
import type { IconPickerConfig } from './types';

export interface BoundIconPickerProps<ModObject, ValueType> {
	modObject: ModObject;
	config: IconPickerConfig<ModObject, ValueType>;
	testId?: string;
}

export const BoundIconPicker = <ModObject, ValueType>({ modObject, config, testId }: BoundIconPickerProps<ModObject, ValueType>) => {
	const { value, setValue, hidden: showWhenHidden, disabled, revision, chrome } = useBoundField(modObject, config);
	const hidden = config.actionId == null || showWhenHidden;

	// A two-state source stores a boolean. The hide/restore writes go through `write` as well, so a restored number is stored as a boolean too.
	const toSource = (n: number) => (config.states === 2 ? Boolean(n) : n) as ValueType;
	const write = (next: ValueType) => setValue(toSource(Number(next)));
	useHideZeroRestore({ value, setValue: write, hidden, revision }, toSource(0));

	const icon = useResolvedIcon(config.actionId ?? undefined);
	const improvedIcon = useResolvedIcon(config.improvedId);
	const improvedIcon2 = useResolvedIcon(config.improvedId2);

	if (hidden) return null;

	return (
		<IconPicker
			{...chrome}
			testId={testId}
			states={config.states}
			icon={icon}
			improvedIcon={improvedIcon}
			improvedIcon2={improvedIcon2}
			value={Number(value)}
			onChange={n => setValue(toSource(n))}
			disabled={disabled}
		/>
	);
};
