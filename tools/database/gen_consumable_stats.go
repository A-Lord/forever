package database

import (
	"fmt"
	"go/format"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/wowsims/forever/tools/database/dbc"
)

const statConsumablesFile = "sim/common/forever/consumable_stats_auto_gen.go"

// Registers every consumable db.json carries whose on-use spell routes as a stat aura, except those
// whose on-use shares the potion or conjured category, which GeneratePotionAndConjuredEffects
// registers.
func GenerateStatConsumables(instance *dbc.DBC, db *WowDatabase) error {
	var out strings.Builder
	out.WriteString("package forever\n\n")
	out.WriteString("import \"github.com/wowsims/forever/sim/common/shared\"\n\n")
	out.WriteString("func RegisterAllStatConsumables() {\n")
	for _, id := range slices.Sorted(maps.Keys(db.Consumables)) {
		consumable, ok := instance.Consumables[int(id)]
		if category := OnUseCategory(db.Consumables[id]); !ok || category == PotionCategory || category == ConjuredCategory {
			continue
		}
		if routing := RouteConsumable(&consumable, instance); routing.Supported() && routing.Kind == ConsumableStatAura {
			fmt.Fprintf(&out, "\tshared.NewSpellDataStatConsumable(%d) // %s\n", id, consumable.Name)
		}
	}
	out.WriteString("}\n")

	formatted, err := format.Source([]byte(out.String()))
	if err != nil {
		return fmt.Errorf("generated %s is not valid Go: %w", statConsumablesFile, err)
	}
	return os.WriteFile(statConsumablesFile, formatted, 0644)
}
