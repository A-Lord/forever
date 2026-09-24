package database

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
)

func shippedConsumable(t *testing.T, instance *dbc.DBC, id int) *proto.Consumable {
	t.Helper()
	catalogued, ok := instance.Consumables[id]
	if !ok {
		t.Fatalf("%d is not in the catalogue", id)
	}
	routing := RouteConsumable(&catalogued, instance)
	if !ShipsConsumable(routing) {
		t.Fatalf("%d %s does not ship", id, catalogued.Name)
	}
	return ShippedConsumable(&catalogued, routing, instance)
}

func TestAConsumableTakesItsExclusiveGroup(t *testing.T) {
	inRepositoryRoot(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		name       string
		id         int
		key, label string
	}{
		{"Elixir of the Mongoose, A_MOD_STAT Agility then crit", 13452, "elixir 29/1", "Agility"},
		{"Elixir of the Sages, A_MOD_STAT Intellect then Spirit", 13447, "elixir 29/3", "Intellect"},
		{"Arcane Elixir, A_MOD_DAMAGE_DONE every school", 9155, "elixir 13/126", "Spell Power"},
		{"Elixir of Fire Power, A_MOD_DAMAGE_DONE fire", 6373, "elixir 13/4", "Fire Spell Power"},
		{"Gift of Arthas, A_MOD_RESISTANCE shadow", 9088, "elixir 22/32", "Shadow Resistance"},
		{"Elixir of Greater Strength, A_MOD_STAT Strength", 9206, "elixir 29/0", "Strength"},
		{"Juju Power joins the Strength elixirs", 12451, "elixir 29/0", "Strength"},
		{"Winterfall Firewater, A_MOD_ATTACK_POWER", 12820, "elixir 99/0", "Attack Power"},
		{"Juju Might joins the Attack Power elixirs", 12460, "elixir 99/0", "Attack Power"},
		{"Scroll of Agility IV, the scrolls' Agility", 10309, "scroll 29/1", "Agility"},
		{"Scroll of Agility, the same group as its rank IV", 3012, "scroll 29/1", "Agility"},
		{"Scroll of Protection IV, A_MOD_RESISTANCE armor", 10305, "scroll 22/1", "Armor"},
		{"Spirit of Zanza", 20079, "Zanza", "Zanza"},
		{"R.O.I.D.S., category 103 in its rows, Zanza by the override", 8410, "Zanza", "Zanza"},
		{"Fire-toasted Bun", 23327, "Hit food", "Hit food"},
		{"Dark Desire", 22237, "Hit food", "Hit food"},
		{"Flask of the Titans stacks with everything", 13510, "", ""},
		{"Mighty Rage Potion stacks with everything", 13442, "", ""},
	} {
		if got := shippedConsumable(t, instance, tc.id); got.GroupKey != tc.key || got.GroupName != tc.label {
			t.Errorf("%s: group %q %q, want %q %q", tc.name, got.GroupKey, got.GroupName, tc.key, tc.label)
		}
	}
}

// Every elixir db.json ships has a group, and of the other consumables only the Blasted Lands buffs
// are grouped by their category in their rows: category 103's 1 h lasts as long as each buff.
func TestTheRowsGroupEveryElixirAndOnlyTheBlastedLandsByCategory(t *testing.T) {
	inRepositoryRoot(t)
	instance := dbc.GetDBC()

	byCategory := map[int32]string{}
	for _, id := range slices.Sorted(maps.Keys(instance.Consumables)) {
		catalogued := instance.Consumables[id]
		routing := RouteConsumable(&catalogued, instance)
		if !ShipsConsumable(routing) {
			continue
		}
		consumable := ShippedConsumable(&catalogued, routing, instance)
		group := rowGroupOf(&catalogued, consumable, routing, instance)
		if consumable.Type == proto.ConsumableType_ConsumableTypeElixir && group.Key == "" {
			t.Errorf("elixir %d %s applies no aura to group it by", id, catalogued.Name)
		}
		if strings.HasPrefix(group.Key, "category ") {
			byCategory[int32(id)] = group.Key
		}
	}
	want := map[int32]string{8410: "category 103", 8411: "category 103", 8412: "category 103", 8423: "category 103", 8424: "category 103"}
	if !maps.Equal(byCategory, want) {
		t.Errorf("grouped by category %v, want %v", byCategory, want)
	}
}

func TestEveryExclusiveGroupOverrideShips(t *testing.T) {
	inRepositoryRoot(t)
	instance := dbc.GetDBC()

	seen := map[int32]bool{}
	for _, group := range ConsumableExclusiveGroups {
		if group.Joins.Key != "" && group.Joins.Name != group.Name {
			t.Errorf("group %q joins %q, labelled %q", group.Name, group.Joins.Key, group.Joins.Name)
		}
		for _, id := range group.Items {
			if seen[id] {
				t.Errorf("%d is in more than one group", id)
			}
			seen[id] = true
			shippedConsumable(t, instance, int(id))
		}
	}
}
