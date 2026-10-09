package database

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/common/shared"
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

// Consumables whose first aura is the same share one group, and a different first aura, another
// stat or another school, is another group.
func TestAConsumableTakesItsExclusiveGroup(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	groupNames := map[string]string{}
	for _, tc := range []struct {
		name string
		ids  []int
	}{
		{"Strength: Elixir of Greater Strength, Juju Power, Scroll of Strength IV", []int{9206, 12451, 10310}},
		{"Agility: Elixir of the Mongoose (Agility then crit), Scroll of Agility IV and Scroll of Agility", []int{13452, 10309, 3012}},
		{"Intellect: Elixir of the Sages (Intellect then Spirit), Scroll of Intellect IV", []int{13447, 10308}},
		{"Armor: Elixir of Greater Defense, Scroll of Protection IV", []int{13445, 10305}},
		{"Attack Power: Winterfall Firewater, Juju Might", []int{12820, 12460}},
		{"Arcane Elixir, every school", []int{9155}},
		{"Elixir of Fire Power, fire alone", []int{6373}},
		{"Gift of Arthas, shadow resistance", []int{9088}},
		{"Scroll of Spirit IV", []int{10306}},
		{"Scroll of Stamina IV, which no elixir opens on", []int{10307}},
		{"Scroll of Minor Evocation, A_PERIODIC_ENERGIZE, which no elixir opens on", []int{213564}},
	} {
		first := shippedConsumable(t, instance, tc.ids[0])
		if first.GroupKey == "" || first.GroupName == "" {
			t.Errorf("%s: %d takes group %q %q, want one", tc.name, tc.ids[0], first.GroupKey, first.GroupName)
			continue
		}
		for _, id := range tc.ids[1:] {
			if got := shippedConsumable(t, instance, id); got.GroupKey != first.GroupKey || got.GroupName != first.GroupName {
				t.Errorf("%s: %d takes group %q %q, want %d's %q %q", tc.name, id, got.GroupKey, got.GroupName, tc.ids[0], first.GroupKey, first.GroupName)
			}
		}
		if other, taken := groupNames[first.GroupKey]; taken {
			t.Errorf("%s: group %q is also %s's", tc.name, first.GroupKey, other)
		}
		groupNames[first.GroupKey] = tc.name
	}

	for _, tc := range []struct {
		name       string
		id         int
		key, label string
	}{
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

// Every elixir db.json ships has a group. A consumable grouped by the category in its rows is no
// elixir or scroll, and its category is the one whose cooldown outlasts each buff it puts up.
func TestTheRowsGroupEveryElixirAndByCategoryOnlyWhereTheCategoryOutlastsTheBuff(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	byCategory := 0
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
		if !strings.HasPrefix(group.Key, "category ") {
			continue
		}
		byCategory++
		if consumable.Type == proto.ConsumableType_ConsumableTypeElixir || consumable.Type == proto.ConsumableType_ConsumableTypeScroll {
			t.Errorf("%v %d %s is grouped by its category, want by its aura", consumable.Type, id, catalogued.Name)
		}
		category := shared.OneAtATimeCategory(consumable.ItemEffects[0])
		if want := fmt.Sprintf("category %d", category); category <= 0 || group.Key != want {
			t.Errorf("%d %s is grouped as %q, want %q", id, catalogued.Name, group.Key, want)
		}
	}
	if byCategory == 0 {
		t.Error("no consumable is grouped by its category")
	}
}

func TestEveryExclusiveGroupOverrideShips(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	seen := map[int32]bool{}
	for _, group := range ConsumableExclusiveGroups {
		for _, id := range group.Items {
			if seen[id] {
				t.Errorf("%d is in more than one group", id)
			}
			seen[id] = true
			shippedConsumable(t, instance, int(id))
		}
	}
}
