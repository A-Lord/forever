import * as PresetUtils from '@app/preset_utils';
import { makeSpecChangeWarningToast } from '@features/settings/utils/spec_change_warning_toast';
import { ConsumesSpec, Debuffs, Drums, IndividualBuffs, ItemSlot, PartyBuffs, Profession, RaidBuffs, Spec } from '@generated/proto/common';
import { Mage_Options as MageOptions, Mage_Rotation, MageArmor } from '@generated/proto/mage';
import { SavedTalents } from '@generated/proto/ui';
import { Player } from '@sim/player/player';
import { defaultImprovedShadowBoltSettings } from '@sim/proto/utils';

import ArcaneApl from './apls/arcane.apl.json';
import ArcaneBraidApl from './apls/arcaneBraid.apl.json';
import BlankAPL from './apls/blank.apl.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const BLANK_APL = PresetUtils.makePresetAPLRotation('Blank', BlankAPL);

export const ROTATION_PRESET_ARCANE = PresetUtils.makePresetAPLRotation('Arcane', ArcaneApl);
export const doesNotHaveSerpentCoilBraid = (player: Player<Spec.SpecMage>) =>
	player.getEquippedItem(ItemSlot.ItemSlotTrinket1)?.id != 30720 && player.getEquippedItem(ItemSlot.ItemSlotTrinket2)?.id != 30720;
export const ROTATION_PRESET_ARCANEBRAID = PresetUtils.makePresetAPLRotation('BraidSimple', ArcaneBraidApl, {
	onLoad(player: Player<Spec.SpecMage>) {
		makeSpecChangeWarningToast(
			[
				{
					condition: doesNotHaveSerpentCoilBraid,
					message: 'Check your gear: You do not have Serpent-Coil Braid equipped, but the selected option is for Serpent-Coil Braid.',
				},
			],
			player,
		);
	},
});

export const ArcaneMageSimpleRotation = Mage_Rotation.create({
	conserveStart: 20,
	conserveEnd: 30,
	delayMajorCDs: 10,
});

export const APL_ARCANE_SIMPLE = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecMage, ArcaneMageSimpleRotation);

export const Talents = {
	name: 'Blank',
	data: SavedTalents.create({
		talentsString: '',
	}),
};

export const DefaultOptions = MageOptions.create({
	classOptions: {
		defaultMageArmor: MageArmor.MageArmorMageArmor,
	},
});

export const OtherDefaults = {
	distanceFromTarget: 20,
	profession1: Profession.Engineering,
	profession2: Profession.Tailoring,
};

export const DefaultConsumables = ConsumesSpec.create({
	guardianElixirId: 32067, // Elixir of Draenic Wisdom
	battleElixirId: 28103, // Adept's Elixir
	foodId: 27657, // Blackened Basilisk
	mhImbueId: 20749, // Brilliant Wizard Oil
	potId: 22832, // Super Mana Potion
});

export const DefaultRaidBuffs = RaidBuffs.create({
	bloodlust: true,
	divineSpirit: 2,
	arcaneBrilliance: true,
	giftOfTheWild: 2,
	powerWordFortitude: 2,
	shadowProtection: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	manaSpringTotem: 2,
	manaTideTotems: 1,
	wrathOfAirTotem: 1,
	drums: Drums.LesserDrumsOfBattle,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfWisdom: true,
	innervates: 1,
	blessingOfSalvation: true,
	shadowPriestDps: 1400,
});

export const DefaultDebuffs = Debuffs.create({
	misery: true,
	curseOfElements: 2,
	judgementOfWisdom: true,
	...defaultImprovedShadowBoltSettings(),
});
