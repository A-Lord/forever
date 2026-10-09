package database

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/overrides"
)

// The generated list holds every item in the database whose WeaponProcs row is a vanilla carry-over,
// once and in id order, each on its own line with the item's name, and leaves out a row of any other
// source and an item the database does not carry.
func TestVanillaPPMItemsListTheDatabasesVanillaRates(t *testing.T) {
	db := NewWowDatabase()
	for _, row := range overrides.WeaponProcs {
		db.Items[row.ItemID] = &proto.UIItem{Id: row.ItemID, Name: fmt.Sprintf("Item %d", row.ItemID)}
	}
	vanilla := slices.IndexFunc(overrides.WeaponProcs, func(row overrides.WeaponProc) bool { return row.Source == "vanilla-carryover" })
	if vanilla < 0 {
		t.Fatal("WeaponProcs holds no vanilla carry-over row")
	}
	dropped := overrides.WeaponProcs[vanilla].ItemID
	delete(db.Items, dropped)

	var want []int32
	for _, row := range overrides.WeaponProcs {
		if row.Source == "vanilla-carryover" && row.ItemID != dropped && !slices.Contains(want, row.ItemID) {
			want = append(want, row.ItemID)
		}
	}
	slices.Sort(want)

	items := vanillaPPMItems(db)
	var got []int32
	for _, item := range items {
		got = append(got, item.ItemID)
	}
	if !slices.Equal(got, want) {
		t.Errorf("listed %v, want %v", got, want)
	}
	for _, row := range overrides.WeaponProcs {
		if row.Source != "vanilla-carryover" && slices.Contains(got, row.ItemID) {
			t.Errorf("listed %d, whose rate is a %s row", row.ItemID, row.Source)
		}
	}
	// A trinket's rate lists like a weapon's.
	if !slices.Contains(got, 19287) {
		t.Errorf("Darkmoon Card: Heroism 19287 is not listed")
	}

	var rendered bytes.Buffer
	if err := renderMissingEffects(&rendered, db); err != nil {
		t.Fatal(err)
	}
	var block strings.Builder
	block.WriteString("export const VANILLA_PPM_ITEMS: number[] = [\n")
	for _, item := range items {
		fmt.Fprintf(&block, "\t%d, // %s\n", item.ItemID, item.Name)
	}
	block.WriteString("]\n")
	if !strings.Contains(rendered.String(), block.String()) {
		t.Errorf("the generated file holds no list of one id per line with its name:\n%s", rendered.String())
	}
}
