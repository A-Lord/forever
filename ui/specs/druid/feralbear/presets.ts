import * as PresetUtils from '@app/preset_utils';
import { ConsumesSpec, HealingModel, Profession, Race, Spec } from '@generated/proto/common';
import { FeralBearDruid_Options as DruidOptions, FeralBearDruid_Rotation as DruidRotation } from '@generated/proto/druid';
import { SavedTalents } from '@generated/proto/ui';
import type { OtherDefaults as SimUIOtherDefaults } from '@sim/spec_config';

import StarterBuild from './builds/starter.build.json';

export const STARTER_GEAR = PresetUtils.makePresetGear('Level 60 Starter', StarterBuild.player.equipment, {
	tooltip: 'Items from the game files database; no gems, enchants or random suffixes.',
});
export const STARTER_TALENTS = PresetUtils.makePresetTalents(
	'Level 60 Bear',
	SavedTalents.create({
		talentsString: StarterBuild.player.talentsString,
	}),
);

export const DefaultSimpleRotation = DruidRotation.fromJsonString(
	JSON.stringify(StarterBuild.player.rotation.priorityList[0].action.bearOptimalRotationAction),
);

export const ROTATION_SIMPLE = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecFeralBearDruid, DefaultSimpleRotation);
export const ROTATION_DEFAULT = PresetUtils.makePresetAPLRotation('Tank', StarterBuild.player.rotation);
export const STARTER_BUILD = PresetUtils.makePresetBuildFromJSON('Level 60 Starter Tank', Spec.SpecFeralBearDruid, StarterBuild, {
	gear: STARTER_GEAR,
	talents: STARTER_TALENTS,
});

export const DefaultOptions = DruidOptions.fromJson(StarterBuild.player.feralBearDruid.options);
export const DefaultConsumables = ConsumesSpec.fromJson(StarterBuild.player.consumables);

export const OtherDefaults: Partial<SimUIOtherDefaults> = {
	profession1: Profession.Engineering,
	profession2: Profession.Enchanting,
	race: Race.RaceNightElf,
	distanceFromTarget: 0,
	reactionTime: 250,
	healingModel: HealingModel.fromJson(StarterBuild.player.healingModel),
};
