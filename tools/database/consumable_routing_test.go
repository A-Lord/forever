package database

// Every consumable the catalogue selects, and the shape its on-use spell routes to. One that fits no
// shape is left out of db.json and named in unsupported_consumables.txt with the reason, beside the
// count routed to each shape, so a routing that moves is reviewed rather than discovered. Only the
// damage and heal shapes have builders so far. That file is a local baseline, not tracked: a
// checkout without it writes it on the first run.
//
// Reads the consumables the last gen_db run captured and the committed store.

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/tools/database/dbc"
)

const unsupportedConsumablesPath = "tools/database/unsupported_consumables.txt"

func TestConsumablesRouteByTheirSpell(t *testing.T) {
	inRepositoryRoot(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		itemID int
		want   ConsumableKind
	}{
		{10646, ConsumableDamage},           // Goblin Sapper Charge: 13241
		{11566, ConsumableDamage},           // Crystal Charge, a Quest-class item: 15239
		{215127, ConsumableDamage},          // High-Yield Radiation Bomb: 435963
		{13446, ConsumableHeal},             // Major Healing Potion: 17534
		{13444, ConsumableEnergize},         // Major Mana Potion: 17531
		{13510, ConsumableStatAura},         // Flask of the Titans: 17626
		{954, ConsumableStatAura},           // Scroll of Strength, cast on a friendly target: 8118
		{20749, ConsumableTemporaryEnchant}, // Brilliant Wizard Oil: 25122, enchant 2628
		{13457, ConsumableAbsorb},           // Greater Fire Protection Potion: 17543
	} {
		consumable, ok := instance.Consumables[tc.itemID]
		if !ok {
			t.Errorf("%d is not in the catalogue", tc.itemID)
			continue
		}
		if routing := RouteConsumable(&consumable, instance); !routing.Supported() || routing.Kind != tc.want {
			t.Errorf("%d %s routes as %q (%s), want %q", tc.itemID, consumable.Name, routing.Kind, routing.Reason(), tc.want)
		}
	}

	blastingCharge := instance.Consumables[221325]
	if RouteConsumable(&blastingCharge, instance).Supported() {
		t.Error("221325 routes as supported, but its spell 446504 only opens a lock")
	}
}

func TestEveryConsumableIsRoutedOrListed(t *testing.T) {
	inRepositoryRoot(t)
	instance := dbc.GetDBC()

	var out strings.Builder
	out.WriteString("# Every consumable the catalogue selects whose on-use spell fits no shape the sim builds, as\n")
	out.WriteString("# \"<item id> <spell id> <item name>: <reason>\", after the count routed to each shape. Written by\n")
	out.WriteString("# TestEveryConsumableIsRoutedOrListed in tools/database. Regenerate with\n")
	out.WriteString("#   UPDATE_UNSUPPORTED_CONSUMABLES=1 go test ./tools/database/ -run TestEveryConsumableIsRoutedOrListed -count=1\n\n")

	routed := map[ConsumableKind]int{}
	var listed strings.Builder
	for _, id := range slices.Sorted(maps.Keys(instance.Consumables)) {
		consumable := instance.Consumables[id]
		routing := RouteConsumable(&consumable, instance)
		if routing.Supported() {
			routed[routing.Kind]++
			continue
		}
		fmt.Fprintf(&listed, "%d %d %s: %s\n", id, routing.SpellID, consumable.Name, routing.Reason())
	}

	for _, kind := range ConsumableKinds {
		fmt.Fprintf(&out, "# %s: %d\n", kind, routed[kind])
	}
	out.WriteString("\n")
	out.WriteString(listed.String())

	if routed[ConsumableDamage] == 0 || routed[ConsumableStatAura] == 0 {
		t.Fatal("no consumable routes as damage or as a stat aura, so this proves nothing")
	}

	committed, err := os.ReadFile(unsupportedConsumablesPath)
	if os.IsNotExist(err) || os.Getenv("UPDATE_UNSUPPORTED_CONSUMABLES") != "" {
		if err := os.WriteFile(unsupportedConsumablesPath, []byte(out.String()), 0644); err != nil {
			t.Fatalf("%v", err)
		}
		t.Logf("wrote the baseline %s", unsupportedConsumablesPath)
		return
	}
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(committed) == out.String() {
		return
	}

	for _, line := range changedProcLines(string(committed), out.String()) {
		t.Error(line)
	}
	t.Errorf("%s is not what the rows say; once the moves above are the intended ones, rewrite the baseline:\n"+
		"  UPDATE_UNSUPPORTED_CONSUMABLES=1 go test ./tools/database/ -run TestEveryConsumableIsRoutedOrListed -count=1",
		unsupportedConsumablesPath)
}
