import * as BuffDebuffInputs from '@features/settings/model/buffs_debuffs';
import * as OtherInputs from '@features/settings/model/other_inputs';
import { APLAction, APLListItem, APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { Cooldowns, ItemSlot, PseudoStat, Spec, Stat, TristateEffect } from '@generated/proto/common';
import { FeralCatDruid_Rotation as DruidRotation } from '@generated/proto/druid';
import { SavedTalents } from '@generated/proto/ui';
import { PlayerClasses } from '@sim/player/classes';
import { Player } from '@sim/player/player';
import { Stats, UnitStat } from '@sim/proto/stats';
import { defineSpec } from '@sim/spec_config';

import * as FeralInputs from './inputs';
import * as Presets from './presets';

export default defineSpec<Spec.SpecFeralCatDruid>({
	spec: Spec.SpecFeralCatDruid,

	className: 'feral-druid-sim-ui',
	cssScheme: PlayerClasses.getCssScheme(PlayerClasses.Druid),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],
	warnings: [],

	// Extra stats shown in consumables picker (beyond epStats).
	consumableStats: [Stat.StatHasteRating, Stat.StatMana, Stat.StatSpirit, Stat.StatMP5],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatStrength,
		Stat.StatAgility,
		Stat.StatAttackPower,
		Stat.StatFeralAttackPower,
		Stat.StatPhysicalDamage,
		Stat.StatHitRating,
		Stat.StatExpertiseRating,
		Stat.StatCritRating,
		Stat.StatHasteRating,
		Stat.StatArmorPenetration,
	],
	epPseudoStats: [],
	// Reference stat against which to calculate EP. I think all classes use either spell power or attack power.
	epReferenceStat: Stat.StatAgility,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[
			Stat.StatHealth,
			Stat.StatStrength,
			Stat.StatAgility,
			Stat.StatStamina,
			Stat.StatIntellect,
			Stat.StatSpirit,
			Stat.StatAttackPower,
			Stat.StatMana,
			Stat.StatArmorPenetration,
			Stat.StatArcaneResistance,
			Stat.StatFireResistance,
			Stat.StatFrostResistance,
			Stat.StatNatureResistance,
			Stat.StatShadowResistance,
		],
		[
			PseudoStat.PseudoStatMeleeHitPercent,
			PseudoStat.PseudoStatMeleeCritPercent,
			PseudoStat.PseudoStatMeleeHastePercent,
			PseudoStat.PseudoStatExpertisePercent,
		],
	),

	defaults: {
		// Default equipped gear.
		gear: Presets.Level60Gear.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Presets.Level60Ep,
		statCaps: (() => {
			return new Stats().withPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, 8).withPseudoStat(PseudoStat.PseudoStatExpertisePercent, 6.5);
		})(),
		other: Presets.OtherDefaults,
		// Default consumes settings.
		consumables: Presets.DefaultConsumables,
		// Default rotation settings.
		rotationType: APLRotationType.TypeAuto,
		simpleRotation: Presets.DefaultRotation,
		// Genesis, nature's majesty, ferocity, heart of the wild, shredding attacks,
		// savage fury, sharpened claws, shifting power, predatory strikes, blood frenzy,
		// improved shifting power, leader of the pack, predatory instincts, rend and tear,
		// berserk, furor, naturalist. Shifting Power is absent until that talent is taken.
		talents: SavedTalents.create({
			talentsString: '05002-55000032021032212051-055',
		}),
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		// Default raid/party buffs settings.
		raidBuffs: RaidBuffs.create({
			arcaneBrilliance: true,
			prayerOfSpirit: true,
			giftOfTheWild: true,
			prayerOfFortitude: true,
			prayerOfShadowProtection: true,
		}),
		partyBuffs: PartyBuffs.create({
			battleShout: TristateEffect.TristateEffectRegular,
			windfuryTotem: true,
			manaSpringTotem: TristateEffect.TristateEffectRegular,
			strengthOfEarthTotem: true,
		}),
		individualBuffs: IndividualBuffs.create({
			greaterBlessingOfKings: true,
			greaterBlessingOfMight: true,
		}),
		debuffs: Debuffs.create({
			exposeArmor: true,
			judgementOfWisdom: true,
			curseOfRecklessness: true,
			faerieFire: true,
			sunderArmor: true,
			giftOfArthas: true,
		}),
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [],
	// Inputs to include in the 'Rotation' section on the settings tab.
	rotationInputs: FeralInputs.FeralDruidRotationConfig,
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [Stat.StatMP5, Stat.StatIntellect, Stat.StatStamina],
	excludeBuffDebuffInputs: [BuffDebuffInputs.AtieshMage],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [
			OtherInputs.InputDelay,
			OtherInputs.DistanceFromTarget,
			OtherInputs.TankAssignment,
			OtherInputs.InFrontOfTarget,
			FeralInputs.CannotShredTarget,
		],
	},
	itemSwapSlots: [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: true,
	},

	presets: {
		epWeights: [Presets.Level60EpWeights],
		// Preset talents that the user can quickly select.
		talents: [],
		rotations: [Presets.SIMPLE, Presets.APL],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.Level60Gear],
	},

	autoRotation: (player: Player<Spec.SpecFeralCatDruid>): APLRotation => {
		return catPowershiftRotation(player.getSimpleRotation());
	},

	simpleRotation: (_player: Player<Spec.SpecFeralCatDruid>, simple: DruidRotation, _cooldowns: Cooldowns): APLRotation => {
		return catPowershiftRotation(simple);
	},

	hiddenMCDs: [],

	reforge: {},
});

// Major cooldowns (Berserk, trinkets, potions) fire on cooldown ahead of the cat rotation; core refuses
// one that would leave cat form. The rotation keeps Rake up, finishes with Rip, and casts Shifting
// Power when the builder is unaffordable.
function catPowershiftRotation(simple: DruidRotation): APLRotation {
	const doRotation = APLAction.fromJsonString(
		`{"catOptimalRotationAction":{"finishingMove":${simple.finishingMove},"biteweave":${simple.biteweave},"ripweave":${simple.ripweave},"ripMinComboPoints":${simple.ripMinComboPoints || 5},"biteMinComboPoints":${simple.biteMinComboPoints || 5},"rakeTrick":${simple.rakeTrick},"maintainFaerieFire":${simple.maintainFaerieFire}}}`,
	);

	return APLRotation.create({
		priorityList: [APLListItem.create({ action: APLAction.fromJsonString(`{"autocastOtherCooldowns":{}}`) }), APLListItem.create({ action: doRotation })],
	});
}
