import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { ConsumesSpec, HandType, ItemSlot, Profession, PseudoStat, Race, Spec, Stat } from '@generated/proto/common';
import { SavedTalents } from '@generated/proto/ui';
import { DpsWarrior_Options as WarriorOptions, DpsWarrior_Rotation, DpsWarriorSpec, WarriorStance, WarriorSunder } from '@generated/proto/warrior';
import { Player } from '@sim/player/player';
import { Stats } from '@sim/proto/stats';

import * as WarriorPresets from '../shared/presets';
import DefaultArmsApl from './apls/arms.apl.json';
import ArmsStanceDanceApl from './apls/arms_stance_dance.apl.json';
import DefaultFuryApl from './apls/fury.apl.json';
import PreBiSArmsGearJson from './gear_sets/prebis_arms.gear.json';
import PreBiSFuryGearJson from './gear_sets/prebis_fury.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const isArmsSpec = (player: Player<Spec.SpecDpsWarrior>) =>
	player.getEquippedItem(ItemSlot.ItemSlotMainHand)?.item.handType === HandType.HandTypeTwoHand;

export const isArmsKebabSpec = (player: Player<Spec.SpecDpsWarrior>) => player.getTalents().mortalStrike && isFurySpec(player);

export const isFurySpec = (player: Player<Spec.SpecDpsWarrior>) =>
	player.getTalents().bloodthirst ||
	player.getEquippedItem(ItemSlot.ItemSlotMainHand)?.item.handType === HandType.HandTypeMainHand ||
	player.getEquippedItem(ItemSlot.ItemSlotMainHand)?.item.handType === HandType.HandTypeOneHand;

// Dungeon and world drops, the warrior's dungeon tier sets and crafted items whose reagents include no
// raid materials; no reputation rewards. The cloak, the belt and the Adaptive Combat Assistant are sold by
// the Undermine Trader in Stranglethorn Vale; the second ring has no known source.
export const PreBiSFuryGear = PresetUtils.makePresetGear('Pre-BiS Fury', PreBiSFuryGearJson);

// The best build found on the Pre-BiS Fury gear over a three minute fight.
export const FuryTalents = PresetUtils.makePresetTalents('Fury', SavedTalents.create({ talentsString: '30305013002-25053005052010501' }));

// Measured on the Pre-BiS Fury gear with the Fury talents.
export const PreBiSFuryEpWeights = PresetUtils.makePresetEpWeights(
	'Pre-BiS Fury',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 2.2,
			[Stat.StatAgility]: 2.35,
			[Stat.StatAttackPower]: 1,
			[Stat.StatArmorPenetration]: 0.34,
			[Stat.StatHitRating]: 1.64,
			[Stat.StatHasteRating]: 2.09,
			[Stat.StatCritRating]: 3.06,
			[Stat.StatExpertiseRating]: 1.06,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 7.91,
			[PseudoStat.PseudoStatOffHandDps]: 2.66,
		},
	),
);

// The same sources as the Pre-BiS Fury gear. The neck, cloak and second ring have no known source.
export const PreBiSArmsGear = PresetUtils.makePresetGear('Pre-BiS Arms', PreBiSArmsGearJson);

// The Battle Stance build: Rend keeps Bloodthrill's Overpower procs coming, and Spearing Strike fills the
// global cooldowns Mortal Strike and Slam leave.
export const ArmsTalents = PresetUtils.makePresetTalents('Arms', SavedTalents.create({ talentsString: '30325213132515201-0505-2' }));

// Measured on the Pre-BiS Arms gear with the Arms talents. Hit is its value below the 8% cap.
export const PreBiSArmsEpWeights = PresetUtils.makePresetEpWeights(
	'Pre-BiS Arms',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 2.2,
			[Stat.StatAgility]: 1.86,
			[Stat.StatAttackPower]: 1,
			[Stat.StatArmorPenetration]: 0.37,
			[Stat.StatHitRating]: 4.03,
			[Stat.StatHasteRating]: 1.75,
			[Stat.StatCritRating]: 2.42,
			[Stat.StatExpertiseRating]: 1.31,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 13.92,
			[PseudoStat.PseudoStatOffHandDps]: 0,
		},
	),
);

export const FURY_DEFAULT_ROTATION = PresetUtils.makePresetAPLRotation('Fury', DefaultFuryApl);
export const ARMS_DEFAULT_ROTATION = PresetUtils.makePresetAPLRotation('Arms', DefaultArmsApl);
export const ARMS_STANCE_DANCE_ROTATION = PresetUtils.makePresetAPLRotation('Arms (Stance Dance)', ArmsStanceDanceApl);

export const SIMPLE_ROTATION = DpsWarrior_Rotation.create({
	spec: DpsWarriorSpec.DpsWarriorSpecFury,
	sunderArmor: WarriorSunder.WarriorSunderHelp,
	useOverpower: true,
	useRend: true,
	useRecklessness: false,
});
export const SIMPLE_DEFAULT_ROTATION = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecDpsWarrior, SIMPLE_ROTATION);
export const SIMPLE_ARMS_DEFAULT_ROTATION = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecDpsWarrior, {
	...SIMPLE_ROTATION,
	spec: DpsWarriorSpec.DpsWarriorSpecArms,
});

export const DefaultOptions = WarriorOptions.create({
	classOptions: {
		queueDelay: 250,
		startingRage: 50,
		useBattleShout: true,
		defaultStance: WarriorStance.WarriorStanceBerserker,
		hasBsT2: true,
		stanceSnapshot: true,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	...WarriorPresets.DefaultConsumables,
	mhImbueId: Consumables.DenseWeightstone.id,
	ohImbueId: Consumables.DenseSharpeningStone.id,
	elixirIds: [Consumables.ElixirOfFerocity.id, Consumables.ElixirOfTheMongoose.id, Consumables.WinterfallFirewater.id],
});

export const OtherDefaults = {
	race: Race.RaceOrc,
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	distanceFromTarget: 25,
};
