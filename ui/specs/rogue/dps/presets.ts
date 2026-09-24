import * as PresetUtils from '@app/preset_utils';
import { ConsumesSpec } from '@generated/proto/common';
import { Rogue_Options as RogueOptions } from '@generated/proto/rogue';

import SinisterAPL from './apls/default.apl.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const SINSITER_APL = PresetUtils.makePresetAPLRotation('Default', SinisterAPL);

export const DefaultOptions = RogueOptions.create({
	classOptions: {},
});

export const DefaultConsumables = ConsumesSpec.create({
	conjuredId: 7676,
});

export const OtherDefaults = {
	distanceFromTarget: 5,
};
