package core

import "github.com/wowsims/forever/sim/core/proto"

func (character *Character) HasEnchantEquipped(effectID int32, possibleSlots []proto.ItemSlot) bool {
	return character.hasEnchantEquipped(effectID, possibleSlots)
}
