import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { Class, ConsumesSpec, Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs, TristateEffect } from '@generated/proto/common';
import { defaultExposeWeaknessSettings, defaultRaidBuffMajorDamageCooldowns } from '@sim/proto/utils';

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfMight: true,
	unleashedRage: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	ferociousInspiration: 2,
	braidedEterniumChain: true,
	graceOfAirTotem: TristateEffect.TristateEffectImproved,
	strengthOfEarthTotem: TristateEffect.TristateEffectImproved,
	windfuryTotem: TristateEffect.TristateEffectImproved,
	leaderOfThePack: TristateEffect.TristateEffectRegular,
	totemTwisting: true,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	...defaultRaidBuffMajorDamageCooldowns(Class.ClassWarrior),
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
});

export const DefaultDebuffs = Debuffs.create({
	...defaultExposeWeaknessSettings(),
	misery: true,
	bloodFrenzy: true,
	giftOfArthas: true,
	mangle: true,
	exposeArmor: TristateEffect.TristateEffectImproved,
	faerieFire: TristateEffect.TristateEffectImproved,
	sunderArmor: true,
	curseOfRecklessness: true,
	huntersMark: TristateEffect.TristateEffectImproved,
});

export const DefaultConsumables = ConsumesSpec.create({
	potId: Consumables.MajorFrenzyPotion.id,
	flaskId: Consumables.FlaskOfNaturalAggression.id,
	foodId: Consumables.ProwlerSteak.id,
	explosiveId: Consumables.EzThroDynamiteII.id,
	ohImbueId: Consumables.DenseSharpeningStone.id,
	extraIds: [Consumables.GoblinSapperCharge.id, Consumables.ScrollOfAgilityIV.id, Consumables.ScrollOfStrengthIV.id],
});
