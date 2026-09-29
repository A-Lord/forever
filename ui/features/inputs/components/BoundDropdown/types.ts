import type { DropdownPickerProps } from '@ui-kit/DropdownPicker';

import type { InputConfig } from '../../input';

export interface BoundDropdownProps<M, T, V = T> extends Pick<
	DropdownPickerProps<V>,
	'options' | 'defaultLabel' | 'hideLabelWhenDefault' | 'triggerClassName' | 'testId'
> {
	modObject: M;
	/**
	 * `V` differs from `T` when the option type is not the stored type, bridged by the config's own
	 * `sourceToValue`/`valueToSource`. The APL action-id field stores an `ActionID` message and offers
	 * `ActionId` objects; its unit field stores a `UnitReference` and offers a `UnitValue` carrying
	 * that unit's icon and colour.
	 */
	config: InputConfig<M, T, V> & { id: string };
	/** Two option values are the same selection. Defaults to `===`, which every enum caller wants. */
	equals?: DropdownPickerProps<V>['equals'];
}
