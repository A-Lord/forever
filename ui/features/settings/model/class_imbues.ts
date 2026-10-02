import { Class, ItemSlot, ItemType } from '@generated/proto/common';
import { ShamanImbue } from '@generated/proto/shaman';
import type { Player } from '@sim/player/player';
import { ActionId } from '@sim/proto/action_id';
import type { ShamanSpecs } from '@sim/proto/spec_types';

// A class's own weapon imbues, offered in the consumables imbue picker of a hand beside the oils and
// stones. A weapon holds one temporary enchant, so the two are one pick: choosing either clears the
// other. The class's pick lives in its own options, where the sim reads it, and the picker keys it
// below zero so it never meets a consumable's item id.
export interface ClassImbueHand {
	values: Array<{ value: number; actionId: ActionId; showWhen: (player: Player<any>) => boolean }>;
	// The hand's class imbue as the picker keys it, or 0 for none.
	getPick: (player: Player<any>) => number;
	setPick: (player: Player<any>, value: number) => void;
}

const SHAMAN_IMBUES: Array<[ShamanImbue, number]> = [
	[ShamanImbue.WindfuryWeapon, 16362],
	[ShamanImbue.FlametongueWeapon, 16342],
	[ShamanImbue.FrostbrandWeapon, 16356],
	[ShamanImbue.RockbiterWeapon, 16316],
];

const pickOf = (imbue: ShamanImbue | undefined): number => (imbue ? -imbue : 0);
const imbueOf = (pick: number): ShamanImbue => (pick < 0 ? -pick : ShamanImbue.NoImbue);

const isShaman = (player: Player<any>) => player.getClass() === Class.ClassShaman;
// The check comes first, so the cast names what it established.
const asShaman = (player: Player<any>) => player as Player<ShamanSpecs>;

const holdsWeapon = (player: Player<any>): boolean => player.getEquippedItem(ItemSlot.ItemSlotMainHand)?.item.type === ItemType.ItemTypeWeapon;

// Every shaman imbues the main hand through its class options. Shamans do not dual wield, so the main
// hand is the only weapon a class imbue goes on.
export const MAIN_HAND_CLASS_IMBUES: ClassImbueHand = {
	values: SHAMAN_IMBUES.map(([imbue, spellId]) => ({
		value: pickOf(imbue),
		actionId: ActionId.fromSpellId(spellId),
		showWhen: (player: Player<any>) => isShaman(player) && holdsWeapon(player),
	})),
	getPick: (player: Player<any>) => (isShaman(player) ? pickOf(asShaman(player).getClassOptions().imbueMh) : 0),
	setPick: (player: Player<any>, value: number) => {
		if (isShaman(player)) asShaman(player).setClassOptions({ ...asShaman(player).getClassOptions(), imbueMh: imbueOf(value) });
	},
};
