package database

import (
	"fmt"
	"go/format"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/dbc"
)

type consumableRegistration struct {
	id          int32
	name        string
	constructor string
	// The ids the call passes after the consumable's own.
	args []int32
	// Why the builder refuses the consumable; its call is then written commented out beneath it.
	unsupported []string
}

type registrationGroup struct {
	title   string
	entries []consumableRegistration
}

// Writes the registration of every consumable a builder in sim/common/shared runs from its rows, one
// file per family, and each one a builder refuses as a commented-out call beneath the reason. The ids
// are spelled out rather than read from the loaded database, which the browser's sim does not have
// when the effects register. routes holds the routings gen_db already made, by item id.
func GenerateConsumableRegistrations(instance *dbc.DBC, db *WowDatabase, routes map[int]*ConsumableRouting) error {
	var explosives, refusedExplosives, potions, conjured, others, stats, imbues []consumableRegistration
	ids := map[int32]bool{}
	for id := range instance.Consumables {
		ids[int32(id)] = true
	}
	for id := range db.Consumables {
		ids[id] = true
	}

	for _, id := range slices.Sorted(maps.Keys(ids)) {
		c, shipped := db.Consumables[id]
		catalogued, inCatalogue := instance.Consumables[int(id)]
		routing := routes[int(id)]
		if routing == nil && inCatalogue {
			routing = RouteConsumable(&catalogued, instance)
		}
		name := catalogued.Name
		if shipped {
			name = c.Name
		}
		entry := consumableRegistration{id: id, name: name}

		switch {
		case shipped && c.Type == proto.ConsumableType_ConsumableTypeImbue:
			entry.constructor = "NewSpellDataTemporaryEnchant"
			if inCatalogue {
				entry.args = parsedEquipSpells(catalogued.TemporaryEnchant)
			}
			imbues = append(imbues, entry)
		case inCatalogue && IsExplosive(&catalogued, routing):
			entry.constructor = "NewSpellDataExplosive"
			if routing.Supported() {
				explosives = append(explosives, entry)
				continue
			}
			entry.unsupported = routing.Unsupported
			refusedExplosives = append(refusedExplosives, entry)
		case !shipped:
		case c.CategoryId == dbc.POTION_SPELL_CATEGORY:
			entry.constructor, entry.unsupported = "NewSpellDataOnUseConsumable", onUseUnsupported(c)
			potions = append(potions, entry)
		case c.Type == proto.ConsumableType_ConsumableTypeConjured:
			entry.constructor, entry.unsupported = "NewSpellDataOnUseConsumable", onUseUnsupported(c)
			conjured = append(conjured, entry)
		case !inCatalogue:
		case routing.Kind == ConsumableHeal, routing.Kind == ConsumableAbsorb, routing.Kind == ConsumableEnergize:
			entry.constructor, entry.unsupported = "NewSpellDataOnUseConsumable", onUseUnsupported(c)
			others = append(others, entry)
		case routing.Supported() && routing.Kind == ConsumableStatAura:
			entry.constructor = "NewSpellDataStatConsumable"
			stats = append(stats, entry)
		}
	}

	for _, file := range []struct {
		path, function string
		groups         []registrationGroup
	}{
		{"sim/common/forever/consumables_explosives_auto_gen.go", "RegisterAllExplosives",
			[]registrationGroup{{entries: explosives}, {entries: refusedExplosives}}},
		{"sim/common/forever/consumables_potions_auto_gen.go", "RegisterPotionsAndConjured",
			[]registrationGroup{{"Potions", potions}, {"Conjured", conjured}, {"Other on-use consumables", others}}},
		{"sim/common/forever/consumables_imbues_auto_gen.go", "RegisterAllTemporaryEnchantConsumables",
			[]registrationGroup{{entries: imbues}}},
		{"sim/common/forever/consumable_stats_auto_gen.go", "RegisterAllStatConsumables",
			[]registrationGroup{{entries: stats}}},
	} {
		if err := writeRegistrationFile(file.path, file.function, file.groups); err != nil {
			return err
		}
	}
	return nil
}

// The equip spells of a temporary enchant that the imbue builder parses: those that add no stat,
// resistance or weapon damage to the enchant's own fields, and whose every aura the parser attaches.
// Consecrated Sharpening Stone's enchant 2684 names 28893, its attack power against Undead; a wizard
// oil's equip spell is already the enchant's spell damage.
func parsedEquipSpells(enchant dbc.TemporaryEnchant) []int32 {
	var spellIDs []int32
	for slot, effect := range enchant.Effects {
		spellID := int32(enchant.Args[slot])
		if effect != dbc.ITEM_ENCHANTMENT_EQUIP_SPELL || spellID == 0 {
			continue
		}
		alone := dbc.TemporaryEnchant{ID: enchant.ID}
		alone.Effects[0], alone.Points[0], alone.Args[0] = effect, enchant.Points[slot], enchant.Args[slot]
		fields := alone.ToProto()
		row := spelldata.Find(spellID)
		if slices.ContainsFunc(fields.Stats, isNonZero) || slices.ContainsFunc(fields.PseudoStats, isNonZero) ||
			fields.WeaponDamage != 0 || len(fields.EnchantEffects) > 0 ||
			row == spelldata.Nil || len(equipSpellUnsupported(enchant.ID, row, spellID)) > 0 {
			continue
		}
		spellIDs = append(spellIDs, spellID)
	}
	return spellIDs
}

func isNonZero(v float64) bool {
	return v != 0
}

// Why the potion builder cannot run the spell the consumable's first on-use row casts.
func onUseUnsupported(c *proto.Consumable) []string {
	for _, effect := range c.ItemEffects {
		if effect.GetOnUse() != nil {
			return shared.ConsumableUnsupported(spelldata.Find(effect.BuffId))
		}
	}
	return nil
}

// An empty file, or one whose every call is commented out, imports nothing: gen_db links sim/common,
// so an unused import in a file it just wrote breaks the build the next run needs.
func writeRegistrationFile(path, function string, groups []registrationGroup) error {
	var body strings.Builder
	live := false
	for i, group := range groups {
		switch {
		case group.title != "":
			fmt.Fprintf(&body, "\n\t// %s\n", group.title)
		case i > 0 && len(group.entries) > 0:
			body.WriteString("\n")
		}
		for _, e := range group.entries {
			args := fmt.Sprint(e.id)
			for _, arg := range e.args {
				args += fmt.Sprintf(", %d", arg)
			}
			if len(e.unsupported) > 0 {
				fmt.Fprintf(&body, "\t// unsupported: %s\n\t// shared.%s(%s) // %s\n", strings.Join(e.unsupported, "; "), e.constructor, args, e.name)
				continue
			}
			live = true
			fmt.Fprintf(&body, "\tshared.%s(%s) // %s\n", e.constructor, args, e.name)
		}
	}

	var out strings.Builder
	out.WriteString("// Code generated by tools/database/gen_consumable_registrations.go. DO NOT EDIT.\n\npackage forever\n\n")
	if live {
		out.WriteString("import \"github.com/wowsims/forever/sim/common/shared\"\n\n")
	}
	fmt.Fprintf(&out, "func %s() {\n%s}\n", function, body.String())

	formatted, err := format.Source([]byte(out.String()))
	if err != nil {
		return fmt.Errorf("generated %s is not valid Go: %w", path, err)
	}
	return os.WriteFile(path, formatted, 0644)
}
