package protection

import (
	"math"
	"testing"
)

// The protection talents without Last Stand.
const withoutLastStandTalents = "05--552530233311210531"

// Last Stand 12976 states its 30% on the flat maximum health aura, which the warrior reads by hand as
// a percentage. The talent adds nothing while the aura is down, and the aura adds 30% of maximum
// health and not a flat 30 on top: a parse of the row would add that 30.
func TestLastStandAddsOnlyItsPercentage(t *testing.T) {
	sim, war, _ := shieldTestSim(t, DefaultProtectionTalents, 0)
	_, untalented, _ := shieldTestSim(t, withoutLastStandTalents, 0)

	before := war.MaxHealth()
	if got := untalented.MaxHealth(); got != before {
		t.Errorf("maximum health with the talent and its aura down: %v, want the untalented %v", before, got)
	}

	aura := war.GetAura("Last Stand")
	if aura == nil {
		t.Fatal("the warrior has no Last Stand aura")
	}
	aura.Activate(sim)

	if got, want := war.MaxHealth(), before*1.3; math.Abs(got-want) > 1e-6 {
		t.Errorf("maximum health under Last Stand: %v, want %v", got, want)
	}
}
