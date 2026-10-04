import { BulkEnchantSlotOptions } from '@generated/proto/api';
import { EnchantProcKind, ItemSlot } from '@generated/proto/common';
import type { UIEnchant as Enchant } from '@generated/proto/ui';
import { MISSING_ENCHANT_EFFECTS } from '@sim/constants/missing_effects_auto_gen';
import type { Player } from '@sim/player/player';
import { getWeaponDpsStatsBySlot } from '@sim/proto/equipped_item';
import type { Gear } from '@sim/proto/gear';
import { enchantAppliesToItem } from '@sim/proto/items';
import { Stats } from '@sim/proto/stats';

export type SkippedEnchantSlot = { slot: ItemSlot; reason: 'proc'; blockingEnchant: Enchant } | { slot: ItemSlot; reason: 'unlisted' };
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

// The enchant preset pool lists the ring enchants under Finger 1 for both fingers.
const presetPoolSlot = (slot: ItemSlot) => (slot === ItemSlot.ItemSlotFinger2 ? ItemSlot.ItemSlotFinger1 : slot);

// The static enchants the gear optimizer may pick per slot. A slot whose item can take a proc enchant
// the sim implements is left to the sim: a combat proc for every spec, a survival proc for tanks.
// With a preset pool, a slot draws only on the enchants the checked presets list for it: a listed proc the
// sim implements leaves it to the sim for every spec, survival included, every listed static enchant is a
// candidate whatever its EP, and a slot with no listed enchant that fits is left alone.
export function buildEnchantOptions(
	player: Player<any>,
	gear: Gear,
	frozenSlots: readonly ItemSlot[],
	weights: Stats,
	presetPool?: readonly BulkEnchantSlotOptions[],
): EnchantOptionsResult {
	const options: BulkEnchantSlotOptions[] = [];
	const skipped: SkippedEnchantSlot[] = [];
	const survivalBlocks = !!presetPool || player.playerSpec.isTankSpec;
	for (const slot of ENCHANTABLE_SLOTS) {
		const equipped = gear.getEquippedItem(slot);
		if (!equipped || frozenSlots.includes(slot)) continue;
		let fitting = player.getEnchants(slot).filter(enchant => enchantAppliesToItem(enchant, equipped.item));
		if (presetPool) {
			const listed = presetPool.find(pooled => pooled.slot === presetPoolSlot(slot))?.effectIds ?? [];
			fitting = fitting.filter(enchant => listed.includes(enchant.effectId));
			if (!fitting.length) {
				skipped.push({ slot, reason: 'unlisted' });
				continue;
			}
		}
		const blocking = fitting.find(
			enchant =>
				(enchant.procKind === EnchantProcKind.EnchantProcKindCombat ||
					(survivalBlocks && enchant.procKind === EnchantProcKind.EnchantProcKindSurvival)) &&
				!MISSING_ENCHANT_EFFECTS.has(enchant.effectId),
		);
		if (blocking) {
			skipped.push({ slot, reason: 'proc', blockingEnchant: blocking });
			continue;
		}
		const weaponSpeed = equipped.item.weaponSpeed;
		const epPerWeaponDps = weaponSpeed ? getWeaponDpsStatsBySlot(1, slot).computeEP(weights) : 0;
		const effectIds = fitting
			.filter(enchant => enchant.procKind === EnchantProcKind.EnchantProcKindNone && hasStaticValue(enchant))
			.filter(enchant => !!presetPool || enchantEP(enchant, weights, weaponSpeed, epPerWeaponDps) > 0)
			.map(enchant => enchant.effectId);
		if (effectIds.length) options.push(BulkEnchantSlotOptions.create({ slot, effectIds: [...new Set(effectIds)].sort((a, b) => a - b) }));
	}
	return { options, skipped };
}
