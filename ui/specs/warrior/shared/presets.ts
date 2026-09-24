import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec } from '@generated/proto/common';

export const DefaultIndividualBuffs = IndividualBuffs.create({
	greaterBlessingOfKings: true,
	greaterBlessingOfMight: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	graceOfAirTotem: true,
	strengthOfEarthTotem: true,
	windfuryTotem: true,
	leaderOfThePack: true,
	totemTwisting: true,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	prayerOfFortitude: true,
	giftOfTheWild: true,
});

export const DefaultDebuffs = Debuffs.create({
	giftOfArthas: true,
	mangle: true,
	exposeArmor: true,
	faerieFire: true,
	sunderArmor: true,
	curseOfRecklessness: true,
	huntersMark: true,
});

export const DefaultConsumables = ConsumesSpec.create({
	goblinSapper: true,
});
