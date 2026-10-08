import * as PresetUtils from '@app/preset_utils';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Profession, Race, Spec } from '@generated/proto/common';
import { FeralCatDruid_Options as FeralDruidOptions, FeralCatDruid_Rotation as FeralCatDruidRotation } from '@generated/proto/druid';
import { SavedTalents } from '@generated/proto/ui';

import AutoApl from './apls/auto.apl.json';
import DefaultApl from './apls/default.apl.json';
import PounceApl from './apls/pounce.apl.json';
import RaidDpsApl from './apls/raid-dps.apl.json';
import RavageApl from './apls/ravage.apl.json';
import RaidBuild from './builds/raid.build.json';
import RaidDpsBuild from './builds/raid-dps.build.json';
import StarterBuild from './builds/starter.build.json';
import RaidEp from './ep/raid.ep.json';
import Level60GearJson from './gear_sets/default.gear.json';
import PreRaidGearJson from './gear_sets/pre_raid.gear.json';

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
export const RAID_DPS_BUILD = PresetUtils.makePresetBuildFromJSON('Level 60 Raid DPS Example', Spec.SpecFeralCatDruid, RaidDpsBuild);

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

// Leather at 60 with no set bonus and no proc. The stand-in the rotation was tuned on.
export const Level60Gear = PresetUtils.makePresetGear('Level 60', Level60GearJson, {
	tooltip: 'Leather at level 60 with no set bonus and no proc. Fist of Cenarius and the Howling Idol.',
});

export const PreRaidGear = PresetUtils.makePresetGear('Pre-raid', PreRaidGearJson, {
	tooltip: 'Pre-raid leather with The Unstoppable Force, Wolfshead Helm and full enchants.',
});

export const SIMPLE = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecFeralCatDruid, DefaultRotation);
export const APL = PresetUtils.makePresetAPLRotation('Claw / Rip', DefaultApl);
export const AUTO_APL = PresetUtils.makePresetAPLRotation('Auto', AutoApl);
export const POWERSHIFT_APL = PresetUtils.makePresetAPLRotation('Shred / Rake / Rip (4 CP)', RaidBuild.player.rotation);
export const RAID_DPS_APL = PresetUtils.makePresetAPLRotation('Shred / Rake / Rip (5 CP)', RaidDpsApl);
export const RAVAGE_APL = PresetUtils.makePresetAPLRotation('Ravage opener', RavageApl);
export const POUNCE_APL = PresetUtils.makePresetAPLRotation('Pounce opener', PounceApl);
