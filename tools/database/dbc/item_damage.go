package dbc

type ItemDamageTable struct {
	ItemLevel int
	Quality   []float64
}

// A weapon flagged REDUCED_CASTER_DPS deals a fixed share of its damage table in game, while the
// game files' caster tables match the regular ones. The shares reproduce the in-game damage range of
// all 165 flagged weapons, QualityModifier included; 29/39 of the two-hand table is 29/30 of the one-hand one.
const (
	casterOneHandDpsShare = 2.0 / 3.0
	casterTwoHandDpsShare = 29.0 / 39.0
)

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
	return row.Quality[quality] * item.casterDpsShare(), true
}

func (item *Item) casterDpsShare() float64 {
	if !item.Flags4.Has(REDUCED_CASTER_DPS) {
		return 1
	}
	switch item.InventoryType {
	case INVTYPE_WEAPON, INVTYPE_WEAPONMAINHAND, INVTYPE_WEAPONOFFHAND:
		return casterOneHandDpsShare
	case INVTYPE_2HWEAPON:
		return casterTwoHandDpsShare
	}
	return 1
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
