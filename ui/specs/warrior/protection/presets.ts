import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { ConsumesSpec, HealingModel, Profession, PseudoStat, Race, Stat } from '@generated/proto/common';
import { SavedTalents } from '@generated/proto/ui';
import { ProtectionWarrior_Options as ProtectionWarriorOptions, WarriorStance } from '@generated/proto/warrior';
import { Stats } from '@sim/proto/stats';
import { OtherDefaults as SimUIOtherDefaults } from '@sim/spec_config';

import * as WarriorPresets from '../shared/presets';
import GenericApl from './apls/default.apl.json';
import PreBiSGearJson from './gear_sets/prebis.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const ROTATION_DEFAULT = PresetUtils.makePresetAPLRotation('Generic', GenericApl);

// Dungeon and world drops, the warrior's dungeon tier set, crafted items whose reagents include no raid
// materials and vendor items; no reputation or PvP rewards. 142 defense, past the 140 that makes a level 63
// boss unable to crit. The "of Defense" pieces are world-drop greens; the cloak, the first ring and Defender's
// Grip Stabilizer are sold by the Undermine Trader in Stranglethorn Vale; Deeprock Bracers are a Silithus quest
// reward. The second ring and the Sentinel's Gloves plans have no known source.
export const PreBiSGear = PresetUtils.makePresetGear('Pre-BiS', PreBiSGearJson);

export const ProtectionTalents = PresetUtils.makePresetTalents('Protection', SavedTalents.create({ talentsString: '35300003-05-050531120300021351' }));

// Measured on the Pre-BiS gear with the Protection talents: 0.6 x the threat weights (Strength = 1) plus
// 1.15 x the TMI weights (Stamina = 1). Defense is its value past the 140 cap.
export const PreBiSEpWeights = PresetUtils.makePresetEpWeights(
	'Pre-BiS',
	Stats.fromMap(
		{
			[Stat.StatStamina]: 1.15,
			[Stat.StatStrength]: 0.64,
			[Stat.StatAgility]: 0.85,
			[Stat.StatAttackPower]: 0.27,
			[Stat.StatHitRating]: 1.16,
			[Stat.StatCritRating]: 0.64,
			[Stat.StatHasteRating]: 0.37,
			[Stat.StatExpertiseRating]: 2.34,
			[Stat.StatArmorPenetration]: 0.12,
			[Stat.StatDefenseRating]: 0.51,
			[Stat.StatBlockRating]: 0.04,
			[Stat.StatBlockValue]: 0.92,
			[Stat.StatDodgeRating]: 0.35,
			[Stat.StatParryRating]: 0.38,
			[Stat.StatArmor]: 0.05,
			[Stat.StatBonusArmor]: 0.05,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 2.17,
		},
	),
);

export const DefaultOptions = ProtectionWarriorOptions.create({
	classOptions: {
		queueDelay: 250,
		startingRage: 0,
		useBattleShout: true,
		defaultStance: WarriorStance.WarriorStanceDefensive,
		hasBsT2: true,
		stanceSnapshot: true,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	...WarriorPresets.DefaultConsumables,
	elixirIds: [
		Consumables.ElixirOfFerocity.id,
		Consumables.ElixirOfTheMongoose.id,
		Consumables.WinterfallFirewater.id,
		Consumables.ElixirOfGreaterDefense.id,
		Consumables.ElixirOfThePhalanx.id,
	],
	mhImbueId: Consumables.DenseWeightstone.id,
	conjuredId: Consumables.MajorHealthstone9421.id,
	extraIds: [Consumables.GoblinSapperCharge.id, Consumables.RumseyRum.id],
});

// Classic Era Molten Core logs: the healing that lands on the main tank, overhealing included, comes to 2.4 x
// its damage taken, in heals about 0.3 s apart, 1.5% of it as absorbs. The HPS is 2.4 x the Pre-BiS gear's damage
// taken per second; the Inspiration uptime is not from the logs.
export const OtherDefaults: Partial<SimUIOtherDefaults> = {
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	race: Race.RaceOrc,
	distanceFromTarget: 0,
	healingModel: HealingModel.create({
		hps: 445,
		cadenceSeconds: 0.3,
		cadenceVariation: 0.64,
		absorbFrac: 0.015,
		burstWindow: 6,
		inspirationUptime: 0.25,
	}),
};
