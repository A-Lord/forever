import * as PresetUtils from '@app/preset_utils';
import { Class, ConsumesSpec, Debuffs, IndividualBuffs, PartyBuffs, Profession, Race, RaidBuffs, TristateEffect } from '@generated/proto/common';
import { ElementalShaman_Options as ElementalShamanOptions } from '@generated/proto/shaman';
import { defaultRaidBuffMajorDamageCooldowns } from '@sim/proto/utils';

import DefaultApl from './apls/default.apl.json';

export const ROTATION_PRESET_DEFAULT = PresetUtils.makePresetAPLRotation('Default', DefaultApl);

export const DefaultOptions = ElementalShamanOptions.create({
	classOptions: {
		shieldProcrate: 0,
	},
});

export const OtherDefaults = {
	distanceFromTarget: 20,
	profession1: Profession.Leatherworking,
	profession2: Profession.Enchanting,
	race: Race.RaceDraenei,
};

export const DefaultRaidBuffs = RaidBuffs.create({
	...defaultRaidBuffMajorDamageCooldowns(Class.ClassShaman),
	arcaneBrilliance: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	divineSpirit: TristateEffect.TristateEffectImproved,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	moonkinAura: TristateEffect.TristateEffectImproved,
	chainOfTheTwilightOwl: true,
	eyeOfTheNight: true,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfWisdom: true,
	shadowPriestDps: 800,
});

export const DefaultDebuffs = Debuffs.create({
	bloodFrenzy: true,
	curseOfElements: TristateEffect.TristateEffectImproved,
	curseOfRecklessness: true,
	exposeArmor: TristateEffect.TristateEffectImproved,
	faerieFire: TristateEffect.TristateEffectImproved,
	giftOfArthas: true,
	huntersMark: TristateEffect.TristateEffectImproved,
	judgementOfWisdom: true,
	mangle: true,
	misery: true,
	sunderArmor: true,
});

export const DefaultConsumables = ConsumesSpec.create({
	conjuredId: 12662, // Demonic Rune
	flaskId: 22861, // Flask of Blinding Light
	foodId: 27657, // Blackened Basilisk
	mhImbueId: 20749, // Brilliant Wizard Oil
	potId: 22839, // Destruction Potion
});
