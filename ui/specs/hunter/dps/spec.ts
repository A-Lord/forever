import * as other_inputs from '@features/settings/model/other_inputs';
import { StatCapType } from '@generated/proto/api';
import { APLListItem, APLRotation, APLRotation_Type, APLValueVariable } from '@generated/proto/apl';
import { Cooldowns, EquipmentSpec, HandType, ItemSlot, PseudoStat, Spec, Stat } from '@generated/proto/common';
import { SavedTalents } from '@generated/proto/ui';
import { PlayerClasses } from '@sim/player/classes';
import { Player } from '@sim/player/player';
import * as AplUtils from '@sim/proto/apl_utils';
import { SpecRotation } from '@sim/proto/spec_types';
import { StatCap, Stats, UnitStat } from '@sim/proto/stats';
import { defineSpec } from '@sim/spec_config';

import * as HunterInputs from './inputs';
import * as Presets from './presets';

export default defineSpec<Spec.SpecHunter>({
	spec: Spec.SpecHunter,

	className: 'hunter-sim-ui',
	cssScheme: PlayerClasses.getCssScheme(PlayerClasses.Hunter),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],
	warnings: [],
	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatAgility,
		Stat.StatStrength,
		Stat.StatIntellect,
		Stat.StatMP5,
		Stat.StatAttackPower,
		Stat.StatRangedAttackPower,
		Stat.StatArmorPenetration,
		Stat.StatMeleeHitRating,
		Stat.StatMeleeHasteRating,
		Stat.StatMeleeCritRating,
		Stat.StatExpertiseRating,
		Stat.StatPhysicalDamage,
	],
	gemStats: [Stat.StatStamina, Stat.StatAgility],
	epPseudoStats: [PseudoStat.PseudoStatRangedHitPercent, PseudoStat.PseudoStatRangedCritPercent, PseudoStat.PseudoStatRangedDps],
	consumableStats: [Stat.StatStamina, Stat.StatHealth, Stat.StatMana],
	// Reference stat against which to calculate EP.
	epReferenceStat: Stat.StatAgility,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[
			Stat.StatHealth,
			Stat.StatMana,
			Stat.StatStamina,
			Stat.StatStrength,
			Stat.StatAgility,
			Stat.StatIntellect,
			Stat.StatMP5,
			Stat.StatAttackPower,
			Stat.StatRangedAttackPower,
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
			PseudoStat.PseudoStatRangedHitPercent,
			PseudoStat.PseudoStatRangedCritPercent,
			PseudoStat.PseudoStatRangedHastePercent,
			PseudoStat.PseudoStatExpertisePercent,
		],
	),
	itemSwapSlots: [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, ItemSlot.ItemSlotRanged, ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2],
	defaults: {
		// Default equipped gear.
		gear: EquipmentSpec.create(),
		// Default EP weights for sorting gear in the gear picker.
		epWeights: new Stats(),
		softCapBreakpoints: [
			StatCap.fromPseudoStat(PseudoStat.PseudoStatRangedHitPercent, {
				breakpoints: [9],
				capType: StatCapType.TypeSoftCap,
				postCapEPs: [0],
			}),
		],
		rotationType: APLRotation_Type.TypeSimple,
		simpleRotation: Presets.WeaveRotation,
		other: Presets.OtherDefaults,
		// Default consumes settings.
		consumables: Presets.DefaultConsumables,
		// Default talents.
		talents: SavedTalents.create(),
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		// Default raid/party buffs settings.
		raidBuffs: Presets.DefaultRaidBuffs,
		partyBuffs: Presets.DefaultPartyBuffs,
		individualBuffs: Presets.DefaultIndividualBuffs,
		debuffs: Presets.DefaultDebuffs,
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [HunterInputs.PetTypeInput(), HunterInputs.QuiverInput(), HunterInputs.AmmoInput()],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [Stat.StatSpirit, Stat.StatSpellCritRating, Stat.StatSpellDamage],
	excludeBuffDebuffInputs: [],
	rotationInputs: HunterInputs.RotationInputs,
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [
			other_inputs.TotemTwisting,
			HunterInputs.PetUptime(),
			HunterInputs.PetSingleAbility(),
			other_inputs.InputDelay,
			other_inputs.DistanceFromTarget,
			other_inputs.TankAssignment,
		],
	},
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		epWeights: [],
		// Preset talents that the user can quickly select.
		talents: [],
		// Preset rotations that the user can quickly select.
		rotations: [Presets.WeaveSimple, Presets.TurretSimple, Presets.DefaultRotation],
		// Preset gear configurations that the user can quickly select.
		gear: [],
	},

	autoRotation: (player: Player<Spec.SpecHunter>): APLRotation => {
		const rotation = APLRotation.clone(Presets.DefaultRotation.rotation.rotation!);
		const gear = player.getGear();
		const mainHandType = gear.getEquippedItem(ItemSlot.ItemSlotMainHand)?.item.handType;
		if (mainHandType !== HandType.HandTypeTwoHand) {
			rotation.valueVariables[2] = APLValueVariable.fromJson({
				name: 'Melee weave',
				value: { const: { val: 'false' } },
			});
		}
		return rotation;
	},

	simpleRotation: (player: Player<Spec.SpecHunter>, simple: SpecRotation<Spec.SpecHunter>, cooldowns: Cooldowns): APLRotation => {
		const actions = AplUtils.simpleCooldownActions(cooldowns);
		const rotation = APLRotation.clone(Presets.DefaultRotation.rotation.rotation!);

		const {
			viperStartManaPercent = 0.05,
			viperStopManaPercent = 0.25,
			meleeWeave = player.getEquippedItem(ItemSlot.ItemSlotMainHand)?.item.handType === HandType.HandTypeTwoHand,
			useMulti = true,
			useArcane = true,
			timeToWeave = 400,
		} = simple;

		const viperStartManaPercentValue = APLValueVariable.fromJson({
			name: 'Viper start',
			value: { const: { val: `${viperStartManaPercent * 100}%` } },
		});

		const viperStopManaPercentValue = APLValueVariable.fromJson({
			name: 'Viper stop',
			value: { const: { val: `${viperStopManaPercent * 100}%` } },
		});

		const meleeWeaveValue = APLValueVariable.fromJson({
			name: 'Melee weave',
			value: { const: { val: String(meleeWeave) } },
		});

		const useMultiValue = APLValueVariable.fromJson({
			name: 'Use Multi-Shot',
			value: { const: { val: String(useMulti) } },
		});

		const useArcaneValue = APLValueVariable.fromJson({
			name: 'Use Arcane Shot',
			value: { const: { val: String(useArcane) } },
		});

		const timeToWeaveValue = APLValueVariable.fromJson({
			name: 'Time to weave',
			value: { const: { val: `${timeToWeave}ms` } },
		});

		const overrides: Record<string, APLValueVariable> = {
			'Viper start': viperStartManaPercentValue,
			'Viper stop': viperStopManaPercentValue,
			'Melee weave': meleeWeaveValue,
			'Time to weave': timeToWeaveValue,
			'Use Multi-Shot': useMultiValue,
			'Use Arcane Shot': useArcaneValue,
		};
		rotation.valueVariables = rotation.valueVariables.map(v => overrides[v.name] ?? v);

		return APLRotation.create({
			prepullActions: rotation.prepullActions,
			priorityList: [
				...actions.map(action =>
					APLListItem.create({
						action: action,
					}),
				),
				...rotation.priorityList,
			],
			groups: rotation.groups,
			valueVariables: rotation.valueVariables,
		});
	},

	reforge: {},
});
