import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
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
	elixirIds: [Consumables.ElixirOfTheMongoose.id, Consumables.ElixirOfGreaterFortitude.id],
	foodId: Consumables.SwiftstrikeSteak.id,
	potId: Consumables.MajorFrenzyPotion.id,
	conjuredId: Consumables.ThistleTea.id,
});

export const OtherDefaults = {
	distanceFromTarget: 5,
};
