//go:build with_db

package feralcat

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
)

func TestRotationWithoutShiftingPowerDoesNotReportManaExhaustion(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	cat.SpendMana(sim, cat.CurrentMana(), cat.NewManaMetrics(core.ActionID{SpellID: 1}))
	cat.doRotation(sim)
	if cat.IsOOM() {
		t.Fatal("rotation without Shifting Power reported mana exhaustion")
	}
}
