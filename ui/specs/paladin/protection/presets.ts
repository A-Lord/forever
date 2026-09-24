import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, HealingModel, Profession, Spec, TristateEffect } from '@generated/proto/common';
import {
	PaladinAura,
	PaladinJudgement,
	ProtectionPaladin_Options as ProtectionPaladinOptions,
	ProtectionPaladin_Rotation as ProtectionPaladinRotation,
} from '@generated/proto/paladin';

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
	flaskId: Consumables.FlaskOfSupremePower.id,
	foodId: Consumables.SunriseOmelette.id,
	potId: Consumables.GreaterStoneshieldPotion.id,
	conjuredId: Consumables.DemonicRune.id,
	mhImbueId: Consumables.BrilliantWizardOil.id,
	explosiveId: Consumables.EzThroDynamiteII.id,
	extraIds: [Consumables.GoblinSapperCharge.id, Consumables.ScrollOfStrengthIV.id, Consumables.ScrollOfAgilityIV.id, Consumables.ScrollOfProtectionIV.id],
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
	graceOfAirTotem: false,
	strengthOfEarthTotem: true,
	windfuryTotem: false,
	battleShout: TristateEffect.TristateEffectMissing,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	greaterBlessingOfKings: true,
	greaterBlessingOfWisdom: true,
	greaterBlessingOfMight: true,
});

export const DefaultDebuffs = Debuffs.create({
	curseOfElements: true,
	judgementOfWisdom: true,
	judgementOfLight: true,
	huntersMark: true,
	curseOfRecklessness: true,
	sunderArmor: true,
	faerieFire: true,
	exposeArmor: true,
	insectSwarm: true,
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
