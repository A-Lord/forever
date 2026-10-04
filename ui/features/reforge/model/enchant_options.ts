import { BulkEnchantSlotOptions } from '@generated/proto/api';
import { EnchantProcKind, ItemSlot } from '@generated/proto/common';
import type { UIEnchant as Enchant } from '@generated/proto/ui';
import { MISSING_ENCHANT_EFFECTS } from '@sim/constants/missing_effects_auto_gen';
import type { Player } from '@sim/player/player';
import { getWeaponDpsStatsBySlot } from '@sim/proto/equipped_item';
import type { Gear } from '@sim/proto/gear';
import { enchantAppliesToItem } from '@sim/proto/items';
import { Stats } from '@sim/proto/stats';

export type SkippedEnchantSlot = { slot: ItemSlot; blockingEnchant: Enchant };
export type EnchantOptionsResult = { options: BulkEnchantSlotOptions[]; skipped: SkippedEnchantSlot[] };

export const ENCHANTABLE_SLOTS: readonly ItemSlot[] = [
	ItemSlot.ItemSlotHead,
	ItemSlot.ItemSlotShoulder,
	ItemSlot.ItemSlotBack,
	ItemSlot.ItemSlotChest,
	ItemSlot.ItemSlotWrist,
	ItemSlot.ItemSlotHands,
	ItemSlot.ItemSlotLegs,
	ItemSlot.ItemSlotFeet,
	ItemSlot.ItemSlotFinger1,
	ItemSlot.ItemSlotFinger2,
	ItemSlot.ItemSlotMainHand,
	ItemSlot.ItemSlotOffHand,
	ItemSlot.ItemSlotRanged,
];

const hasStaticValue = (enchant: Enchant) =>
	enchant.stats.some(value => value !== 0) || enchant.pseudoStats.some(value => value !== 0) || enchant.weaponDamage !== 0;

// Player.computeEnchantEP's formula, on the weights handed in rather than the player's own.
const enchantEP = (enchant: Enchant, weights: Stats, weaponSpeed: number, epPerWeaponDps: number) => {
	const ep = new Stats(enchant.stats, enchant.pseudoStats).computeEP(weights);
	return enchant.weaponDamage && weaponSpeed ? ep + (enchant.weaponDamage / weaponSpeed) * epPerWeaponDps : ep;
};

// The static enchants the gear optimizer may pick per slot. A slot whose item can take a proc enchant
// the sim implements is left to the sim: a combat proc for every spec, a survival proc for tanks.
export function buildEnchantOptions(player: Player<any>, gear: Gear, frozenSlots: readonly ItemSlot[], weights: Stats): EnchantOptionsResult {
	const options: BulkEnchantSlotOptions[] = [];
	const skipped: SkippedEnchantSlot[] = [];
	const isTank = player.playerSpec.isTankSpec;
	for (const slot of ENCHANTABLE_SLOTS) {
		const equipped = gear.getEquippedItem(slot);
		if (!equipped || frozenSlots.includes(slot)) continue;
		const fitting = player.getEnchants(slot).filter(enchant => enchantAppliesToItem(enchant, equipped.item));
		const blocking = fitting.find(
			enchant =>
				(enchant.procKind === EnchantProcKind.EnchantProcKindCombat || (isTank && enchant.procKind === EnchantProcKind.EnchantProcKindSurvival)) &&
				!MISSING_ENCHANT_EFFECTS.has(enchant.effectId),
		);
		if (blocking) {
			skipped.push({ slot, blockingEnchant: blocking });
			continue;
		}
		const weaponSpeed = equipped.item.weaponSpeed;
		const epPerWeaponDps = weaponSpeed ? getWeaponDpsStatsBySlot(1, slot).computeEP(weights) : 0;
		const effectIds = fitting
			.filter(enchant => enchant.procKind === EnchantProcKind.EnchantProcKindNone && hasStaticValue(enchant))
			.filter(enchant => enchantEP(enchant, weights, weaponSpeed, epPerWeaponDps) > 0)
			.map(enchant => enchant.effectId);
		if (effectIds.length) options.push(BulkEnchantSlotOptions.create({ slot, effectIds: [...new Set(effectIds)].sort((a, b) => a - b) }));
	}
	return { options, skipped };
}
