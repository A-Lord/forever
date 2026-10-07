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
	quality := item.OverallQuality
	if item.OverallQuality > 6 {
		quality = 4 // Heirlooms = epic
	}

	ilvl := 0
	if itemLevel > 0 {
		ilvl = itemLevel
	} else {
		ilvl = item.ItemLevel
	}

	switch item.InventoryType {
	case INVTYPE_WEAPON, INVTYPE_WEAPONMAINHAND, INVTYPE_WEAPONOFFHAND:
		{
			dps := dbcInstance.ItemDamageTable["ItemDamageOneHand"][ilvl].Quality[quality]
			if item.Flags1.Has(CASTER_WEAPON) {
				dps = dbcInstance.ItemDamageTable["ItemDamageOneHandCaster"][ilvl].Quality[quality]
			}
			if item.Flags4.Has(REDUCED_CASTER_DPS) {
				dps *= casterOneHandDpsShare
			}
			return dps
		}
	case INVTYPE_2HWEAPON:
		dps := dbcInstance.ItemDamageTable["ItemDamageTwoHand"][ilvl].Quality[quality]
		if item.Flags1.Has(CASTER_WEAPON) {
			dps = dbcInstance.ItemDamageTable["ItemDamageTwoHandCaster"][ilvl].Quality[quality]
		}
		if item.Flags4.Has(REDUCED_CASTER_DPS) {
			dps *= casterTwoHandDpsShare
		}
		return dps
	case INVTYPE_RANGED, INVTYPE_THROWN, INVTYPE_RANGEDRIGHT:
		switch item.ItemSubClass {
		case ITEM_SUBCLASS_WEAPON_BOW, ITEM_SUBCLASS_WEAPON_GUN, ITEM_SUBCLASS_WEAPON_CROSSBOW:
			return dbcInstance.ItemDamageTable["ItemDamageRanged"][ilvl].Quality[quality]
		case ITEM_SUBCLASS_WEAPON_THROWN:
			return dbcInstance.ItemDamageTable["ItemDamageThrown"][ilvl].Quality[quality]
		case ITEM_SUBCLASS_WEAPON_WAND:
			return dbcInstance.ItemDamageTable["ItemDamageWand"][ilvl].Quality[quality]
		}
	}
	return 0
}
