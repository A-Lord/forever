//go:build with_db

package common_test

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
)

// Every consumable db.json ships is one the sim applies: a hand-written or data-driven registration
// for its id.
func TestEveryShippedConsumableIsRegistered(t *testing.T) {
	registerAllEffects()

	var missing []int32
	for _, id := range core.ConsumableIDs() {
		if !core.HasConsumableEffect(id) {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		t.Errorf("%d consumables in db.json register no effect: %v", len(missing), missing)
	}
}
