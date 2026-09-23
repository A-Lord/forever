import * as PresetUtils from '@app/preset_utils';
import { ConsumesSpec, HealingModel, Profession, Race } from '@generated/proto/common';
import { ProtectionWarrior_Options as ProtectionWarriorOptions, WarriorShout, WarriorStance } from '@generated/proto/warrior';
import { OtherDefaults as SimUIOtherDefaults } from '@sim/spec_config';

import * as WarriorPresets from '../shared/presets';
import GenericApl from './apls/default.apl.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const ROTATION_DEFAULT = PresetUtils.makePresetAPLRotation('Generic', GenericApl);

export const DefaultOptions = ProtectionWarriorOptions.create({
	classOptions: {
		startingRage: 100,
		defaultShout: WarriorShout.WarriorShoutBattle,
		defaultStance: WarriorStance.WarriorStanceDefensive,
		hasBsT2: true,
		stanceSnapshot: true,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	...WarriorPresets.DefaultConsumables,
	conjuredId: 9421,
	foodId: 27667,
	flaskId: undefined,
	battleElixirId: 22831,
	guardianElixirId: 9088,
	potId: 22849,
	extraIds: [10646, 10309, 10310, 10305],
});

export const OtherDefaults: Partial<SimUIOtherDefaults> = {
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	race: Race.RaceOrc,
	distanceFromTarget: 0,
	healingModel: HealingModel.create({
		hps: 2200,
		cadenceSeconds: 0.4,
		cadenceVariation: 1.2,
		absorbFrac: 0.02,
		burstWindow: 6,
		inspirationUptime: 0.25,
	}),
	// Morogrim
	// healingModel: HealingModel.create({
	// 	hps: 3300,
	// 	cadenceSeconds: 1.5,
	// 	cadenceVariation: 1.0,
	// 	absorbFrac: 0.02,
	// 	burstWindow: 6,
	// 	inspirationUptime: 0.12,
	// }),
};
