package database

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
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

func TestCasterWeaponSpellPower(t *testing.T) {
	withDBCInputs(t)
	items := dbc.GetDBC().Items

	for _, tc := range []struct {
		name            string
		id              int
		healing, damage float64
	}{
		{"Simone's Cultivating Hammer, healing weapon", 22380, 139, 46},
		{"Blessed Qiraji Augur Staff, healing weapon", 21275, 203, 67},
		{"Staff of Conjuring, spell power and no stat of its own", 1933, 8, 8},
		{"Crackling Staff, spell power on top of its own", 19102, 99, 99},
		{"Death Speaker Scepter, reduced damage only", 2816, 68, 0},
	} {
		item, ok := items[tc.id]
		if !ok {
			t.Errorf("%s: item %d is not in the DBC inputs", tc.name, tc.id)
			continue
		}
		stats := item.GetStats(item.ItemLevel)
		if got, want := [2]float64{stats[proto.Stat_StatHealingPower], stats[proto.Stat_StatSpellDamage]}, [2]float64{tc.healing, tc.damage}; got != want {
			t.Errorf("%s: healing %v, spell damage %v, want %v and %v", tc.name, got[0], got[1], want[0], want[1])
		}
	}
}
