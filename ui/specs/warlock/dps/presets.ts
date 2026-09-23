import * as PresetUtils from '@app/preset_utils';
import { ConsumesSpec, Debuffs, Drums, IndividualBuffs, PartyBuffs, Profession, RaidBuffs, TristateEffect } from '@generated/proto/common';
import { Warlock_Options as WarlockOptions, WarlockOptions_Armor, WarlockOptions_CurseOptions, WarlockOptions_Summon } from '@generated/proto/warlock';
import { defaultExposeWeaknessSettings, defaultImprovedShadowBoltSettings, defaultRaidBuffMajorDamageCooldowns } from '@sim/proto/utils';

import AfflictionRot from './apls/affliction.apl.json';
import BlankAPL from './apls/blank.apl.json';
import DemoRot from './apls/demonology.apl.json';
import DestroFireRot from './apls/destro_fire.apl.json';
import DestroRot from './apls/destruction.apl.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const BLANK_APL = PresetUtils.makePresetAPLRotation('Blank', BlankAPL);

// Rotations
export const AfflictionAPL = PresetUtils.makePresetAPLRotation('Affliction', AfflictionRot);
export const DemoAPL = PresetUtils.makePresetAPLRotation('Demonology', DemoRot);
export const DestroAPL = PresetUtils.makePresetAPLRotation('Destruction', DestroRot);
export const DestroFireAPL = PresetUtils.makePresetAPLRotation('Destruction (Fire)', DestroFireRot);

// Defaults
export const DefaultOptions = WarlockOptions.create({
	classOptions: {
		armor: WarlockOptions_Armor.FelArmor,
		curseOptions: WarlockOptions_CurseOptions.Recklessness,
		sacrificeSummon: true,
		summon: WarlockOptions_Summon.Succubus,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: 22866, // Flask of Pure Death
	foodId: 27657, // Blackened Basilisk
	conjuredId: 12662, // Demonic Rune
	mhImbueId: 25122, // Brilliant Wizard Oil
	potId: 22839, // Destruction Potion
	explosiveId: 18588,
	petScrollAgi: true,
	petScrollStr: true,
});

export const OtherDefaults = {
	distanceFromTarget: 20,
	profession1: Profession.Engineering,
	profession2: Profession.Tailoring,
};

export const DefaultRaidBuffs = RaidBuffs.create({
	...defaultRaidBuffMajorDamageCooldowns(),
	arcaneBrilliance: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	divineSpirit: TristateEffect.TristateEffectImproved,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	manaSpringTotem: TristateEffect.TristateEffectRegular,
	moonkinAura: TristateEffect.TristateEffectRegular,
	totemOfWrath: 1,
	wrathOfAirTotem: TristateEffect.TristateEffectImproved,
	eyeOfTheNight: true,
	chainOfTheTwilightOwl: true,
	drums: Drums.LesserDrumsOfBattle,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfWisdom: true,
	shadowPriestDps: 0,
});

export const DefaultDebuffs = Debuffs.create({
	...defaultExposeWeaknessSettings(),
	...defaultImprovedShadowBoltSettings(),
	judgementOfWisdom: true,
	misery: true,
	shadowWeaving: true,
	sunderArmor: true,
	screech: true,
	faerieFire: TristateEffect.TristateEffectImproved,
	curseOfRecklessness: true,
	shadowEmbrace: true,
	curseOfElements: TristateEffect.TristateEffectImproved,
	bloodFrenzy: true,
	giftOfArthas: true,
	mangle: true,
	exposeArmor: TristateEffect.TristateEffectImproved,
	huntersMark: TristateEffect.TristateEffectImproved,
});
