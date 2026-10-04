//go:build with_db

package feralcat

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/stats"
)

func TestGenesisIncreasesCatBleedDamage(t *testing.T) {
	_, base := newCat(t, "", nil)
	_, talented := newCat(t, "0500000000000000-00000000000000000000-0000000000000000", nil)
	for _, bleed := range []struct {
		name       string
		base, with float64
	}{
		{"Rake", base.Rake.Dot(base.CurrentTarget).PeriodicDamageMultiplier, talented.Rake.Dot(talented.CurrentTarget).PeriodicDamageMultiplier},
		{"Rip", base.Rip.Dot(base.CurrentTarget).PeriodicDamageMultiplier, talented.Rip.Dot(talented.CurrentTarget).PeriodicDamageMultiplier},
	} {
		if got, want := bleed.with/bleed.base, 1.05; math.Abs(got-want) > 1e-9 {
			t.Errorf("%s periodic damage multiplier ratio = %v, want %v", bleed.name, got, want)
		}
	}
}

func TestNaturesMajestyRaisesCatMeleeAndSpellCrit(t *testing.T) {
	_, base := newCat(t, "", nil)
	_, talented := newCat(t, "0000200000000000-00000000000000000000-0000000000000000", nil)
	want := 4.0
	for _, stat := range []stats.Stat{stats.PhysicalCritPercent, stats.SpellCritPercent} {
		if got := talented.GetStat(stat) - base.GetStat(stat); math.Abs(got-want) > 1e-9 {
			t.Errorf("Nature's Majesty %s = %v, want %v", stat.StatName(), got, want)
		}
	}
}

func TestNaturesReachRaisesCatMeleeAndSpellHit(t *testing.T) {
	_, base := newCat(t, "", nil)
	_, talented := newCat(t, "0000020000000000-00000000000000000000-0000000000000000", nil)
	want := 4.0
	for _, stat := range []stats.Stat{stats.PhysicalHitPercent, stats.SpellHitPercent} {
		if got := talented.GetStat(stat) - base.GetStat(stat); math.Abs(got-want) > 1e-9 {
			t.Errorf("Nature's Reach %s = %v, want %v", stat.StatName(), got, want)
		}
	}
}
