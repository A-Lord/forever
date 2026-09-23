package core

import (
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

const consecratedSharpeningStoneID = 23122

// Puts the temporary enchant each imbue the spec names on its hand, where the enchant takes that
// hand's weapon; the main hand's only without Windfury Totem. The consumable's registered effect
// then applies the enchant to the hands carrying it.
func applyTemporaryEnchants(character *Character, consumables *proto.ConsumesSpec, partyBuffs *proto.PartyBuffs) {
	if partyBuffs.WindfuryTotem == proto.TristateEffect_TristateEffectMissing {
		character.applyTemporaryEnchant(consumables.MhImbueId, proto.ItemSlot_ItemSlotMainHand)
	}
	character.applyTemporaryEnchant(consumables.OhImbueId, proto.ItemSlot_ItemSlotOffHand)
}

func (character *Character) applyTemporaryEnchant(itemID int32, slot proto.ItemSlot) {
	if itemID == consecratedSharpeningStoneID {
		applyConsecratedSharpeningStone(character)
		return
	}

	consumable := GetConsumableByID(itemID)
	if consumable.TemporaryEnchant == nil || !consumable.TakesItem(character.Equipment[slot]) {
		return
	}
	character.Equipment[slot].TempEnchant = consumable.TemporaryEnchant.EffectID
	character.ItemSwap.AddTempEnchant(consumable.TemporaryEnchant.EffectID, slot, false)
}

// Its equip spell 28893 grants 99 melee and ranged attack power versus Undead, through auras the
// enchant reading does not parse.
func applyConsecratedSharpeningStone(character *Character) {
	character.Env.RegisterPostFinalizeEffect(func() {
		for _, at := range character.AttackTables {
			at.MobTypeBonusStats[proto.MobType_MobTypeUndead] = at.MobTypeBonusStats[proto.MobType_MobTypeUndead].Add(stats.Stats{
				stats.AttackPower:       99,
				stats.RangedAttackPower: 99,
			})
		}
	})
}

// Whether the temporary enchant's SpellEquippedItems masks take the item.
func (consumable Consumable) TakesItem(item Item) bool {
	if item.ID == 0 {
		return false
	}
	return (consumable.WeaponSubclassMask == 0 || consumable.WeaponSubclassMask&itemSubclassBit(item) != 0) &&
		(consumable.WeaponInventoryTypeMask == 0 || consumable.WeaponInventoryTypeMask&inventoryTypeBit(item) != 0)
}

// The item's bit in a weapon subclass mask: 1 << ItemSubclassWeapon.
func itemSubclassBit(item Item) int32 {
	twoHand := item.HandType == proto.HandType_HandTypeTwoHand
	switch item.WeaponType {
	case proto.WeaponType_WeaponTypeAxe:
		return Ternary[int32](twoHand, 1<<1, 1<<0)
	case proto.WeaponType_WeaponTypeMace:
		return Ternary[int32](twoHand, 1<<5, 1<<4)
	case proto.WeaponType_WeaponTypePolearm:
		return 1 << 6
	case proto.WeaponType_WeaponTypeSword:
		return Ternary[int32](twoHand, 1<<8, 1<<7)
	case proto.WeaponType_WeaponTypeStaff:
		return 1 << 10
	case proto.WeaponType_WeaponTypeFist:
		return 1 << 13
	case proto.WeaponType_WeaponTypeDagger:
		return 1 << 15
	}
	return 0
}

// The item's bit in an inventory type mask: 1 << InventoryType.
func inventoryTypeBit(item Item) int32 {
	switch item.WeaponType {
	case proto.WeaponType_WeaponTypeShield:
		return 1 << 14
	case proto.WeaponType_WeaponTypeOffHand:
		return 1 << 23
	}
	switch item.HandType {
	case proto.HandType_HandTypeOneHand:
		return 1 << 13
	case proto.HandType_HandTypeTwoHand:
		return 1 << 17
	case proto.HandType_HandTypeMainHand:
		return 1 << 21
	case proto.HandType_HandTypeOffHand:
		return 1 << 22
	}
	return 0
}
