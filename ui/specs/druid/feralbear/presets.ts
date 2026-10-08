import * as PresetUtils from '@app/preset_utils';
import { ConsumesSpec, HealingModel, Profession, PseudoStat, Race, Spec, Stat } from '@generated/proto/common';
import {
	FeralBearDruid_Options as DruidOptions,
	FeralBearDruid_Rotation as DruidRotation,
	FeralBearDruid_Rotation_SwipeUsage as SwipeUsage,
} from '@generated/proto/druid';
import { Stats } from '@sim/proto/stats';
import { Encounter } from '@sim/raid/encounter';
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
import Level60GearJson from './gear_sets/default.gear.json';
export const ROTATION_SIMPLE = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecFeralBearDruid, DefaultSimpleRotation);
export const ROTATION_DEFAULT = PresetUtils.makePresetAPLRotation('APL', DefaultApl);

// Leather at 60 with no set bonus and no proc, about the item level of the cat preset.
export const Level60Gear = PresetUtils.makePresetGear('Level 60', Level60GearJson, {
	tooltip: 'Leather at level 60 with no set bonus and no proc. The idol slot is empty: every idol has an effect.',
});

// The default raid target swings for 15113 before armor, which kills a level 60 bear on every pull.
const level60Encounter = Encounter.defaultEncounterProto();
level60Encounter.targets[0].minBaseDamage = 6000;
export const Level60Encounter = {
	...PresetUtils.makePresetEncounter('Level 60 boss', level60Encounter),
	tooltip: 'The raid target, hitting for 6000 before armor instead of 15113.',
};

// EP on the Level 60 preset, the APL preset and the Level 60 boss, 100000 iterations.
// DPS has agility as 1. A bear's agility is crit, armor and dodge but no attack power, so strength
// is worth more than one agility. Dodge and defense cost damage: fewer hits taken is less rage.
export const Level60DpsEp = Stats.fromMap(
	{
		[Stat.StatStrength]: 2.0751,
		[Stat.StatAgility]: 1,
		[Stat.StatStamina]: 0.1044,
		[Stat.StatAttackPower]: 0.9432,
		[Stat.StatFeralAttackPower]: 0.9432,
		[Stat.StatArmor]: 0.0276,
		[Stat.StatBonusArmor]: 0.0066,
		[Stat.StatDodgeRating]: -0.1384,
		[Stat.StatDefenseRating]: -0.0159,
		[Stat.StatHitRating]: 3.9008,
		[Stat.StatCritRating]: 1.4031,
		[Stat.StatHasteRating]: 0.6615,
		[Stat.StatExpertiseRating]: 7.3663,
		[Stat.StatPhysicalDamage]: 6.1037,
		[Stat.StatArmorPenetration]: 0.3603,
	},
	{
		[PseudoStat.PseudoStatMainHandDps]: 13.2052,
	},
);

// Survival is TMI with stamina as 1, as the EP dialog scales it for the bear. Hit, haste and
// physical damage are within twice their estimated sampling noise, so they are 0. Expertise is
// kept: fewer parries mean fewer parry-hasted boss swings.
export const Level60SurvivalEp = Stats.fromMap({
	[Stat.StatStamina]: 1,
	[Stat.StatAgility]: 0.186,
	[Stat.StatArmor]: 0.1643,
	[Stat.StatBonusArmor]: 0.0363,
	[Stat.StatDodgeRating]: 0.1385,
	[Stat.StatDefenseRating]: 0.8278,
	[Stat.StatCritRating]: -0.0016,
	[Stat.StatExpertiseRating]: 0.3003,
});

export const Level60DpsEpWeights = PresetUtils.makePresetEpWeights('Level 60 DPS', Level60DpsEp);
export const Level60SurvivalEpWeights = PresetUtils.makePresetEpWeights('Level 60 Survival', Level60SurvivalEp);

export const DefaultOptions = DruidOptions.create({
	startingRage: 0,
});

export const DefaultConsumables = ConsumesSpec.create({
	elixirIds: [Consumables.ElixirOfTheMongoose.id, Consumables.GiftOfArthas.id],
	foodId: Consumables.DirgesKickinChimaerokChops.id,
	potId: Consumables.GreaterStoneshieldPotion.id,
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
