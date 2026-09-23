//go:build with_db

package common_test

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core"
)

// The consumables applyConsumeEffects handles itself rather than through a registration: Gift of
// Arthas 9088.
var consumablesAppliedInline = []int32{9088}

// Every consumable db.json ships is one the sim applies: a hand-written or data-driven registration
// for its id, or one applyConsumeEffects handles itself.
func TestEveryShippedConsumableIsRegistered(t *testing.T) {
	registerAllEffects()

	var missing []int32
	for _, id := range core.ConsumableIDs() {
		if !core.HasConsumableEffect(id) && !slices.Contains(consumablesAppliedInline, id) {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		t.Errorf("%d consumables in db.json register no effect: %v", len(missing), missing)
	}
}
