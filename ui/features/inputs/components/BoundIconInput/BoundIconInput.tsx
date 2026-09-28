import { BoundIconEnumPicker } from '../BoundIconEnumPicker';
import { BoundIconPicker } from '../BoundIconPicker';
import type { BoundIconInputConfig } from './types';

export interface BoundIconInputProps<ModObject> {
	modObject: ModObject;
	config: BoundIconInputConfig<ModObject, any>;
	testId?: string;
}

export const BoundIconInput = <ModObject,>({ modObject, config, testId }: BoundIconInputProps<ModObject>) =>
	config.type === 'icon' ? (
		<BoundIconPicker modObject={modObject} config={config} testId={testId} />
	) : (
		<BoundIconEnumPicker modObject={modObject} config={config} testId={testId} />
	);
