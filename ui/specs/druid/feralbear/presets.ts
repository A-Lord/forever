import * as PresetUtils from '@app/preset_utils';
import { ConsumesSpec, HealingModel, Profession, PseudoStat, Race, Stat } from '@generated/proto/common';
import { FeralBearDruid_Options as DruidOptions } from '@generated/proto/druid';
import { SavedTalents } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';
import { Encounter } from '@sim/raid/encounter';
import { OtherDefaults as SimUIOtherDefaults } from '@sim/spec_config';

import * as Consumables from '@features/settings/model/consumables_auto_gen';

import DefaultApl from './apls/default.apl.json';
import LacerateFirstApl from './apls/lacerate_first.apl.json';
import Level60GearJson from './gear_sets/default.gear.json';
export const ROTATION_DEFAULT = PresetUtils.makePresetAPLRotation('APL', DefaultApl);
// For a target no one else bleeds (Assume Bleed Always Active off): Berserk waits for Lacerate, so
// Rend and Tear covers it.
export const ROTATION_LACERATE_FIRST = PresetUtils.makePresetAPLRotation('Lacerate before Berserk', LacerateFirstApl);

// Nature's Reach (melee hit) and Nature's Majesty (melee crit) are worth more to the bear than Furor and
// Naturalist. Feral Instinct raises only Swipe, which the single-target APL does not cast.
export const BearTalents = PresetUtils.makePresetTalents(
	'Bear',
	SavedTalents.create({ talentsString: '0500220000000000-55200322120132012551-0000000000000000' }),
);
// Tanks and plays cat on one set of talents. As a bear, about the same threat as the Bear talents and a
// TMI 2 to 3 higher, for no Natural Reaction.
export const HybridTalents = PresetUtils.makePresetTalents(
	'Hybrid 9/42/0',
	SavedTalents.create({ talentsString: '0500220000000000-35230332121132212051-0000000000000000' }),
);

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
		[Stat.StatStrength]: 2.2833,
		[Stat.StatAgility]: 1,
		[Stat.StatStamina]: 0.1094,
		[Stat.StatAttackPower]: 1.0379,
		[Stat.StatFeralAttackPower]: 0,
		[Stat.StatArmor]: 0.0325,
		[Stat.StatBonusArmor]: 0.006,
		[Stat.StatDodgeRating]: -0.3,
		[Stat.StatDefenseRating]: -0.197,
		[Stat.StatHitRating]: 4.2963,
		[Stat.StatCritRating]: 1.5314,
		[Stat.StatHasteRating]: 0.6891,
		[Stat.StatExpertiseRating]: 8.1496,
		[Stat.StatPhysicalDamage]: 6.7094,
		[Stat.StatArmorPenetration]: 0.3958,
	},
	{
		[PseudoStat.PseudoStatMainHandDps]: 14.5303,
	},
);

// Survival is TMI with stamina as 1, as the EP dialog scales it for the bear. Hit, crit, haste and
// physical damage moved TMI by no more than about twice their noise, so they are 0. Expertise is
// kept: fewer parries mean fewer parry-hasted boss swings.
export const Level60SurvivalEp = Stats.fromMap({
	[Stat.StatStamina]: 1,
	[Stat.StatAgility]: 0.1909,
	[Stat.StatArmor]: 0.1661,
	[Stat.StatBonusArmor]: 0.0365,
	[Stat.StatDodgeRating]: 0.1509,
	[Stat.StatDefenseRating]: 0.506,
	[Stat.StatExpertiseRating]: 0.2922,
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
