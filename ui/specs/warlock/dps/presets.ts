import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Profession, TristateEffect } from '@generated/proto/common';
import { Warlock_Options as WarlockOptions, WarlockOptions_Armor, WarlockOptions_CurseOptions, WarlockOptions_Summon } from '@generated/proto/warlock';

import AffWrackApl from './apls/aff_wrack.apl.json';
import DefaultApl from './apls/default.apl.json';
import DemoFireBrandApl from './apls/demo_fire_brand.apl.json';
import DemoShadowBrandApl from './apls/demo_shadow_brand.apl.json';
import DestroApl from './apls/destro.apl.json';
import SmRuinApl from './apls/sm_ruin.apl.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const DEFAULT_APL = PresetUtils.makePresetAPLRotation('Default', DefaultApl);
export const DEMO_FIRE_BRAND_APL = PresetUtils.makePresetAPLRotation('Demo (Fire Brand)', DemoFireBrandApl);
export const DEMO_SHADOW_BRAND_APL = PresetUtils.makePresetAPLRotation('Demo (Shadow Brand)', DemoShadowBrandApl);
export const DESTRO_APL = PresetUtils.makePresetAPLRotation('Destro', DestroApl);
export const SM_RUIN_APL = PresetUtils.makePresetAPLRotation('SM/Ruin', SmRuinApl);
export const AFF_WRACK_APL = PresetUtils.makePresetAPLRotation('Aff (Wrack)', AffWrackApl);

// Defaults
export const DefaultOptions = WarlockOptions.create({
	classOptions: {
		armor: WarlockOptions_Armor.DemonArmor,
		curseOptions: WarlockOptions_CurseOptions.Recklessness,
		sacrificedDemon: WarlockOptions_Summon.Succubus,
		summon: WarlockOptions_Summon.Imp,
		// A full Felcloth Bag (21341), the largest Soul Bag a warlock can carry.
		soulShards: 32,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: Consumables.FlaskOfSupremePower.id,
	foodId: Consumables.NightfinSoup.id,
	conjuredId: Consumables.DemonicRune.id,
	mhImbueId: Consumables.BrilliantWizardOil.id,
	offensivePotId: Consumables.MajorSpellblastingPotion.id,
	explosiveId: Consumables.EzThroDynamiteII.id,
});

export const OtherDefaults = {
	distanceFromTarget: 20,
	profession1: Profession.Engineering,
	profession2: Profession.Tailoring,
};

export const DefaultRaidBuffs = RaidBuffs.create({
	arcaneBrilliance: true,
	giftOfTheWild: true,
	prayerOfFortitude: true,
	prayerOfSpirit: true,
	moonkinAura: true,
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
	sunderArmor: true,
	faerieFire: true,
	curseOfRecklessness: true,
	curseOfElements: true,
	giftOfArthas: true,
	exposeArmor: true,
});
