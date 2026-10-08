import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { ConsumesSpec, Profession, Race, Spec, Stat } from '@generated/proto/common';
import {
	FeralCatDruid_Options as FeralDruidOptions,
	FeralCatDruid_Rotation as FeralCatDruidRotation,
	FeralCatDruid_Rotation_FinishingMove as FinishingMove,
} from '@generated/proto/druid';
import { Stats } from '@sim/proto/stats';

import DefaultApl from './apls/default.apl.json';
import PounceApl from './apls/pounce.apl.json';
import RavageApl from './apls/ravage.apl.json';
import Level60GearJson from './gear_sets/default.gear.json';
import PreRaidGearJson from './gear_sets/pre_raid.gear.json';

export const DefaultOptions = FeralDruidOptions.create({});

export const DefaultConsumables = ConsumesSpec.create({
	potId: Consumables.MightyRagePotion.id,
	flaskId: Consumables.FlaskOfNaturalPrecision.id,
	elixirIds: [Consumables.ElixirOfTheMongoose.id],
	foodId: Consumables.SwiftstrikeSteak.id,
	mhImbueId: Consumables.ElementalSharpeningStone.id,
	conjuredId: Consumables.ThistleTea.id,
	extraIds: [Consumables.JujuPower.id, Consumables.JujuMight.id, Consumables.GroundScorpokAssay.id],
});

export const OtherDefaults = {
	distanceFromTarget: 0,
	profession1: Profession.Engineering,
	profession2: Profession.Enchanting,
	race: Race.RaceTauren,
	reactionTime: 250,
};

export const DefaultRotation = FeralCatDruidRotation.create({
	finishingMove: FinishingMove.Rip,
	ripMinComboPoints: 5,
	biteMinComboPoints: 5,
	rakeTrick: true,
	faerieFireForOmen: true,
});

// Leather at 60 with no set bonus and no proc. The stand-in the rotation was tuned on.
export const Level60Gear = PresetUtils.makePresetGear('Level 60', Level60GearJson, {
	tooltip: 'Leather at level 60 with no set bonus and no proc. Fist of Cenarius and the Howling Idol.',
});

// Pre-raid leather on a Tauren with The Unstoppable Force and Wolfshead Helm. Shifting Cloak, Dunestalker's
// Boots, Ferocious Watcher's Signet and the two hit and attack power trinkets stand in for Cape of the Black
// Baron, Swiftwalker Boots, Tarnished Elven Ring, Blackhand's Breadth and Hand of Justice, which the item
// database does not carry.
export const PreRaidGear = PresetUtils.makePresetGear('Pre-raid', PreRaidGearJson, {
	tooltip: 'Pre-raid leather with The Unstoppable Force, Wolfshead Helm and full enchants.',
});

// EP with agility as 1, on the Level 60 preset, 50000 iterations.
// Feral attack power is unread, so it stays 0. Armor penetration is 0 because
// Sunder, Expose and Faerie Fire already remove the target's armor.
export const Level60Ep = Stats.fromMap({
	[Stat.StatStrength]: 1.0102,
	[Stat.StatAgility]: 1,
	[Stat.StatAttackPower]: 0.4041,
	[Stat.StatFeralAttackPower]: 0,
	[Stat.StatPhysicalDamage]: 3.9557,
	[Stat.StatHitRating]: 0.2046,
	[Stat.StatExpertiseRating]: 0.6976,
	[Stat.StatCritRating]: 0.7201,
	[Stat.StatHasteRating]: 0.4679,
	[Stat.StatArmorPenetration]: 0,
});

export const Level60EpWeights = PresetUtils.makePresetEpWeights('Level 60', Level60Ep);

// EP with agility as 1, on the Pre-raid preset with the default buffs, debuffs and consumables, 180s.
// Hit is weighed 1% below the 8% cap, where the gear picker reads it.
export const PreRaidEp = Stats.fromMap({
	[Stat.StatStrength]: 0.94,
	[Stat.StatAgility]: 1,
	[Stat.StatAttackPower]: 0.39,
	[Stat.StatFeralAttackPower]: 0,
	[Stat.StatPhysicalDamage]: 3.71,
	[Stat.StatHitRating]: 0.78,
	[Stat.StatExpertiseRating]: 0.74,
	[Stat.StatCritRating]: 0.74,
	[Stat.StatHasteRating]: 0.76,
	[Stat.StatArmorPenetration]: 0.14,
});

export const PreRaidEpWeights = PresetUtils.makePresetEpWeights('Pre-raid', PreRaidEp);

export const SIMPLE = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecFeralCatDruid, DefaultRotation);

export const APL = PresetUtils.makePresetAPLRotation('APL', DefaultApl);
export const RAVAGE_APL = PresetUtils.makePresetAPLRotation('Ravage opener', RavageApl);
export const POUNCE_APL = PresetUtils.makePresetAPLRotation('Pounce opener', PounceApl);
