package dbc

import "math"

// WeaponDamage computes the equipped endpoints from a selected damage-table cell.
func (item *Item) WeaponDamage(dps float64) (float64, float64) {
	speed := float64(item.ItemDelay) / 1000
	low := dps*speed*(1-item.DmgVariance/2) + item.QualityModifier*speed
	high := dps*speed*(1+item.DmgVariance/2) + item.QualityModifier*speed
	if low < 0 {
		low = 1
	}
	if high < 0 {
		high = 1
	}
	return math.Floor(low), math.Floor(high + 0.5)
}
