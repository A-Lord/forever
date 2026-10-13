import * as PresetUtils from '@app/preset_utils';
import { ConsumesSpec, HealingModel, Profession, Race, Spec } from '@generated/proto/common';
import {
	FeralBearDruid_Options as DruidOptions,
	FeralBearDruid_Rotation as DruidRotation,
	FeralBearDruid_Rotation_SwipeUsage as SwipeUsage,
} from '@generated/proto/druid';
import { OtherDefaults as SimUIOtherDefaults } from '@sim/spec_config';

export const DefaultSimpleRotation = DruidRotation.create({
	maintainFaerieFire: true,
	maintainDemoralizingRoar: true,
	maulRageThreshold: 50,
	swipeUsage: SwipeUsage.SwipeUsage_WithEnoughAP,
	swipeApThreshold: 2700,
});

import * as Consumables from '@features/settings/model/consumables_auto_gen';

import DefaultApl from './apls/default.apl.json';
export const ROTATION_SIMPLE = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecFeralBearDruid, DefaultSimpleRotation);
export const ROTATION_DEFAULT = PresetUtils.makePresetAPLRotation('APL', DefaultApl);

export const DefaultOptions = DruidOptions.create({
	startingRage: 0,
});

export const DefaultConsumables = ConsumesSpec.create({
	elixirIds: [Consumables.ElixirOfTheMongoose.id, Consumables.GiftOfArthas.id],
	foodId: Consumables.DirgesKickinChimaerokChops.id,
	defensivePotId: Consumables.GreaterStoneshieldPotion.id,
	conjuredId: Consumables.MajorHealthstone9421.id,
	mhImbueId: Consumables.DenseWeightstone.id,
	extraIds: [Consumables.GoblinSapperCharge.id, Consumables.ScrollOfStrengthIV.id, Consumables.ScrollOfProtectionIV.id],
});

export const OtherDefaults: Partial<SimUIOtherDefaults> = {
	profession1: Profession.Engineering,
	profession2: Profession.Enchanting,
	race: Race.RaceNightElf,
	distanceFromTarget: 0,
	reactionTime: 250,
	healingModel: HealingModel.create({
		hps: 2200,
		cadenceSeconds: 0.4,
		cadenceVariation: 1.2,
		absorbFrac: 0.02,
		burstWindow: 6,
		inspirationUptime: 0.25,
	}),
};
