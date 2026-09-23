package database

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
)

func keepConsumableLists(t *testing.T) {
	allow, deny, overrides := ConsumableAllowList, ConsumableDenyList, ConsumableOverrides
	t.Cleanup(func() {
		ConsumableAllowList, ConsumableDenyList, ConsumableOverrides = allow, deny, overrides
	})
}

func TestAnAllowListedOrOverriddenConsumableShipsUnsupported(t *testing.T) {
	inRepositoryRoot(t)
	keepConsumableLists(t)

	blastingCharge := dbc.GetDBC().Consumables[221325]
	routing := RouteConsumable(&blastingCharge, dbc.GetDBC())
	if routing.Supported() {
		t.Fatal("221325 routes as supported, so this proves nothing")
	}
	if ShipsConsumable(routing) {
		t.Error("221325 ships while nothing names it")
	}

	ConsumableAllowList = append(slices.Clone(ConsumableAllowList), 221325)
	if !ShipsConsumable(routing) {
		t.Error("221325 is dropped although ConsumableAllowList names it")
	}

	ConsumableAllowList = slices.DeleteFunc(slices.Clone(ConsumableAllowList), func(id int32) bool { return id == 221325 })
	ConsumableOverrides = append(slices.Clone(ConsumableOverrides), &proto.Consumable{Id: 221325, CooldownDuration: 60})
	if !ShipsConsumable(routing) {
		t.Error("221325 is dropped although ConsumableOverrides names it")
	}
}

func TestTheDenyListRemovesAConsumable(t *testing.T) {
	keepConsumableLists(t)

	sapper := &proto.Consumable{Id: 10646, Name: "Goblin Sapper Charge"}
	if !KeepConsumable(sapper) {
		t.Fatal("10646 is filtered out while no list names it")
	}

	ConsumableDenyList = append(slices.Clone(ConsumableDenyList), 10646)
	if KeepConsumable(sapper) {
		t.Error("10646 is kept although ConsumableDenyList names it")
	}

	ConsumableAllowList = append(slices.Clone(ConsumableAllowList), 10646)
	if !KeepConsumable(sapper) {
		t.Error("10646 is filtered out although ConsumableAllowList names it, which wins over the deny list")
	}
}

func TestACustomConsumableTheClientLacksIsKept(t *testing.T) {
	custom := &proto.Consumable{Id: 999001, Name: "Custom Elixir", Type: proto.ConsumableType_ConsumableTypeBattleElixir}
	if !KeepConsumable(custom) {
		t.Error("a consumable ConsumableOverrides adds is filtered out")
	}
}
