package core

import (
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
)

// The client's SpellEquippedItems for a weapon: the item class, the weapon subclass and inventory-type
// masks, and the hand an ATTR_EX_3 hand bit names. A zero field is no restriction.
type EquipRequirement struct {
	Class        int32
	SubclassMask int32
	InvTypeMask  int32
	Hand         ProcMask
}

func (req EquipRequirement) IsZero() bool {
	return req.SubclassMask == 0 && req.InvTypeMask == 0 && req.Hand == 0
}

// Whether the weapon behind an attack with this proc mask meets the requirement. The weapon is the
// main hand's for a main-hand bit, else the off hand's, else the ranged slot's. An attack with no
// weapon bit, a unit that is not a player, and a shield or held item in the off hand never match.
func (req EquipRequirement) Matches(unit *Unit, spellProcMask ProcMask) bool {
	if req.Class != 0 && req.Class != dbcenums.ITEM_CLASS_WEAPON {
		return false
	}
	if unit == nil || unit.Type != PlayerUnit || unit.Env == nil {
		return false
	}
	agent := unit.Env.Raid.GetPlayerFromUnit(unit)
	if agent == nil {
		return false
	}
	equipment := &agent.GetCharacter().Equipment

	if req.Hand != 0 {
		spellProcMask &= req.Hand
	}
	var slot proto.ItemSlot
	switch {
	case spellProcMask.Matches(ProcMaskMeleeMH):
		slot = proto.ItemSlot_ItemSlotMainHand
	case spellProcMask.Matches(ProcMaskMeleeOH):
		slot = proto.ItemSlot_ItemSlotOffHand
	case spellProcMask.Matches(ProcMaskRanged):
		slot = proto.ItemSlot_ItemSlotRanged
	default:
		return false
	}

	item := &equipment[slot]
	subclass := item.WeaponSubclassBit()
	if subclass == 0 {
		return false
	}
	if req.SubclassMask != 0 && req.SubclassMask&subclass == 0 {
		return false
	}
	if req.InvTypeMask != 0 && req.InvTypeMask&InvTypeBit(slot, item) == 0 {
		return false
	}
	return true
}

// The item's bit in a spell's weapon subclass mask, or 0 for an item that is not a weapon.
func (item *Item) WeaponSubclassBit() int32 {
	switch item.RangedWeaponType {
	case proto.RangedWeaponType_RangedWeaponTypeBow:
		return dbcenums.ITEM_SUBCLASS_BIT_WEAPON_BOW
	case proto.RangedWeaponType_RangedWeaponTypeGun:
		return dbcenums.ITEM_SUBCLASS_BIT_WEAPON_GUNS
	case proto.RangedWeaponType_RangedWeaponTypeCrossbow:
		return dbcenums.ITEM_SUBCLASS_BIT_WEAPON_CROSSBOW
	case proto.RangedWeaponType_RangedWeaponTypeThrown:
		return dbcenums.ITEM_SUBCLASS_BIT_WEAPON_THROWN
	case proto.RangedWeaponType_RangedWeaponTypeWand:
		return dbcenums.ITEM_SUBCLASS_BIT_WEAPON_WAND
	}

	twoHand := item.HandType == proto.HandType_HandTypeTwoHand
	switch item.WeaponType {
	case proto.WeaponType_WeaponTypeAxe:
		return Ternary[int32](twoHand, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_2H_AXE, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_1H_AXE)
	case proto.WeaponType_WeaponTypeMace:
		return Ternary[int32](twoHand, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_MACE_2H, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_MACE_1H)
	case proto.WeaponType_WeaponTypeSword:
		return Ternary[int32](twoHand, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_2H, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_1H)
	case proto.WeaponType_WeaponTypePolearm:
		return dbcenums.ITEM_SUBCLASS_BIT_WEAPON_POLEARM
	case proto.WeaponType_WeaponTypeStaff:
		return dbcenums.ITEM_SUBCLASS_BIT_WEAPON_STAFF
	case proto.WeaponType_WeaponTypeFist:
		return dbcenums.ITEM_SUBCLASS_BIT_WEAPON_FIST
	case proto.WeaponType_WeaponTypeDagger:
		return dbcenums.ITEM_SUBCLASS_BIT_WEAPON_DAGGERS
	}
	return 0
}

// The item's bits in a spell's inventory-type mask while it sits in this slot. A one-hand weapon also
// counts as the main-hand or off-hand weapon of the hand holding it.
func InvTypeBit(slot proto.ItemSlot, item *Item) int32 {
	switch item.RangedWeaponType {
	case proto.RangedWeaponType_RangedWeaponTypeBow:
		return dbcenums.INVTYPE_BIT_RANGED
	case proto.RangedWeaponType_RangedWeaponTypeThrown:
		return dbcenums.INVTYPE_BIT_THROWN
	case proto.RangedWeaponType_RangedWeaponTypeGun, proto.RangedWeaponType_RangedWeaponTypeCrossbow, proto.RangedWeaponType_RangedWeaponTypeWand:
		return dbcenums.INVTYPE_BIT_RANGEDRIGHT
	}

	switch item.WeaponType {
	case proto.WeaponType_WeaponTypeShield:
		return dbcenums.INVTYPE_BIT_SHIELD
	case proto.WeaponType_WeaponTypeOffHand:
		return dbcenums.INVTYPE_BIT_HOLDABLE
	}

	switch item.HandType {
	case proto.HandType_HandTypeTwoHand:
		return dbcenums.INVTYPE_BIT_2HWEAPON
	case proto.HandType_HandTypeMainHand:
		return dbcenums.INVTYPE_BIT_WEAPONMAINHAND
	case proto.HandType_HandTypeOffHand:
		return dbcenums.INVTYPE_BIT_WEAPONOFFHAND
	case proto.HandType_HandTypeOneHand:
		switch slot {
		case proto.ItemSlot_ItemSlotMainHand:
			return dbcenums.INVTYPE_BIT_WEAPON | dbcenums.INVTYPE_BIT_WEAPONMAINHAND
		case proto.ItemSlot_ItemSlotOffHand:
			return dbcenums.INVTYPE_BIT_WEAPON | dbcenums.INVTYPE_BIT_WEAPONOFFHAND
		}
		return dbcenums.INVTYPE_BIT_WEAPON
	}
	return 0
}

// The weapon subclass mask naming every weapon of these types, one- and two-handed alike.
func WeaponSubclassMask(types ...proto.WeaponType) int32 {
	var mask int32
	for _, weaponType := range types {
		for _, handType := range []proto.HandType{proto.HandType_HandTypeOneHand, proto.HandType_HandTypeTwoHand} {
			mask |= (&Item{WeaponType: weaponType, HandType: handType}).WeaponSubclassBit()
		}
	}
	return mask
}
