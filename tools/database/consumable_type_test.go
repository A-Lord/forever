package database

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
)

func TestAConsumableIsTypedByTheCategoryItsOnUseShares(t *testing.T) {
	inRepositoryRoot(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		name     string
		id       int
		category int32
		want     proto.ConsumableType
	}{
		{"Major Mana Potion, category 4 on the item", 13444, dbc.POTION_SPELL_CATEGORY, proto.ConsumableType_ConsumableTypePotion},
		{"Minor Spellblasting Potion, category 4 on its spell 1251919 alone", 250932, dbc.POTION_SPELL_CATEGORY, proto.ConsumableType_ConsumableTypePotion},
		{"Potion of Demonslaying, an Elixir-subclass item whose spell 11406 shares category 4", 9224, dbc.POTION_SPELL_CATEGORY, proto.ConsumableType_ConsumableTypePotion},
		{"Thistle Tea, category 1153 on the item over its spell's 60", 7676, dbc.CONJURED_SPELL_CATEGORY, proto.ConsumableType_ConsumableTypeConjured},
		{"Demonic Rune, not a Consumable-class item", 12662, dbc.CONJURED_SPELL_CATEGORY, proto.ConsumableType_ConsumableTypeConjured},
		{"Keg of Thunderbrew, an Other-subclass drink", 1262, dbc.DRINK_SPELL_CATEGORY, proto.ConsumableType_ConsumableTypeFood},
		{"Mageblood Elixir, a Potion-subclass item on category 79", 20007, 79, proto.ConsumableType_ConsumableTypePotion},
		{"Gift of Arthas, an elixir on category 79", 9088, 79, proto.ConsumableType_ConsumableTypeElixir},
	} {
		catalogued, ok := instance.Consumables[tc.id]
		if !ok {
			t.Fatalf("%s: %d is not in the catalogue", tc.name, tc.id)
		}
		if got := catalogued.ToProto(); got.CategoryId != tc.category || got.Type != tc.want {
			t.Errorf("%s: category %d, type %v; want %d and %v", tc.name, got.CategoryId, got.Type, tc.category, tc.want)
		}
	}
}

// An elixir's group is the aura type and misc value of the first aura its on-use spell applies, and
// every elixir db.json ships has one.
func TestAnElixirIsGroupedByItsFirstAura(t *testing.T) {
	inRepositoryRoot(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		name       string
		id         int
		aura, misc int32
	}{
		{"Elixir of the Mongoose, A_MOD_STAT Agility then crit", 13452, 29, 1},
		{"Elixir of the Sages, A_MOD_STAT Intellect then Spirit", 13447, 29, 3},
		{"Arcane Elixir, A_MOD_DAMAGE_DONE every school", 9155, 13, 126},
		{"Elixir of Fire Power, A_MOD_DAMAGE_DONE fire", 6373, 13, 4},
		{"Gift of Arthas, A_MOD_RESISTANCE shadow", 9088, 22, 32},
	} {
		catalogued, ok := instance.Consumables[tc.id]
		if !ok {
			t.Fatalf("%s: %d is not in the catalogue", tc.name, tc.id)
		}
		if got := catalogued.ToProto(); got.ElixirGroupAura != tc.aura || got.ElixirGroupMisc != tc.misc {
			t.Errorf("%s: group %d/%d, want %d/%d", tc.name, got.ElixirGroupAura, got.ElixirGroupMisc, tc.aura, tc.misc)
		}
	}

	for id, catalogued := range instance.Consumables {
		if !ShipsConsumable(RouteConsumable(&catalogued, instance)) {
			continue
		}
		if got := catalogued.ToProto(); got.Type == proto.ConsumableType_ConsumableTypeElixir && got.ElixirGroupAura == 0 {
			t.Errorf("elixir %d %s applies no aura to group it by", id, catalogued.Name)
		}
	}
}

// The explosive picker offers what shares the explosives' category 24 and routes as damage.
func TestOnlyADamagingExplosiveIsTypedAnExplosive(t *testing.T) {
	inRepositoryRoot(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		name      string
		id        int
		explosive bool
	}{
		{"Goblin Sapper Charge, 13241 deals damage", 10646, true},
		{"Bag of Marbles, category 24, 5917 lowers the target's hit chance", 1191, false},
		{"Large Rope Net, category 24, 8312 roots", 835, false},
		{"EZ-Splode Blasting Charge, a Trade Goods explosive on no category", 221325, false},
	} {
		catalogued, ok := instance.Consumables[tc.id]
		if !ok {
			t.Fatalf("%s: %d is not in the catalogue", tc.name, tc.id)
		}
		got := ShippedConsumable(&catalogued, RouteConsumable(&catalogued, instance)).Type
		if (got == proto.ConsumableType_ConsumableTypeExplosive) != tc.explosive {
			t.Errorf("%s: typed %v", tc.name, got)
		}
	}
}
