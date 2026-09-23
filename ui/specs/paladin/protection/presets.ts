import * as PresetUtils from '@app/preset_utils';
import { ConsumesSpec, Debuffs, HealingModel, IndividualBuffs, PartyBuffs, Profession, RaidBuffs, Spec, TristateEffect } from '@generated/proto/common';
import {
	PaladinAura,
	PaladinJudgement,
	ProtectionPaladin_Options as ProtectionPaladinOptions,
	ProtectionPaladin_Rotation as ProtectionPaladinRotation,
} from '@generated/proto/paladin';
import { defaultExposeWeaknessSettings } from '@sim/proto/utils';

import DefaultApl from './apls/default.apl.json';

export const APL_PRESET = PresetUtils.makePresetAPLRotation('Default', DefaultApl);

export const DefaultSimpleRotation = ProtectionPaladinRotation.create({
	prioritizeHolyShield: true,
	consecrationRank: 6,
	useExorcism: true,
	useHammerOfWrath: false,
	maintainJudgement: PaladinJudgement.JudgementNone,
	aura: PaladinAura.DevotionAura,
});

export const APL_SIMPLE = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecProtectionPaladin, DefaultSimpleRotation);

export const DefaultOptions = ProtectionPaladinOptions.create({
	classOptions: {},
});

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: 22861, // Flask of Blinding Light
	foodId: 27657, // Blackened Basilisk
	potId: 22849, // Ironshield Potion
	conjuredId: 12662, // Dark Rune
	mhImbueId: 20749, // Brilliant Wizard Oil
	explosiveId: 18588,
	extraIds: [10646, 10310, 10309, 10305],
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
	wrathOfAirTotem: TristateEffect.TristateEffectRegular,
	graceOfAirTotem: TristateEffect.TristateEffectMissing,
	strengthOfEarthTotem: TristateEffect.TristateEffectImproved,
	windfuryTotem: TristateEffect.TristateEffectMissing,
	battleShout: TristateEffect.TristateEffectMissing,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfWisdom: true,
	blessingOfMight: true,
});

export const DefaultDebuffs = Debuffs.create({
	misery: true,
	curseOfElements: TristateEffect.TristateEffectImproved,
	judgementOfWisdom: true,
	judgementOfLight: true,
	bloodFrenzy: true,
	huntersMark: TristateEffect.TristateEffectImproved,
	curseOfRecklessness: true,
	sunderArmor: true,
	faerieFire: TristateEffect.TristateEffectImproved,
	exposeArmor: TristateEffect.TristateEffectImproved,
	insectSwarm: true,
	...defaultExposeWeaknessSettings(),
});

export const OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Enchanting,
	distanceFromTarget: 5,
	iterationCount: 25000,
	healingModel: HealingModel.create({
		hps: 2200,
		cadenceSeconds: 0.4,
		cadenceVariation: 1.2,
		absorbFrac: 0.02,
		burstWindow: 6,
		inspirationUptime: 0.25,
	}),
};
