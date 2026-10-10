import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { ConsumesSpec, HealingModel, Profession, PseudoStat, Race, Stat } from '@generated/proto/common';
import { FeralBearDruid_Options as DruidOptions } from '@generated/proto/druid';
import { SavedTalents } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';
import { Encounter } from '@sim/raid/encounter';
import { OtherDefaults as SimUIOtherDefaults } from '@sim/spec_config';

import DefaultApl from './apls/default.apl.json';
import Level60GearJson from './gear_sets/default.gear.json';
export const ROTATION_DEFAULT = PresetUtils.makePresetAPLRotation('APL', DefaultApl);

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

// A level 60 target swings for 6000 before armor.
const level60Encounter = Encounter.defaultEncounterProto();
level60Encounter.targets[0].minBaseDamage = 6000;
export const Level60Encounter = {
	...PresetUtils.makePresetEncounter('Level 60 boss', level60Encounter),
	tooltip: 'The raid target, hitting for 6000 before armor.',
};

// EP on the Level 60 preset, the APL preset and the Level 60 boss, with a raid bleed assumed, 200000
// iterations. DPS and threat have agility as 1. A bear's agility is crit, armor and dodge but no attack
// power, so strength is worth about two agility. Dodge costs damage: fewer hits taken is less rage.
// Defense is close to 0, its Thick Hide armor against the hits it avoids. Stamina keeps the bear alive longer.
export const Level60DpsEp = Stats.fromMap(
	{
		[Stat.StatStrength]: 2.0609,
		[Stat.StatAgility]: 1,
		[Stat.StatStamina]: 0.1094,
		[Stat.StatAttackPower]: 0.9368,
		[Stat.StatFeralAttackPower]: 0.9368,
		[Stat.StatArmor]: 0.0258,
		[Stat.StatBonusArmor]: 0.0258,
		[Stat.StatDodgeRating]: -0.1057,
		[Stat.StatDefenseRating]: 0.0067,
		[Stat.StatHitRating]: 3.753,
		[Stat.StatCritRating]: 1.3774,
		[Stat.StatHasteRating]: 0.8336,
		[Stat.StatExpertiseRating]: 7.1215,
		[Stat.StatPhysicalDamage]: 6.0207,
		[Stat.StatArmorPenetration]: 0.3582,
	},
	{
		[PseudoStat.PseudoStatMainHandDps]: 13.1147,
	},
);

// Threat with agility as 1, on the same run.
export const Level60ThreatEp = Stats.fromMap(
	{
		[Stat.StatStrength]: 1.9952,
		[Stat.StatAgility]: 1,
		[Stat.StatStamina]: 0.1076,
		[Stat.StatAttackPower]: 0.9069,
		[Stat.StatFeralAttackPower]: 0.9069,
		[Stat.StatArmor]: 0.0253,
		[Stat.StatBonusArmor]: 0.0253,
		[Stat.StatDodgeRating]: -0.1067,
		[Stat.StatDefenseRating]: -0.0128,
		[Stat.StatHitRating]: 3.676,
		[Stat.StatCritRating]: 1.3786,
		[Stat.StatHasteRating]: 0.8125,
		[Stat.StatExpertiseRating]: 6.9497,
		[Stat.StatPhysicalDamage]: 5.8386,
		[Stat.StatArmorPenetration]: 0.3468,
	},
	{
		[PseudoStat.PseudoStatMainHandDps]: 12.6969,
	},
);

// Survival is TMI with stamina as 1, as the EP dialog scales it for the bear. Hit, crit, haste and
// physical damage moved TMI by less than 0.04, so they are 0. The bear forms scale item armor and bonus
// armor alike, so the two weigh the same. Expertise is kept: fewer parries mean fewer parry-hasted boss
// swings.
export const Level60SurvivalEp = Stats.fromMap({
	[Stat.StatStamina]: 1,
	[Stat.StatAgility]: 0.1964,
	[Stat.StatArmor]: 0.1617,
	[Stat.StatBonusArmor]: 0.1617,
	[Stat.StatDodgeRating]: 0.1777,
	[Stat.StatDefenseRating]: 0.8496,
	[Stat.StatExpertiseRating]: 0.3212,
});

export const Level60DpsEpWeights = PresetUtils.makePresetEpWeights('Level 60 DPS', Level60DpsEp);
export const Level60ThreatEpWeights = PresetUtils.makePresetEpWeights('Level 60 Threat', Level60ThreatEp);
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
