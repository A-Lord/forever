import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { ConsumesSpec, HealingModel, Profession, PseudoStat, Race, Spec, Stat } from '@generated/proto/common';
import { FeralBearDruid_Options as DruidOptions } from '@generated/proto/druid';
import { Stats } from '@sim/proto/stats';
import { Encounter } from '@sim/raid/encounter';
import { OtherDefaults as SimUIOtherDefaults } from '@sim/spec_config';

import DefaultApl from './apls/default.apl.json';
import LacerateFirstApl from './apls/lacerate_first.apl.json';
import StarterBuild from './builds/starter.build.json';
import Level60GearJson from './gear_sets/default.gear.json';
export const ROTATION_DEFAULT = PresetUtils.makePresetAPLRotation('APL', DefaultApl);
// For a target no one else bleeds (Assume Bleed Always Active off): Berserk waits for Lacerate, so
// Rend and Tear covers it.
export const ROTATION_LACERATE_FIRST = PresetUtils.makePresetAPLRotation('Lacerate before Berserk', LacerateFirstApl);

// Leather at 60 with no set bonus and no proc, about the item level of the cat preset.
export const Level60Gear = PresetUtils.makePresetGear('Level 60', Level60GearJson, {
	tooltip: 'Leather at level 60 with no set bonus and no proc. The idol slot is empty: every idol has an effect.',
});

// A fixed 6000-damage target keeps the reference setup consistent.
const level60Encounter = Encounter.defaultEncounterProto();
level60Encounter.targets[0].minBaseDamage = 6000;
export const Level60Encounter = {
	...PresetUtils.makePresetEncounter('Level 60 boss', level60Encounter),
	tooltip: 'A reference raid target hitting for 6000 before armor.',
};

// EP on the Level 60 preset, the APL preset and the Level 60 boss, 50000 requested iterations on the committed input cohort.
// DPS has agility as 1. A bear's agility is crit, armor and dodge but no attack power, so strength
// is worth more than one agility. Dodge and defense cost damage: fewer hits taken is less rage.
export const Level60DpsEp = Stats.fromMap(
	{
		[Stat.StatStrength]: 2.0093,
		[Stat.StatAgility]: 1,
		[Stat.StatStamina]: 0.1002,
		[Stat.StatAttackPower]: 0.9133,
		[Stat.StatFeralAttackPower]: 0.9133,
		[Stat.StatArmor]: 0.0235,
		[Stat.StatBonusArmor]: 0.0065,
		[Stat.StatDodgeRating]: -0.0875,
		[Stat.StatDefenseRating]: 0.0351,
		[Stat.StatHitRating]: 3.796,
		[Stat.StatCritRating]: 1.3727,
		[Stat.StatHasteRating]: 0.6619,
		[Stat.StatExpertiseRating]: 7.1969,
		[Stat.StatPhysicalDamage]: 5.9387,
		[Stat.StatArmorPenetration]: 0.3495,
	},
	{
		[PseudoStat.PseudoStatMainHandDps]: 12.7863,
	},
);

// Survival is TMI with stamina as 1. Signed model weights are retained; small values can be noisy.
export const Level60SurvivalEp = Stats.fromMap({
	[Stat.StatStamina]: 1,
	[Stat.StatHitRating]: 0.0165,
	[Stat.StatHasteRating]: -0.0396,
	[Stat.StatPhysicalDamage]: 0.012,
	[Stat.StatAgility]: 0.1836,
	[Stat.StatArmor]: 0.1648,
	[Stat.StatBonusArmor]: 0.0361,
	[Stat.StatDodgeRating]: 0.1525,
	[Stat.StatDefenseRating]: 0.844,
	[Stat.StatCritRating]: -0.0054,
	[Stat.StatExpertiseRating]: 0.2892,
});

export const Level60DpsEpWeights = PresetUtils.makePresetEpWeights('Level 60 DPS', Level60DpsEp);
export const Level60SurvivalEpWeights = PresetUtils.makePresetEpWeights('Level 60 Survival', Level60SurvivalEp);

// Another raid member's bleed keeps Rend and Tear up.
export const DefaultOptions = DruidOptions.create({
	startingRage: 0,
	assumeBleedActive: true,
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

export const STARTER_GEAR = PresetUtils.makePresetGear('Level 60 Starter', StarterBuild.player.equipment);
export const STARTER_BUILD = PresetUtils.makePresetBuildFromJSON('Level 60 Starter', Spec.SpecFeralBearDruid, StarterBuild, {
	gear: STARTER_GEAR,
	rotation: ROTATION_LACERATE_FIRST,
});
