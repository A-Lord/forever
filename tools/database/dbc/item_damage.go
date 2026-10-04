package dbc

type ItemDamageTable struct {
	ItemLevel int
	Quality   []float64
}

func (item *Item) WeaponDps(itemLevel int) float64 {
	dps, _ := dbcInstance.WeaponDPS(item, itemLevel)
	return dps
}

// WeaponDPS preserves whether the game files contain the selected quality cell.
func (d *DBC) WeaponDPS(item *Item, itemLevel int) (float64, bool) {
	quality := item.OverallQuality
	if quality > 6 {
		quality = 4
	}
	if itemLevel <= 0 {
		itemLevel = item.ItemLevel
	}
	table := item.weaponDamageTable()
	row, present := d.ItemDamageTable[table][itemLevel]
	if !present || quality < 0 || int(quality) >= len(row.Quality) {
		return 0, false
	}
	return row.Quality[quality], true
}

func (item *Item) weaponDamageTable() string {
	switch item.InventoryType {
	case INVTYPE_WEAPON, INVTYPE_WEAPONMAINHAND, INVTYPE_WEAPONOFFHAND:
		if item.Flags1.Has(CASTER_WEAPON) {
			return "ItemDamageOneHandCaster"
		}
		return "ItemDamageOneHand"
	case INVTYPE_2HWEAPON:
		if item.Flags1.Has(CASTER_WEAPON) {
			return "ItemDamageTwoHandCaster"
		}
		return "ItemDamageTwoHand"
	case INVTYPE_RANGED, INVTYPE_THROWN, INVTYPE_RANGEDRIGHT:
		switch item.ItemSubClass {
		case ITEM_SUBCLASS_WEAPON_BOW, ITEM_SUBCLASS_WEAPON_GUN, ITEM_SUBCLASS_WEAPON_CROSSBOW:
			return "ItemDamageRanged"
		case ITEM_SUBCLASS_WEAPON_THROWN:
			return "ItemDamageThrown"
		case ITEM_SUBCLASS_WEAPON_WAND:
			return "ItemDamageWand"
		}
	}
	return ""
}
