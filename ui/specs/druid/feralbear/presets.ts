import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { ConsumesSpec, HealingModel, Profession, PseudoStat, Race, Spec, Stat } from '@generated/proto/common';
import { FeralBearDruid_Options as DruidOptions } from '@generated/proto/druid';
import { SavedTalents } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';
import { Encounter } from '@sim/raid/encounter';
import { OtherDefaults as SimUIOtherDefaults } from '@sim/spec_config';

import DefaultApl from './apls/default.apl.json';
import StarterBuild from './builds/starter.build.json';
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
		[Stat.StatStrength]: 2.0597,
		[Stat.StatAgility]: 1,
		[Stat.StatStamina]: 0.0967,
		[Stat.StatAttackPower]: 0.9362,
		[Stat.StatFeralAttackPower]: 0.9362,
		[Stat.StatArmor]: 0.0265,
		[Stat.StatBonusArmor]: 0.0059,
		[Stat.StatDodgeRating]: -0.1122,
		[Stat.StatDefenseRating]: 0.0056,
		[Stat.StatHitRating]: 3.622,
		[Stat.StatCritRating]: 1.386,
		[Stat.StatHasteRating]: 0.7935,
		[Stat.StatExpertiseRating]: 7.0556,
		[Stat.StatPhysicalDamage]: 6.0751,
		[Stat.StatArmorPenetration]: 0.358,
	},
	{ [PseudoStat.PseudoStatMainHandDps]: 13.107 },
);

// Survival is TMI with stamina as 1. Signed model weights are retained; small values can be noisy.
export const Level60SurvivalEp = Stats.fromMap({
	[Stat.StatStamina]: 1,
	[Stat.StatHitRating]: -0.0034,
	[Stat.StatHasteRating]: -0.0019,
	[Stat.StatPhysicalDamage]: -0.0047,
	[Stat.StatAgility]: 0.1967,
	[Stat.StatArmor]: 0.1671,
	[Stat.StatBonusArmor]: 0.0365,
	[Stat.StatDodgeRating]: 0.1816,
	[Stat.StatDefenseRating]: 0.8494,
	[Stat.StatCritRating]: 0.0013,
	[Stat.StatExpertiseRating]: 0.3424,
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
	rotation: ROTATION_DEFAULT,
});
