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
import PreBiSThreatGearJson from './gear_sets/prebis_threat.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const ROTATION_DEFAULT = PresetUtils.makePresetAPLRotation('Generic', GenericApl);

// Dungeon and world drops, the warrior's dungeon tier set, crafted items whose reagents include no raid
// materials and vendor items; no reputation or PvP rewards. The most threat at 142 defense, past the 140 that
// makes a level 63 boss unable to crit. The "of Defense" pieces are world-drop greens; the cloak, the first ring
// and Defender's Grip Stabilizer are sold by the Undermine Trader in Stranglethorn Vale; Beads of Ogre Might is a
// Dire Maul quest reward and Deeprock Bracers a Silithus quest reward. Horned Pauldron, the Dawn Armor pattern,
// the second ring and the Sentinel's Gloves plans have no known source.
export const PreBiSGear = PresetUtils.makePresetGear('Pre-BiS (Defense Cap)', PreBiSGearJson);

// The same sources with no defense floor: the most threat that survives a Golemagg-strength boss on the default
// healing. Heart of the Scale drops in Lower Blackrock Spire; Band of Earthen Might is a Silithus quest reward.
export const PreBiSThreatGear = PresetUtils.makePresetGear('Pre-BiS (Threat)', PreBiSThreatGearJson);

export const ProtectionTalents = PresetUtils.makePresetTalents('Protection', SavedTalents.create({ talentsString: '35300003-05-050531120300021351' }));

// Measured on each gear set with the Protection talents: 0.6 x the threat weights (Strength = 1) plus 1.15 x the
// TMI weights (Stamina = 1). On the defense cap gear, defense is its value past the cap; on the threat gear, hit is
// its value past the 8% cap.
export const PreBiSEpWeights = PresetUtils.makePresetEpWeights(
	'Pre-BiS (Defense Cap)',
	Stats.fromMap(
		{
			[Stat.StatStamina]: 1.15,
			[Stat.StatStrength]: 0.63,
			[Stat.StatAgility]: 0.77,
			[Stat.StatAttackPower]: 0.27,
			[Stat.StatHitRating]: 1.14,
			[Stat.StatCritRating]: 0.61,
			[Stat.StatHasteRating]: 0.38,
			[Stat.StatExpertiseRating]: 2.28,
			[Stat.StatArmorPenetration]: 0.12,
			[Stat.StatDefenseRating]: 0.42,
			[Stat.StatBlockRating]: 0.05,
			[Stat.StatBlockValue]: 0.79,
			[Stat.StatDodgeRating]: 0.28,
			[Stat.StatParryRating]: 0.28,
			[Stat.StatArmor]: 0.05,
			[Stat.StatBonusArmor]: 0.05,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 2.15,
		},
	),
);

export const PreBiSThreatEpWeights = PresetUtils.makePresetEpWeights(
	'Pre-BiS (Threat)',
	Stats.fromMap(
		{
			[Stat.StatStamina]: 1.15,
			[Stat.StatStrength]: 0.63,
			[Stat.StatAgility]: 0.71,
			[Stat.StatAttackPower]: 0.27,
			[Stat.StatHitRating]: 0.07,
			[Stat.StatCritRating]: 0.58,
			[Stat.StatHasteRating]: 0.45,
			[Stat.StatExpertiseRating]: 2.34,
			[Stat.StatArmorPenetration]: 0.13,
			[Stat.StatDefenseRating]: 0.44,
			[Stat.StatBlockRating]: 0.06,
			[Stat.StatBlockValue]: 0.78,
			[Stat.StatDodgeRating]: 0.25,
			[Stat.StatParryRating]: 0.26,
			[Stat.StatArmor]: 0.05,
			[Stat.StatBonusArmor]: 0.05,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 2.31,
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
	mhImbueId: Consumables.DenseSharpeningStone.id,
	conjuredId: Consumables.MajorHealthstone9421.id,
	extraIds: [Consumables.GoblinSapperCharge.id, Consumables.RumseyRum.id],
});

// Classic Era Molten Core logs: the healing that lands on the main tank, overhealing included, comes to 2.4 x
// its damage taken, in heals about 0.3 s apart, 1.5% of it as absorbs. The HPS is 2.4 x the defense cap gear's
// damage taken per second against the default target; the Inspiration uptime is not from the logs.
export const OtherDefaults: Partial<SimUIOtherDefaults> = {
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	race: Race.RaceOrc,
	distanceFromTarget: 0,
	healingModel: HealingModel.create({
		hps: 530,
		cadenceSeconds: 0.3,
		cadenceVariation: 0.64,
		absorbFrac: 0.015,
		burstWindow: 6,
		inspirationUptime: 0.25,
	}),
};
