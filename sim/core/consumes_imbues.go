package core

import (
	"github.com/wowsims/forever/sim/core/proto"
)

// Puts the temporary enchant each imbue the spec names on its hand, where the enchant takes that
// hand's weapon; the main hand's only where neither of the party's weapon totems, Windfury Totem and
// Flametongue Totem, takes its place. It runs before the weapons are built, and a weapon built from a
// hand's item takes the enchant's flat weapon damage. The consumable's registered effect then applies
// the rest of the enchant to the hands carrying it.
func applyTemporaryEnchants(character *Character, consumables *proto.ConsumesSpec, partyBuffs *proto.PartyBuffs) {
	weaponTotem := partyBuffs.GetWindfuryTotem() || partyBuffs.GetFlametongueTotem()
	if !weaponTotem || !character.TakesTotemEnchant() {
		character.applyTemporaryEnchant(consumables.MhImbueId, proto.ItemSlot_ItemSlotMainHand)
	}
	character.applyTemporaryEnchant(consumables.OhImbueId, proto.ItemSlot_ItemSlotOffHand)
}

// Windfury Totem and Flametongue Totem enchant the main hand, which does nothing for a druid in Cat or
// Bear Form, so the feral specs neither proc them nor give their main-hand imbue up for them.
func (character *Character) TakesTotemEnchant() bool {
	return !character.isFeralDruid()
}

func (character *Character) isFeralDruid() bool {
	return character.Spec == proto.Spec_SpecFeralCatDruid || character.Spec == proto.Spec_SpecFeralBearDruid
}

func (character *Character) applyTemporaryEnchant(itemID int32, slot proto.ItemSlot) {
	consumable := GetConsumableByID(itemID)
	if consumable.TemporaryEnchant == nil || !consumable.TakesItem(slot, character.Equipment[slot]) {
		return
	}
	character.Equipment[slot].TempEnchant = consumable.TemporaryEnchant.EffectID
	character.Equipment[slot].TempEnchantWeaponDamage = consumable.TemporaryEnchant.WeaponDamage
}

// Whether the temporary enchant's SpellEquippedItems masks take the item in this slot.
func (consumable Consumable) TakesItem(slot proto.ItemSlot, item Item) bool {
	if item.ID == 0 {
		return false
	}
	return (consumable.WeaponSubclassMask == 0 || consumable.WeaponSubclassMask&item.WeaponSubclassBit() != 0) &&
		(consumable.WeaponInventoryTypeMask == 0 || consumable.WeaponInventoryTypeMask&invTypeBit(slot, &item) != 0)
}
