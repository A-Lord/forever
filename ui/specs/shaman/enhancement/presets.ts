import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Profession, Race, Spec, TristateEffect } from '@generated/proto/common';
import { EnhancementShaman_Options as EnhancementShamanOptions, ShamanImbue } from '@generated/proto/shaman';
import { SavedTalents } from '@generated/proto/ui';

import DefaultApl from './apls/default.apl.json';
import TestingBuildJson from './builds/testing.build.json';
import PreraidGearJson from './gear_sets/preraid.gear.json';

export const ROTATION_PRESET_DEFAULT = PresetUtils.makePresetAPLRotation('Default', DefaultApl);

export const PreraidGear = PresetUtils.makePresetGear('Pre-raid', PreraidGearJson);

export const TestingBuild = PresetUtils.makePresetBuildFromJSON('Testing', Spec.SpecEnhancementShaman, TestingBuildJson);

export const StandardTalents = {
	name: 'Standard',
	data: SavedTalents.create({
		talentsString: '0505331-055030031005112251',
	}),
};

export const DefaultIndividualBuffs = IndividualBuffs.create({
	greaterBlessingOfKings: true,
	greaterBlessingOfMight: true,
});

export const DefaultOptions = EnhancementShamanOptions.create({
	classOptions: {
		shieldProcrate: 0,
		imbueMh: ShamanImbue.WindfuryWeapon,
	},
});

export const OtherDefaults = {
	distanceFromTarget: 5,
	profession1: Profession.Engineering,
	profession2: Profession.Leatherworking,
	race: Race.RaceOrc,
};

export const DefaultConsumables = ConsumesSpec.create({
	potId: Consumables.MajorFrenzyPotion.id,
	flaskId: Consumables.FlaskOfNaturalAggression.id,
	conjuredId: Consumables.DemonicRune.id,
	elixirIds: [
		Consumables.ElixirOfTheOwl.id,
		Consumables.ElixirOfSages.id,
		Consumables.GreaterMagebloodElixir.id,
		Consumables.ElixirOfNaturePower.id,
		Consumables.GreaterArcaneElixir.id,
		Consumables.WinterfallFirewater.id,
	],
	explosiveId: Consumables.EzThroDynamiteII.id,
	extraIds: [Consumables.GoblinSapperCharge.id, Consumables.ScrollOfAgilityIV.id, Consumables.ScrollOfStrengthIV.id],
});

export const DefaultPartyBuffs = PartyBuffs.create({
	leaderOfThePack: true,
	battleShout: TristateEffect.TristateEffectRegular,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	prayerOfFortitude: true,
	giftOfTheWild: true,
	arcaneBrilliance: true,
});

export const DefaultDebuffs = Debuffs.create({
	judgementOfWisdom: true,
	giftOfArthas: true,
	exposeArmor: true,
	faerieFire: true,
	sunderArmor: true,
	curseOfElements: true,
	curseOfRecklessness: true,
	huntersMark: true,
});
