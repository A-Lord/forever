import { Class, ItemSlot, ItemType, Spec, WeaponType } from '@generated/proto/common';
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

// A shield or held-in-off-hand item takes no imbue.
const holdsWeapon = (player: Player<any>, slot: ItemSlot): boolean => {
	const item = player.getEquippedItem(slot)?.item;
	return (
		!!item && item.type === ItemType.ItemTypeWeapon && item.weaponType !== WeaponType.WeaponTypeShield && item.weaponType !== WeaponType.WeaponTypeOffHand
	);
};

const shamanHand = (
	slot: ItemSlot,
	offered: (player: Player<any>) => boolean,
	getImbue: (player: Player<any>) => ShamanImbue | undefined,
	setImbue: (player: Player<any>, imbue: ShamanImbue) => void,
): ClassImbueHand => ({
	values: SHAMAN_IMBUES.map(([imbue, spellId]) => ({
		value: pickOf(imbue),
		actionId: ActionId.fromSpellId(spellId),
		showWhen: (player: Player<any>) => offered(player) && holdsWeapon(player, slot),
	})),
	getPick: (player: Player<any>) => (offered(player) ? pickOf(getImbue(player)) : 0),
	setPick: (player: Player<any>, value: number) => {
		if (offered(player)) setImbue(player, imbueOf(value));
	},
});

// Every shaman imbues the main hand through its class options; only Enhancement has an off-hand
// imbue, in its spec options. The checks come first, so the casts name what they established.
const isShaman = (player: Player<any>) => player.getClass() === Class.ClassShaman;
const isEnhancement = (player: Player<any>) => player.getSpec() === Spec.SpecEnhancementShaman;
const asShaman = (player: Player<any>) => player as Player<ShamanSpecs>;
const asEnhancement = (player: Player<any>) => player as Player<Spec.SpecEnhancementShaman>;

export const MAIN_HAND_CLASS_IMBUES: ClassImbueHand = shamanHand(
	ItemSlot.ItemSlotMainHand,
	isShaman,
	player => asShaman(player).getClassOptions().imbueMh,
	(player, imbue) => asShaman(player).setClassOptions({ ...asShaman(player).getClassOptions(), imbueMh: imbue }),
);

export const OFF_HAND_CLASS_IMBUES: ClassImbueHand = shamanHand(
	ItemSlot.ItemSlotOffHand,
	isEnhancement,
	player => asEnhancement(player).getSpecOptions().imbueOh,
	(player, imbue) => asEnhancement(player).setSpecOptions({ ...asEnhancement(player).getSpecOptions(), imbueOh: imbue }),
);
