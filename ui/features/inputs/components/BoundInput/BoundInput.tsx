import { usePlayer } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';

import { BoundBooleanPicker } from '../BoundBooleanPicker';
import { BoundEnumPicker } from '../BoundEnumPicker';
import { BoundNumberPicker } from '../BoundNumberPicker';
import type { BoundInputConfig } from './types';

export interface BoundInputProps {
	config: BoundInputConfig<Player<any>>;
	/**
	 * Lay the label and the control on one line. The rotation tab's simple-rotation section sets it
	 * for every input; the settings blocks do not.
	 */
	inline?: boolean;
}

export const BoundInput = ({ config, inline }: BoundInputProps) => {
	const player = usePlayer();
	const shared = inline ? { ...config, layout: 'inline' as const } : config;

	switch (shared.type) {
		case 'number':
			return <BoundNumberPicker modObject={player} config={shared} />;
		case 'boolean':
			return <BoundBooleanPicker modObject={player} config={{ ...shared, reverse: true }} />;
		case 'enum':
			return <BoundEnumPicker modObject={player} config={shared} />;
	}
};
