import { BoundBooleanPicker } from '@features/inputs/components/BoundBooleanPicker';
import { BoundNumberPicker } from '@features/inputs/components/BoundNumberPicker';
import { usePlayer } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import type { InputConfig } from '@sim/spec_config';
import { EnumPicker } from '@ui-kit/EnumPicker';

export interface InputPickerProps {
	config: InputConfig<Player<any>>;
	/**
	 * Lay the label and the control on one line. The rotation tab's simple-rotation section sets it
	 * for every input; the settings blocks do not.
	 */
	inline?: boolean;
}

export const InputPicker = ({ config, inline }: InputPickerProps) => {
	const player = usePlayer();
	const shared = inline ? { ...config, layout: 'inline' as const } : config;

	switch (shared.type) {
		case 'number':
			return <BoundNumberPicker modObject={player} config={shared} />;
		case 'boolean':
			return <BoundBooleanPicker modObject={player} config={{ ...shared, reverse: true }} />;
		case 'enum':
			return <EnumPicker modObject={player} config={shared} />;
	}
};
