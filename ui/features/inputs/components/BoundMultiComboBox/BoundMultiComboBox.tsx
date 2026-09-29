import { MultiComboBox } from '@ui-kit/MultiComboBox';

import { useBoundField } from '../../hooks/useBoundField';
import type { MultiComboBoxConfig } from './types';

export interface BoundMultiComboBoxProps<M> {
	modObject: M;
	config: MultiComboBoxConfig<M>;
	className?: string;
	testId?: string;
}

export const BoundMultiComboBox = <M,>({ modObject, config, className, testId }: BoundMultiComboBoxProps<M>) => {
	const {
		value,
		setValue,
		hidden,
		disabled,
		chrome: { className: chromeClassName, ...chrome },
	} = useBoundField(modObject, config);
	if (hidden) return null;
	return (
		<MultiComboBox
			{...chrome}
			className={[chromeClassName, className]}
			testId={testId}
			value={value}
			onChange={setValue}
			disabled={disabled}
			values={config.values}
			placeholder={config.placeholder}
			emptyText={config.emptyText}
			removeLabel={config.removeLabel}
			openLabel={config.openLabel}
		/>
	);
};
