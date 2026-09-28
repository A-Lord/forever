import { useAplInput } from '@features/apl/hooks/useAplInput';
import type { InputConfig } from '@features/inputs/input';
import type { Player } from '@sim/player/player';
import { Field } from '@ui-kit/Field';
import { ListItemAction } from '@ui-kit/ListPicker';

export interface HidePickerProps {
	player: Player<any>;
	config: InputConfig<Player<any>, boolean>;
}

/**
 * The eye that disables one APL list item.
 *
 * It deliberately has no tooltip: the "Enable/Disable" string it would want was never translated,
 * so adding it would put one English label into an otherwise localised header.
 */
export const HidePicker = ({ player, config }: HidePickerProps) => {
	const { value, setValue, hidden, disabled, chrome } = useAplInput(player, config);

	if (hidden) return null;
	return (
		<Field {...chrome} className={['m-0 w-auto', chrome.className]} testId="hide-picker-root" disabled={disabled}>
			<ListItemAction icon={value ? 'eye-slash' : 'eye'} testId="hide-picker-button" onClick={() => setValue(!value)} />
		</Field>
	);
};
