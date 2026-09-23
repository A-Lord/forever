import * as PresetUtils from '@app/preset_utils';
import { ConsumesSpec, Debuffs, Drums, IndividualBuffs, PartyBuffs, Profession, Race, RaidBuffs, Spec, TristateEffect } from '@generated/proto/common';
import { PaladinAura, RetributionPaladin_Options as RetributionPaladinOptions, RetributionPaladin_Rotation as PaladinRotation } from '@generated/proto/paladin';
import { defaultExposeWeaknessSettings } from '@sim/proto/utils';

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
	potId: 22838,
	flaskId: 22854,
	foodId: 27658,
	conjuredId: 12662,
	goblinSapper: true,
	scrollAgi: true,
	scrollStr: true,
	explosiveId: 18588,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	bloodlust: true,
	divineSpirit: TristateEffect.TristateEffectImproved,
	arcaneBrilliance: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	shadowProtection: true,
	thorns: TristateEffect.TristateEffectImproved,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	manaSpringTotem: TristateEffect.TristateEffectRegular,
	leaderOfThePack: TristateEffect.TristateEffectImproved,
	battleShout: TristateEffect.TristateEffectImproved,
	strengthOfEarthTotem: TristateEffect.TristateEffectImproved,
	totemTwisting: true,
	windfuryTotem: TristateEffect.TristateEffectImproved,
	graceOfAirTotem: TristateEffect.TristateEffectImproved,
	drums: Drums.LesserDrumsOfBattle,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfWisdom: true,
	blessingOfMight: true,
	unleashedRage: true,
});

export const DefaultDebuffs = Debuffs.create({
	misery: true,
	curseOfElements: TristateEffect.TristateEffectImproved,
	judgementOfWisdom: true,
	bloodFrenzy: true,
	huntersMark: TristateEffect.TristateEffectImproved,
	curseOfRecklessness: true,
	sunderArmor: true,
	faerieFire: TristateEffect.TristateEffectImproved,
	exposeArmor: TristateEffect.TristateEffectImproved,
	...defaultExposeWeaknessSettings(),
});

export const OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	distanceFromTarget: 5,
	iterationCount: 25000,
	race: Race.RaceBloodElf,
};
