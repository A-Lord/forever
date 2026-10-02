import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { ConsumesSpec, Profession, Race, Spec } from '@generated/proto/common';
import {
	FeralCatDruid_Options as FeralDruidOptions,
	FeralCatDruid_Rotation as FeralCatDruidRotation,
	FeralCatDruid_Rotation_FinishingMove as FinishingMove,
} from '@generated/proto/druid';

import DefaultApl from './apls/default.apl.json';
import Level60GearJson from './gear_sets/default.gear.json';

export const DefaultOptions = FeralDruidOptions.create({});

export const DefaultConsumables = ConsumesSpec.create({
	potId: Consumables.MajorFrenzyPotion.id,
	elixirIds: [Consumables.ElixirOfTheMongoose.id, Consumables.ElixirOfTheSages.id],
	foodId: Consumables.SwiftstrikeSteak.id,
	mhImbueId: Consumables.DenseWeightstone.id,
	conjuredId: Consumables.DemonicRune.id,
	extraIds: [Consumables.GoblinSapperCharge.id, Consumables.ScrollOfStrengthIV.id],
});

export const OtherDefaults = {
	distanceFromTarget: 0,
	profession1: Profession.Engineering,
	profession2: Profession.Enchanting,
	race: Race.RaceNightElf,
	reactionTime: 250,
};

export const DefaultRotation = FeralCatDruidRotation.create({
	finishingMove: FinishingMove.Rip,
	ripMinComboPoints: 5,
	biteMinComboPoints: 5,
	rakeTrick: true,
});

// Leather at 60 with no set bonus and no proc. The stand-in the rotation was tuned on.
export const Level60Gear = PresetUtils.makePresetGear('Level 60', Level60GearJson, {
	tooltip: 'Leather at level 60 with no set bonus and no proc. Fist of Cenarius and the Howling Idol.',
});

export const SIMPLE = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecFeralCatDruid, DefaultRotation);

export const APL = PresetUtils.makePresetAPLRotation('APL', DefaultApl);
