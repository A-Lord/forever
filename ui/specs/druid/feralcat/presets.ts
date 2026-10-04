import * as PresetUtils from '@app/preset_utils';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Profession, Race, Spec } from '@generated/proto/common';
import { FeralCatDruid_Options as FeralDruidOptions, FeralCatDruid_Rotation as FeralCatDruidRotation } from '@generated/proto/druid';
import { SavedTalents } from '@generated/proto/ui';

import DefaultApl from './apls/default.apl.json';
import RaidBuild from './builds/raid.build.json';
import StarterBuild from './builds/starter.build.json';
import RaidEp from './ep/raid.ep.json';

export const RAID_EP = PresetUtils.makePresetEpWeightsFromJSON(RaidEp);

export const STARTER_GEAR = PresetUtils.makePresetGear('Level 60 Starter', StarterBuild.player.equipment, {
	tooltip: 'Items from the game files database; no gems, enchants or random suffixes.',
});
export const STARTER_TALENTS = PresetUtils.makePresetTalents(
	'Level 60 Cat',
	SavedTalents.create({
		talentsString: StarterBuild.player.talentsString,
	}),
);

export const STARTER_BUILD = PresetUtils.makePresetBuildFromJSON('Level 60 Starter', Spec.SpecFeralCatDruid, StarterBuild, {
	gear: STARTER_GEAR,
	talents: STARTER_TALENTS,
});

export const RAID_GEAR = PresetUtils.makePresetGear('Level 60 Raid Example', RaidBuild.player.equipment, {
	tooltip: 'Representative gear from the game files, including AQ items. No launch-phase or BiS claim.',
});
export const RAID_BUILD = PresetUtils.makePresetBuildFromJSON('Level 60 Raid Example', Spec.SpecFeralCatDruid, RaidBuild, {
	gear: RAID_GEAR,
	talents: STARTER_TALENTS,
});

export const DefaultOptions = FeralDruidOptions.fromJson(StarterBuild.player.feralCatDruid.options);
export const DefaultConsumables = ConsumesSpec.fromJson(RaidBuild.player.consumables);
export const DefaultRaidBuffs = RaidBuffs.fromJson(RaidBuild.raidBuffs);
export const DefaultPartyBuffs = PartyBuffs.fromJson(RaidBuild.partyBuffs);
export const DefaultIndividualBuffs = IndividualBuffs.fromJson(RaidBuild.player.buffs);
export const DefaultDebuffs = Debuffs.fromJson(RaidBuild.debuffs);

export const OtherDefaults = {
	distanceFromTarget: 0,
	profession1: Profession.Engineering,
	profession2: Profession.Enchanting,
	race: Race.RaceNightElf,
	reactionTime: 250,
};

export const DefaultRotation = FeralCatDruidRotation.fromJsonString(JSON.stringify(RaidBuild.player.rotation.priorityList[1].action.catOptimalRotationAction));

export const SIMPLE = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecFeralCatDruid, DefaultRotation);
export const APL = PresetUtils.makePresetAPLRotation('Claw / Rip', DefaultApl);
export const POWERSHIFT_APL = PresetUtils.makePresetAPLRotation('Shred / Rip', RaidBuild.player.rotation);
