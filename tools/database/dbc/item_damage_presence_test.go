package dbc

import "testing"

func TestWeaponDPSPresenceWhenSourceTableIsMissing(t *testing.T) {
	// Given
	once.Do(func() {})
	previous := dbcInstance
	dbcInstance = NewDBC()
	t.Cleanup(func() { dbcInstance = previous })
	item := Item{ItemClass: ITEM_CLASS_WEAPON, InventoryType: INVTYPE_WEAPON, OverallQuality: 2, ItemLevel: 18}
	// When
	row := item.ToUIItem().ScalingOptions[0]
	// Then
	if row.UnroundedDps != nil {
		t.Fatal("missing source row must preserve absent table DPS")
	}
}
