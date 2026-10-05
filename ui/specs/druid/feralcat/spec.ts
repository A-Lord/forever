import * as BuffDebuffInputs from '@features/settings/model/buffs_debuffs';
import * as OtherInputs from '@features/settings/model/other_inputs';
import { APLAction, APLListItem, APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Cooldowns, ItemSlot, PseudoStat, Spec, Stat } from '@generated/proto/common';
import { FeralCatDruid_Rotation as DruidRotation } from '@generated/proto/druid';
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
	knownIssues: [
		"Rake uses rank 1's measured 0.055 AP per tick at higher ranks and reads current AP each tick; higher ranks and snapshotting need in-game verification.",
		"Rip/Bite AP scaling, Omen's proc rate, base crit and Furor state retention use stated assumptions where the game files are silent.",
	],
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
		gear: Presets.RAID_GEAR.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Presets.RAID_EP.epWeights,
		statCaps: (() => {
			return new Stats().withPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, 8).withPseudoStat(PseudoStat.PseudoStatExpertisePercent, 6.5);
		})(),
		other: Presets.OtherDefaults,
		// Default consumes settings.
		consumables: Presets.DefaultConsumables,
		// Default rotation settings.
		rotationType: APLRotationType.TypeAuto,
		simpleRotation: Presets.DefaultRotation,
		talents: Presets.STARTER_TALENTS.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		// Default raid/party buffs settings.
		raidBuffs: Presets.DefaultRaidBuffs,
		partyBuffs: Presets.DefaultPartyBuffs,
		individualBuffs: Presets.DefaultIndividualBuffs,
		debuffs: Presets.DefaultDebuffs,
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

	defaultBuild: Presets.RAID_BUILD,

	presets: {
		builds: [Presets.RAID_BUILD, Presets.RAID_DPS_BUILD, Presets.STARTER_BUILD],
		epWeights: [Presets.RAID_EP],
		// Preset talents that the user can quickly select.
		talents: [Presets.STARTER_TALENTS],
		rotations: [Presets.SIMPLE, Presets.POWERSHIFT_APL, Presets.RAID_DPS_APL, Presets.APL],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.RAID_GEAR, Presets.STARTER_GEAR],
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

// Major cooldowns fire ahead of the Cat action; all selected rotation options reach that action.
function catPowershiftRotation(simple: DruidRotation): APLRotation {
	const doRotation = APLAction.create({ action: { oneofKind: 'catOptimalRotationAction', catOptimalRotationAction: { ...simple } } });

	return APLRotation.create({
		priorityList: [
			APLListItem.create({ action: APLAction.create({ action: { oneofKind: 'autocastOtherCooldowns', autocastOtherCooldowns: {} } }) }),
			APLListItem.create({ action: doRotation }),
		],
	});
}
