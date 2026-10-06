import * as BuffDebuffInputs from '@features/settings/model/buffs_debuffs';
import * as OtherInputs from '@features/settings/model/other_inputs';
import { StatCapType } from '@generated/proto/api';
import { APLRotation } from '@generated/proto/apl';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { EquipmentSpec, ItemSlot, PseudoStat, Spec, Stat, TristateEffect } from '@generated/proto/common';
import { SavedTalents } from '@generated/proto/ui';
import * as Mechanics from '@sim/constants/mechanics';
import { PlayerClasses } from '@sim/player/classes';
import { Player } from '@sim/player/player';
import { WeaponSkillGroup } from '@sim/player/weapon_skill_groups';
import { StatCap, Stats, UnitStat } from '@sim/proto/stats';
import { defineSpec } from '@sim/spec_config';

import { MainHandPoison, OffHandPoison } from './inputs';
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
		Stat.StatCritRating,
		Stat.StatHasteRating,
		Stat.StatHitRating,
		Stat.StatExpertiseRating,
		Stat.StatArmorPenetration,
		Stat.StatSpellPiercing,
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
			Stat.StatSpellPiercing,
			Stat.StatArcaneResistance,
			Stat.StatFireResistance,
			Stat.StatFrostResistance,
			Stat.StatNatureResistance,
			Stat.StatShadowResistance,
		],
		[
			PseudoStat.PseudoStatMeleeHitPercent,
			PseudoStat.PseudoStatMeleeCritPercent,
			PseudoStat.PseudoStatOffHandHitPercent,
			PseudoStat.PseudoStatOffHandCritPercent,
			PseudoStat.PseudoStatMeleeHastePercent,
			PseudoStat.PseudoStatExpertisePercent,
			PseudoStat.PseudoStatSpellHitPercent,
			PseudoStat.PseudoStatSpellCritPercent,
		],
	),
	displayWeaponSkillGroups: [WeaponSkillGroup.Melee],

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
				breakpoints: [8, 27],
				capType: StatCapType.TypeSoftCap,
				postCapEPs: [3.06 * Mechanics.HIT_RATING_PER_HIT_PERCENT, 0],
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
			windfuryTotem: true,
			leaderOfThePack: true,
		}),
		individualBuffs: IndividualBuffs.create({
			greaterBlessingOfKings: true,
			greaterBlessingOfMight: true,
		}),
		debuffs: Debuffs.create({
			curseOfRecklessness: true,
			faerieFire: true,
			giftOfArthas: true,
			sunderArmor: true,
			curseOfElements: true,
		}),
	},

	playerInputs: {
		inputs: [],
	},
	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [MainHandPoison(), OffHandPoison()],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [Stat.StatHitRating, Stat.StatSpellDamage, Stat.StatSpellPiercing],
	excludeBuffDebuffInputs: [BuffDebuffInputs.AtieshMage],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [OtherInputs.InFrontOfTarget, OtherInputs.InputDelay],
	},
	itemSwapSlots: [ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2, ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: true,
	},

	presets: {
		epWeights: [],
		// Preset talents that the user can quickly select.
		talents: [
			Presets.TALENTS_MUTILATE_ASN,
			Presets.TALENTS_MUTILATE_COM,
			Presets.TALENTS_COMBAT_SS,
			Presets.TALENTS_COMBAT_BS,
			Presets.TALENTS_SUB_HEMO,
			Presets.TALENTS_SUB_BS,
		],
		// Preset rotations that the user can quickly select.
		rotations: [Presets.DEFAULT_APL, Presets.BACKSTAB_APL],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.P1_BS_COMBAT_GEAR, Presets.P1_MUTI_GEAR, Presets.P1_SLOW_MH_GEAR, Presets.P1_BS_SUB_GEAR],
		builds: [
			Presets.PRESET_BUILD_MUTILATE_ASN,
			Presets.PRESET_BUILD_MUTILATE_COM,
			Presets.PRESET_BUILD_COMBAT_SS,
			Presets.PRESET_BUILD_COMBAT_BS,
			Presets.PRESET_BUILD_SUB_HEMO,
			Presets.PRESET_BUILD_SUB_BS,
		],
	},

	autoRotation: (_player: Player<Spec.SpecRogue>): APLRotation => {
		return Presets.DEFAULT_APL.rotation.rotation!;
	},

	// reforge: {},
});
