import * as PresetUtils from '@app/preset_utils';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Profession, Race, Spec, TristateEffect } from '@generated/proto/common';
import { PaladinAura, RetributionPaladin_Options as RetributionPaladinOptions, RetributionPaladin_Rotation as PaladinRotation } from '@generated/proto/paladin';

import DefaultApl from './apls/default.apl.json';

export const DefaultSimpleRotation = PaladinRotation.create({
	useExorcism: false,
	consecrationRank: 0,
	delayMajorCDs: 11,
	prepullSotC: true,
	aura: PaladinAura.RetributionAura,
});

export const APL_PRESET = PresetUtils.makePresetAPLRotation('Default', DefaultApl);
export const APL_SIMPLE = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecRetributionPaladin, DefaultSimpleRotation);

export const DefaultOptions = RetributionPaladinOptions.create({
	classOptions: {},
});

export const DefaultConsumables = ConsumesSpec.create({
	conjuredId: 12662,
	goblinSapper: true,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	prayerOfSpirit: true,
	arcaneBrilliance: true,
	giftOfTheWild: true,
	prayerOfFortitude: true,
	prayerOfShadowProtection: true,
	thorns: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	manaSpringTotem: TristateEffect.TristateEffectRegular,
	leaderOfThePack: true,
	battleShout: TristateEffect.TristateEffectRegular,
	strengthOfEarthTotem: true,
	totemTwisting: true,
	windfuryTotem: true,
	graceOfAirTotem: true,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	greaterBlessingOfKings: true,
	greaterBlessingOfWisdom: true,
	greaterBlessingOfMight: true,
});

export const DefaultDebuffs = Debuffs.create({
	curseOfElements: true,
	judgementOfWisdom: true,
	huntersMark: true,
	curseOfRecklessness: true,
	sunderArmor: true,
	faerieFire: true,
	exposeArmor: true,
});

export const OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	distanceFromTarget: 5,
	iterationCount: 25000,
	race: Race.RaceHuman,
};
