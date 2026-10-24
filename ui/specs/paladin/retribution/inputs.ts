import * as InputHelpers from '@features/inputs/input_helpers';
import { Spec } from '@generated/proto/common';
import { PaladinAura, PaladinSeal } from '@generated/proto/paladin';
import i18n from '@i18n/config';
import { Player } from '@sim/player/player';

// One yard outside Judgement's 10 yd range. With an assigned seal, the Twist of Light rotation's walk in
// waits for Judgement to come into range, and a wait whose condition already holds never ends, so a start
// inside that range never reaches melee.
export const MIN_PULL_RANGE = 11;

export const RotationInputs = {
	inputs: [
		InputHelpers.makeRotationEnumInput<Spec.SpecRetributionPaladin, PaladinAura>({
			fieldName: 'aura',
			label: i18n.t('rotation_tab.options.paladin.retribution.aura.label'),
			labelTooltip: i18n.t('rotation_tab.options.paladin.retribution.aura.tooltip'),
			values: [
				{ name: i18n.t('rotation_tab.options.paladin.retribution.aura.values.retribution'), value: PaladinAura.RetributionAura },
				{ name: i18n.t('rotation_tab.options.paladin.retribution.aura.values.devotion'), value: PaladinAura.DevotionAura },
				{ name: i18n.t('rotation_tab.options.paladin.retribution.aura.values.concentration'), value: PaladinAura.ConcentrationAura },
				{ name: i18n.t('rotation_tab.options.paladin.retribution.aura.values.fire_resistance'), value: PaladinAura.FireResistanceAura },
				{ name: i18n.t('rotation_tab.options.paladin.retribution.aura.values.frost_resistance'), value: PaladinAura.FrostResistanceAura },
				{ name: i18n.t('rotation_tab.options.paladin.retribution.aura.values.shadow_resistance'), value: PaladinAura.ShadowResistanceAura },
			],
		}),
		InputHelpers.makeRotationEnumInput<Spec.SpecRetributionPaladin, PaladinSeal>({
			fieldName: 'assignedSeal',
			label: i18n.t('rotation_tab.options.paladin.retribution.assigned_seal.label'),
			labelTooltip: i18n.t('rotation_tab.options.paladin.retribution.assigned_seal.tooltip'),
			values: [
				{ name: i18n.t('rotation_tab.options.paladin.retribution.assigned_seal.values.none'), value: PaladinSeal.SealNone },
				{ name: i18n.t('rotation_tab.options.paladin.retribution.assigned_seal.values.crusader'), value: PaladinSeal.SealOfTheCrusader },
				{ name: i18n.t('rotation_tab.options.paladin.retribution.assigned_seal.values.light'), value: PaladinSeal.SealOfLight },
				{ name: i18n.t('rotation_tab.options.paladin.retribution.assigned_seal.values.wisdom'), value: PaladinSeal.SealOfWisdom },
				{ name: i18n.t('rotation_tab.options.paladin.retribution.assigned_seal.values.justice'), value: PaladinSeal.SealOfJustice },
			],
		}),
		InputHelpers.makeRotationNumberInput<Spec.SpecRetributionPaladin>({
			fieldName: 'pullRange',
			label: i18n.t('rotation_tab.options.paladin.retribution.pull_range.label'),
			labelTooltip: i18n.t('rotation_tab.options.paladin.retribution.pull_range.tooltip', { min: MIN_PULL_RANGE }),
			positive: true,
			setValue: (player: Player<Spec.SpecRetributionPaladin>, newValue: number) => {
				const rotation = player.getSimpleRotation();
				rotation.pullRange = Math.max(MIN_PULL_RANGE, newValue);
				player.setSimpleRotation(rotation);
			},
		}),
		InputHelpers.makeRotationNumberInput<Spec.SpecRetributionPaladin>({
			fieldName: 'delayMajorCds',
			label: i18n.t('rotation_tab.options.paladin.retribution.delay_major_cds.label'),
			labelTooltip: i18n.t('rotation_tab.options.paladin.retribution.delay_major_cds.tooltip'),
			positive: true,
		}),
	],
};
