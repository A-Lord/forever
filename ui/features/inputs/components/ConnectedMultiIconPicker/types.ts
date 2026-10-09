import type { Player } from '@sim/player/player';
import type { ActionId } from '@sim/proto/action_id';

import type { IconPickerConfig } from '../ConnectedIconPicker/types';

export type MultiIconPickerItemConfig<ModObject> = IconPickerConfig<ModObject, any>;

export interface MultiIconPickerConfig<ModObject> {
	inputs: Array<MultiIconPickerItemConfig<ModObject>>;
	label?: string;
	categoryId?: ActionId;
	showWhen?: (obj: Player<any>) => boolean;
	// Dims the trigger with the kept tooltip while it holds: a selection the stat filter hides, still offered because it is selected.
	keptBy?: (obj: ModObject) => boolean;
}
