import * as PresetUtils from '@app/preset_utils';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Profession, Race, Spec, TristateEffect } from '@generated/proto/common';
import {
	Hunter_Options as HunterOptions,
	Hunter_Rotation,
	HunterOptions_Ammo,
	HunterOptions_PetType as PetType,
	HunterOptions_QuiverBonus,
} from '@generated/proto/hunter';

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
	greaterBlessingOfKings: true,
	greaterBlessingOfMight: true,
	greaterBlessingOfWisdom: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	battleShout: TristateEffect.TristateEffectRegular,
	graceOfAirTotem: true,
	leaderOfThePack: true,
	strengthOfEarthTotem: true,
	totemTwisting: true,
	windfuryTotem: true,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	arcaneBrilliance: true,
	prayerOfSpirit: true,
	giftOfTheWild: true,
	prayerOfFortitude: true,
	prayerOfShadowProtection: true,
});

export const DefaultDebuffs = Debuffs.create({
	curseOfElements: true,
	curseOfRecklessness: true,
	exposeArmor: true,
	faerieFire: true,
	giftOfArthas: true,
	huntersMark: true,
	insectSwarm: true,
	judgementOfLight: true,
	judgementOfWisdom: true,
	mangle: true,
	sunderArmor: true,
});

export const DefaultConsumables = ConsumesSpec.create({
	conjuredId: 12662,
	goblinSapper: true,
});

export const OtherDefaults = {
	distanceFromTarget: 7,
	iterationCount: 25000,
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	race: Race.RaceOrc,
};
