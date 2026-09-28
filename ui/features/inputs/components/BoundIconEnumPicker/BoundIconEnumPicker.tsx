import { useActionIds } from '@sim/hooks/useActionIds';
import { useResolvedIcon } from '@sim/hooks/useResolvedIcon';
import { IconEnumPicker } from '@ui-kit/IconEnumPicker';
import { useEffect, useRef } from 'react';

import { useBoundField } from '../../hooks/useBoundField';
import type { IconEnumPickerConfig, IconEnumValueConfig } from './types';
import { iconEnumPickerShown } from './utils';

export interface BoundIconEnumPickerProps<ModObject, T> {
	modObject: ModObject;
	config: IconEnumPickerConfig<ModObject, T>;
	testId?: string;
}

export const BoundIconEnumPicker = <ModObject, T>({ modObject, config, testId }: BoundIconEnumPickerProps<ModObject, T>) => {
	const { value, setValue, disabled, revision, chrome } = useBoundField(modObject, config);
	const shows = (valueConfig: IconEnumValueConfig<ModObject, T>) => !valueConfig.showWhen || valueConfig.showWhen(modObject);

	// Keyed on which values carry an actionId, not on which icons have resolved: a notification landing before the icons arrive must not read as hidden and zero the source.
	const hidden = !iconEnumPickerShown(config, modObject);
	// One effect, not `useHideZeroRestore` plus a second one: the hidden-option rule has to see the value the hide/restore step just wrote.
	const stored = useRef<T | undefined>(undefined);
	const lastRevision = useRef<number | null>(null);
	useEffect(() => {
		if (lastRevision.current === revision) return;
		const construction = lastRevision.current === null;
		lastRevision.current = revision;
		if (construction) return;

		let current = value;
		if (hidden) {
			if (stored.current === undefined) {
				stored.current = value;
				current = config.zeroValue;
				setValue(current);
			}
		} else if (stored.current !== undefined) {
			const restored = stored.current;
			stored.current = undefined;
			if (config.equals(value, config.zeroValue)) {
				current = restored;
				setValue(current);
			}
		}

		if (config.values.some(valueConfig => !shows(valueConfig) && valueConfig.value === current)) setValue(config.zeroValue);
	});

	const icons = useActionIds(config.values.map(valueConfig => valueConfig.actionId));
	const selected = config.values.find(valueConfig => config.equals(valueConfig.value, value));
	const backupIcon = useResolvedIcon(selected ? undefined : config.backupIconUrl?.(value));

	if (hidden) return null;

	return (
		<IconEnumPicker
			{...chrome}
			testId={testId}
			options={config.values.map((valueConfig, index) => ({
				value: valueConfig.value,
				icon: icons[index],
				color: valueConfig.color,
				iconUrl: valueConfig.iconUrl,
				text: valueConfig.text,
				tooltip: valueConfig.tooltip,
				hidden: !shows(valueConfig),
			}))}
			equals={config.equals}
			zeroValue={config.zeroValue}
			backupIcon={backupIcon}
			tooltip={config.tooltip}
			direction={config.direction}
			numColumns={typeof config.numColumns === 'function' ? config.numColumns(modObject) : config.numColumns}
			value={value}
			onChange={next => {
				stored.current = undefined;
				setValue(next);
			}}
			disabled={disabled}
		/>
	);
};
