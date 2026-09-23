package database

import (
	"fmt"
	"go/format"
	"os"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/dbc"
)

// The categories potions and conjured items share their cooldown through: 4 for potions, 1153 for
// healthstones, mana gems, runes and Thistle Tea.
const (
	PotionCategory   int32 = 4
	ConjuredCategory int32 = 1153
)

// The category the consumable's on-use shares: the item's row, or where it states none, its spell's.
// The Spellblasting, Frenzy and Mender's potions state theirs on the spell alone.
func OnUseCategory(c *proto.Consumable) int32 {
	for _, effect := range c.ItemEffects {
		onUse := effect.GetOnUse()
		if onUse == nil {
			continue
		}
		if onUse.CategoryId > 0 {
			return onUse.CategoryId
		}
		return int32(spelldata.Find(effect.BuffId).Category)
	}
	return 0
}

// Sets each consumable's category to the one its on-use shares, which the settings file it by, and
// types one sharing the potion category a potion: Potion of Demonslaying 9224 is an Elixir-subclass
// item whose row states no category, and its spell 11406 shares category 4.
func AddOnUseCategories(db *WowDatabase) {
	for _, c := range db.Consumables {
		category := OnUseCategory(c)
		if category <= 0 {
			continue
		}
		c.CategoryId = category
		if category == PotionCategory {
			c.Type = proto.ConsumableType_ConsumableTypePotion
		}
	}
}

// Writes the registration of every potion and conjured item db.json ships, by the category its on-use
// shares, and of every other consumable whose on-use routes as a heal, an absorb or a gain of a
// resource, to sim/common/forever/consumables_potions_auto_gen.go. One the builder cannot run is left commented
// out with the reason.
func GeneratePotionAndConjuredEffects(instance *dbc.DBC, db *WowDatabase) {
	type entry struct {
		id          int32
		name        string
		constructor string
		unsupported []string
	}
	var potions, conjured, others []entry
	for _, c := range db.Consumables {
		var constructor string
		switch OnUseCategory(c) {
		case PotionCategory:
			constructor = "NewSpellDataPotion"
		case ConjuredCategory:
			constructor = "NewSpellDataConjured"
		default:
			consumable, ok := instance.Consumables[int(c.Id)]
			if !ok {
				continue
			}
			switch RouteConsumable(&consumable, instance).Kind {
			case ConsumableHeal, ConsumableAbsorb, ConsumableEnergize:
				constructor = "NewSpellDataOnUseConsumable"
			default:
				continue
			}
		}

		e := entry{id: c.Id, name: c.Name, constructor: constructor}
		for _, effect := range c.ItemEffects {
			if effect.GetOnUse() != nil {
				e.unsupported = shared.ConsumableUnsupported(spelldata.Find(effect.BuffId))
				break
			}
		}
		switch constructor {
		case "NewSpellDataPotion":
			potions = append(potions, e)
		case "NewSpellDataConjured":
			conjured = append(conjured, e)
		default:
			others = append(others, e)
		}
	}

	var out strings.Builder
	out.WriteString("package forever\n\n")
	out.WriteString("import \"github.com/wowsims/forever/sim/common/shared\"\n\n")
	out.WriteString("func RegisterPotionsAndConjured() {\n")
	for _, group := range []struct {
		name    string
		entries []entry
	}{{"Potions", potions}, {"Conjured", conjured}, {"Other on-use consumables", others}} {
		slices.SortFunc(group.entries, func(a, b entry) int { return int(a.id - b.id) })
		fmt.Fprintf(&out, "\n\t// %s\n", group.name)
		for _, e := range group.entries {
			if len(e.unsupported) > 0 {
				fmt.Fprintf(&out, "\t// unsupported: %s\n\t// shared.%s(%d) // %s\n", strings.Join(e.unsupported, "; "), e.constructor, e.id, e.name)
				continue
			}
			fmt.Fprintf(&out, "\tshared.%s(%d) // %s\n", e.constructor, e.id, e.name)
		}
	}
	out.WriteString("}\n")

	outFile := "sim/common/forever/consumables_potions_auto_gen.go"
	formatted, err := format.Source([]byte(out.String()))
	if err != nil {
		panic(fmt.Sprintf("generated %s is not valid Go: %v", outFile, err))
	}
	if err := os.WriteFile(outFile, formatted, 0644); err != nil {
		panic(fmt.Sprintf("failed to write file %s: %v", outFile, err))
	}
}
