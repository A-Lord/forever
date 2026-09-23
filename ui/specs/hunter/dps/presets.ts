import * as PresetUtils from '@app/preset_utils';
import { Class, ConsumesSpec, Debuffs, IndividualBuffs, PartyBuffs, Profession, Race, RaidBuffs, Spec, TristateEffect } from '@generated/proto/common';
import {
	Hunter_Options as HunterOptions,
	Hunter_Rotation,
	HunterOptions_Ammo,
	HunterOptions_PetType as PetType,
	HunterOptions_QuiverBonus,
} from '@generated/proto/hunter';
import { defaultExposeWeaknessSettings, defaultRaidBuffMajorDamageCooldowns } from '@sim/proto/utils';

import DefaultAPL from './apls/default.apl.json';

export const DefaultRotation = PresetUtils.makePresetAPLRotation('APL', DefaultAPL);

export const TurretRotation = Hunter_Rotation.create({
	viperStartManaPercent: 0.05,
	viperStopManaPercent: 0.25,
	meleeWeave: false,
	timeToWeave: 400,
	useMulti: true,
	useArcane: true,
});
export const TurretSimple = PresetUtils.makePresetSimpleRotation('Turret', Spec.SpecHunter, TurretRotation);

export const WeaveRotation = Hunter_Rotation.create({
	viperStartManaPercent: 0.05,
	viperStopManaPercent: 0.25,
	meleeWeave: true,
	timeToWeave: 400,
	useMulti: true,
	useArcane: true,
});
export const WeaveSimple = PresetUtils.makePresetSimpleRotation('Weave', Spec.SpecHunter, WeaveRotation);

export const DefaultOptions = HunterOptions.create({
	classOptions: {
		ammo: HunterOptions_Ammo.WardensArrow,
		quiverBonus: HunterOptions_QuiverBonus.Speed15,
		petType: PetType.Ravager,
		petUptime: 1,
		petSingleAbility: false,
	},
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfMight: true,
	blessingOfWisdom: true,
	unleashedRage: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	battleShout: TristateEffect.TristateEffectImproved,
	braidedEterniumChain: true,
	ferociousInspiration: 1,
	graceOfAirTotem: TristateEffect.TristateEffectImproved,
	leaderOfThePack: TristateEffect.TristateEffectImproved,
	strengthOfEarthTotem: TristateEffect.TristateEffectImproved,
	totemTwisting: true,
	windfuryTotem: TristateEffect.TristateEffectImproved,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	...defaultRaidBuffMajorDamageCooldowns(Class.ClassWarrior),
	arcaneBrilliance: true,
	divineSpirit: TristateEffect.TristateEffectImproved,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	shadowProtection: true,
});

export const DefaultDebuffs = Debuffs.create({
	bloodFrenzy: true,
	curseOfElements: TristateEffect.TristateEffectImproved,
	curseOfRecklessness: true,
	exposeArmor: TristateEffect.TristateEffectImproved,
	...defaultExposeWeaknessSettings(),
	faerieFire: TristateEffect.TristateEffectImproved,
	giftOfArthas: true,
	huntersMark: TristateEffect.TristateEffectImproved,
	insectSwarm: true,
	judgementOfLight: true,
	judgementOfWisdom: true,
	mangle: true,
	misery: true,
	sunderArmor: true,
});

export const DefaultConsumables = ConsumesSpec.create({
	battleElixirId: 22831, // Elixir of Major Agility
	guardianElixirId: 22840, // Elixir of Major Mageblood
	foodId: 27659, // Warp Burger
	potId: 22838, // Haste Potion
	conjuredId: 12662,
	explosiveId: 18588,
	petFoodId: 33874, // Kibler's Bits
	petScrollAgi: true,
	petScrollStr: true,
	goblinSapper: true,
	scrollAgi: true,
	scrollStr: true,
});

export const OtherDefaults = {
	distanceFromTarget: 7,
	iterationCount: 25000,
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	race: Race.RaceOrc,
};
