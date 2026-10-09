import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Profession, Race, TristateEffect } from '@generated/proto/common';
import { DpsPriest_Options as Options, PriestOptions_Armor } from '@generated/proto/priest';
import { SavedTalents } from '@generated/proto/ui';

import DefaultApl from './apls/default.apl.json';
import SmiteApl from './apls/smite.apl.json';
import PreraidGearJson from './gear_sets/preraid.gear.json';
import PreraidSmiteGear from './gear_sets/preraid-smite.gear.json';

export const ROTATION_PRESET_DEFAULT = PresetUtils.makePresetAPLRotation('Default', DefaultApl);
export const ROTATION_SMITE_DEFAULT = PresetUtils.makePresetAPLRotation('Smite', SmiteApl);

export const PreraidGear = PresetUtils.makePresetGear('Pre-raid', PreraidGearJson);
export const PreraidSmite = PresetUtils.makePresetGear('Smite Pre-raid', PreraidSmiteGear);

export const StandardTalents = PresetUtils.makePresetTalents('Standard', SavedTalents.create({ talentsString: '005300211303--500320501201312251' }));
export const SmiteTalents = PresetUtils.makePresetTalents('Smite', SavedTalents.create({ talentsString: '525230031305001001-30515100002-3' }));

export const DefaultOptions = Options.create({
	classOptions: {
		armor: PriestOptions_Armor.InnerFire,
		preShadowform: true,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: Consumables.FlaskOfSupremePower.id,
	foodId: Consumables.NightfinSoup.id,
	conjuredId: Consumables.DemonicRune.id,
	mhImbueId: Consumables.BrilliantWizardOil.id,
	potId: Consumables.MajorSpellblastingPotion.id,
	explosiveId: Consumables.EzThroDynamiteII.id,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	arcaneBrilliance: true,
	giftOfTheWild: true,
	prayerOfFortitude: true,
	prayerOfSpirit: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	manaSpringTotem: TristateEffect.TristateEffectRegular,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	greaterBlessingOfKings: true,
	greaterBlessingOfWisdom: true,
});

export const DefaultDebuffs = Debuffs.create({
	judgementOfWisdom: true,
	faerieFire: true,
	curseOfElements: true,
	exposeArmor: true,
});

export const OtherDefaults = {
	channelClipDelay: 100,
	distanceFromTarget: 28,
	profession1: Profession.Enchanting,
	profession2: Profession.Tailoring,
	race: Race.RaceTroll,
};
