package database

import (
	"testing"

	"github.com/wowsims/forever/tools/database/dbc"
)

func TestWeaponDamageRange(t *testing.T) {
	withDBCInputs(t)
	items := dbc.GetDBC().Items

	for _, tc := range []struct {
		name     string
		id       int
		min, max float64
	}{
		{"Death Speaker Scepter, one-hand caster", 2816, 28, 52},
		{"Simone's Cultivating Hammer, one-hand caster", 22380, 35, 67},
		{"Benediction, two-hand caster with a QualityModifier", 18608, 89, 155},
		{"Bearded Boneaxe, not a caster weapon", 2878, 25, 47},
	} {
		item, ok := items[tc.id]
		if !ok {
			t.Errorf("%s: item %d is not in the DBC inputs", tc.name, tc.id)
			continue
		}
		if got, want := [2]float64{item.WeaponDmgMin(0), item.WeaponDmgMax(0)}, [2]float64{tc.min, tc.max}; got != want {
			t.Errorf("%s: damage %v-%v, want %v-%v", tc.name, got[0], got[1], want[0], want[1])
		}
	}
}
