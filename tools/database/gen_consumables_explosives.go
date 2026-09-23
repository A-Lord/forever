package database

import (
	"fmt"
	"go/format"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
)

const explosivesEffectsFile = "sim/common/forever/consumables_explosives_auto_gen.go"

// Writes a registration for every explosive the catalogue selects whose on-use spell routes as
// damage, and each one that does not as a commented-out call beneath its reason. The ids are
// spelled out rather than read from the database at start-up: the web build loads its database per
// request, after registration.
func GenerateExplosiveEffects(instance *dbc.DBC) error {
	var live, refused strings.Builder
	for _, id := range slices.Sorted(maps.Keys(instance.Consumables)) {
		consumable := instance.Consumables[id]
		if consumable.GetConsumableType() != proto.ConsumableType_ConsumableTypeExplosive {
			continue
		}

		routing := RouteConsumable(&consumable, instance)
		switch {
		case routing.Supported() && routing.Kind == ConsumableDamage:
			fmt.Fprintf(&live, "\tshared.NewSpellDataExplosive(%d) // %s\n", id, consumable.Name)
		case routing.Supported():
			fmt.Fprintf(&refused, "\t// unsupported: %d routes as %s, which the explosives builder does not build\n", routing.SpellID, routing.Kind)
			fmt.Fprintf(&refused, "\t// shared.NewSpellDataExplosive(%d) // %s\n", id, consumable.Name)
		default:
			fmt.Fprintf(&refused, "\t// unsupported: %s\n", routing.Reason())
			fmt.Fprintf(&refused, "\t// shared.NewSpellDataExplosive(%d) // %s\n", id, consumable.Name)
		}
	}

	var out strings.Builder
	out.WriteString("package forever\n\n")
	if live.Len() > 0 {
		out.WriteString("import \"github.com/wowsims/forever/sim/common/shared\"\n\n")
	}
	out.WriteString("func RegisterAllExplosives() {\n")
	out.WriteString(live.String())
	if refused.Len() > 0 {
		out.WriteString("\n")
		out.WriteString(refused.String())
	}
	out.WriteString("}\n")

	formatted, err := format.Source([]byte(out.String()))
	if err != nil {
		return fmt.Errorf("failed to format %s: %w", explosivesEffectsFile, err)
	}
	return os.WriteFile(explosivesEffectsFile, formatted, 0644)
}
