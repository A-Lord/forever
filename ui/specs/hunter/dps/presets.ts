import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Profession, Race, Spec, TristateEffect } from '@generated/proto/common';
import {
	Hunter_Options as HunterOptions,
	Hunter_Rotation,
	HunterOptions_Ammo,
	HunterOptions_PetAttackSpeed,
	HunterOptions_PetType as PetType,
	HunterOptions_QuiverBonus,
} from '@generated/proto/hunter';
import { SavedTalents } from '@generated/proto/ui';

import DefaultAPL from './apls/default.apl.json';
import P1GearJson from './gear_sets/p1.gear.json';

export const P1Gear = PresetUtils.makePresetGear('P1', P1GearJson);

export const DefaultRotation = PresetUtils.makePresetAPLRotation('APL', DefaultAPL);

export const TurretRotation = Hunter_Rotation.create({
	meleeWeave: false,
	timeToWeave: 400,
	useMulti: false,
	useArcane: true,
});
export const TurretSimple = PresetUtils.makePresetSimpleRotation('Turret', Spec.SpecHunter, TurretRotation);

export const WeaveRotation = Hunter_Rotation.create({
	meleeWeave: true,
	timeToWeave: 400,
	useMulti: false,
	useArcane: true,
});
export const WeaveSimple = PresetUtils.makePresetSimpleRotation('Weave', Spec.SpecHunter, WeaveRotation);

export const BeastMasteryTalents = {
	name: 'Beast Mastery',
	data: SavedTalents.create({
		talentsString: '5320001505101251-005355000101-',
	}),
};

export const DefaultOptions = HunterOptions.create({
	classOptions: {
		ammo: HunterOptions_Ammo.ThoriumHeadedArrow,
		quiverBonus: HunterOptions_QuiverBonus.AncientSinewWrappedLamina,
		petType: PetType.Raptor,
		petUptime: 1,
		petAttackSpeed: HunterOptions_PetAttackSpeed.FasterAttackII,
		cobraReflexes: true,
		petAggression: 5,
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
	insectSwarm: true,
	judgementOfLight: true,
	judgementOfWisdom: true,
	sunderArmor: true,
});

export const DefaultConsumables = ConsumesSpec.create({
	elixirIds: [Consumables.ElixirOfTheMongoose.id, Consumables.GreaterMagebloodElixir.id],
	foodId: Consumables.SwiftstrikeSteak.id,
	potId: Consumables.MajorFrenzyPotion.id,
	conjuredId: Consumables.DemonicRune.id,
	explosiveId: Consumables.EzThroDynamiteII.id,
	extraIds: [Consumables.GoblinSapperCharge.id, Consumables.ScrollOfStrengthIV.id],
});

export const OtherDefaults = {
	distanceFromTarget: 7,
	iterationCount: 25000,
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	race: Race.RaceOrc,
};
