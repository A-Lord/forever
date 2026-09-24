import * as OtherInputs from '@features/settings/model/other_inputs';
import { StatCapType } from '@generated/proto/api';
import { APLRotation } from '@generated/proto/apl';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { EquipmentSpec, ItemSlot, PseudoStat, Spec, Stat, TristateEffect } from '@generated/proto/common';
import { SavedTalents } from '@generated/proto/ui';
import * as Mechanics from '@sim/constants/mechanics';
import { PlayerClasses } from '@sim/player/classes';
import { Player } from '@sim/player/player';
import { StatCap, Stats, UnitStat } from '@sim/proto/stats';
import { defineSpec } from '@sim/spec_config';

import * as Presets from './presets';

export default defineSpec<Spec.SpecRogue>({
	spec: Spec.SpecRogue,

	className: 'rogue-sim-ui',
	cssScheme: PlayerClasses.getCssScheme(PlayerClasses.Rogue),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [
		'The APL is in constant flux due to bug fixes and new findings; if your DPS drops dramatically, reset it back to "Auto" in the Rotation tab!',
		'Mutilate does not have a default APL currently. It will not be automatically used when talented.',
	],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatStamina,
		Stat.StatAgility,
		Stat.StatStrength,
		Stat.StatMeleeCritRating,
		Stat.StatMeleeHasteRating,
		Stat.StatMeleeHitRating,
		Stat.StatArmorPenetration,
		Stat.StatExpertiseRating,
		Stat.StatAttackPower,
		Stat.StatPhysicalDamage,
	],
	epPseudoStats: [PseudoStat.PseudoStatMainHandDps, PseudoStat.PseudoStatOffHandDps],
	// Reference stat against which to calculate EP.
	epReferenceStat: Stat.StatAttackPower,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[
			Stat.StatHealth,
			Stat.StatStamina,
			Stat.StatAgility,
			Stat.StatStrength,
			Stat.StatAttackPower,
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
		gear: EquipmentSpec.create(),
		// Default EP weights for sorting gear in the gear picker.
		epWeights: new Stats(),
		statCaps: (() => {
			const expCap = new Stats().withPseudoStat(PseudoStat.PseudoStatExpertisePercent, 6.5);
			return expCap;
		})(),
		softCapBreakpoints: (() => {
			const meleeHitSoftCapConfig = StatCap.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, {
				breakpoints: [9, 28],
				capType: StatCapType.TypeSoftCap,
				postCapEPs: [3.06 * Mechanics.PHYSICAL_HIT_RATING_PER_HIT_PERCENT, 0],
			});

			return [meleeHitSoftCapConfig];
		})(),
		other: Presets.OtherDefaults,
		// Default consumes settings.
		consumables: Presets.DefaultConsumables,
		// Default talents.
		talents: SavedTalents.create(),
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		// Default raid/party buffs settings.
		raidBuffs: RaidBuffs.create({
			giftOfTheWild: true,
		}),
		partyBuffs: PartyBuffs.create({
			battleShout: TristateEffect.TristateEffectRegular,
			strengthOfEarthTotem: true,
			graceOfAirTotem: true,
			windfuryTotem: true,
			leaderOfThePack: true,
			totemTwisting: true,
		}),
		individualBuffs: IndividualBuffs.create({
			greaterBlessingOfKings: true,
			greaterBlessingOfMight: true,
		}),
		debuffs: Debuffs.create({
			huntersMark: true,
			mangle: true,
			curseOfRecklessness: true,
			faerieFire: true,
			giftOfArthas: true,
			sunderArmor: true,
		}),
	},

	playerInputs: {
		inputs: [],
	},
	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [Stat.StatSpellHitRating],
	excludeBuffDebuffInputs: [],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [OtherInputs.TotemTwisting, OtherInputs.InFrontOfTarget, OtherInputs.InputDelay],
	},
	itemSwapSlots: [ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2, ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		epWeights: [],
		// Preset talents that the user can quickly select.
		talents: [],
		// Preset rotations that the user can quickly select.
		rotations: [Presets.SINSITER_APL],
		// Preset gear configurations that the user can quickly select.
		gear: [],
	},

	autoRotation: (_player: Player<Spec.SpecRogue>): APLRotation => {
		return Presets.SINSITER_APL.rotation.rotation!;
	},

	reforge: {},
});
