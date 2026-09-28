import { Field } from '@ui-kit/Field';

import { DropdownMenu } from './DropdownMenu';
import type { DropdownPickerProps } from './types';

/**
 * A value picker with rich options — an icon, a label, a class list, a tooltip and a category path
 * per entry.
 *
 * It parameterises the option list, how two values compare and what each option renders; it fixes
 * the trigger-and-menu markup and that picking an option closes the menu. It is **not** an
 * `InputConfig` picker: `value`/`onChange` are the whole binding, so a caller whose selection is
 * UI-local state needs no store to write it through. The label, description and disabled state
 * come from its `Field` root.
 *
 * The APL action-id pickers replace their list on every unit-metadata change, which is a prop
 * update here rather than a rebuild-and-compare.
 */
export const DropdownPicker = <V,>({ id, label, labelTooltip, description, layout, className, testId, disabled, ...menu }: DropdownPickerProps<V>) => (
	<Field
		id={id}
		label={label}
		labelTooltip={labelTooltip}
		description={description}
		layout={layout}
		className={['relative', className]}
		testId={testId ?? 'dropdown-picker-root'}
		disabled={disabled}>
		<DropdownMenu<V> {...menu} id={id} />
	</Field>
);
