package database

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestConsumableTSName(t *testing.T) {
	for name, want := range map[string]string{
		"Mighty Rage Potion":                   "MightyRagePotion",
		"R.O.I.D.S.":                           "ROIDS",
		"Kreeg's Stout Beatdown":               "KreegsStoutBeatdown",
		"Scroll of Arcane Protection - Fire I": "ScrollOfArcaneProtectionFireI",
		"Filet o' Flank":                       "FiletOFlank",
		"EZ-Thro Copper Bomb XL":               "EZThroCopperBombXL",
		"M73 Frag Grenade":                     "M73FragGrenade",
		"7 Days":                               "Item7Days",
	} {
		if got := consumableTSName(name); got != want {
			t.Errorf("consumableTSName(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestConsumableTypeListName(t *testing.T) {
	for consumableType, want := range map[proto.ConsumableType]string{
		proto.ConsumableType_ConsumableTypePotion:       "POTION_CONSUMABLES",
		proto.ConsumableType_ConsumableTypeBattleElixir: "BATTLE_ELIXIR_CONSUMABLES",
		proto.ConsumableType_ConsumableTypePetFood:      "PET_FOOD_CONSUMABLES",
	} {
		if got := consumableTypeListName(consumableType); got != want {
			t.Errorf("consumableTypeListName(%v) = %q, want %q", consumableType, got, want)
		}
	}
}
