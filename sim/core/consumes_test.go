package core

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestSelectedConsumableIDs(t *testing.T) {
	consumes := &proto.ConsumesSpec{
		PotId:         13444,
		Potions:       []int32{13444, 13446},
		FlaskId:       13510,
		FoodId:        238637,
		ExplosiveId:   18588,
		MhImbueId:     20749,
		OhImbueId:     12404,
		ConjuredItems: []int32{7676},
		ExtraIds:      []int32{10646, 10309, 10305, 5206},
	}

	got := selectedConsumableIDs(consumes, true)
	want := []int32{13444, 13510, 238637, 18588, 12404, 20749, 10646, 10309, 10305, 5206, 13446, 7676}
	if !slices.Equal(got, want) {
		t.Errorf("selected %v, want %v", got, want)
	}

	bareMainHand := selectedConsumableIDs(consumes, false)
	if slices.Contains(bareMainHand, 20749) {
		t.Errorf("selected the main-hand imbue while the main hand carries no enchant: %v", bareMainHand)
	}
}

func TestSelectedConsumableIDsKeepTheFirstOfEachGroup(t *testing.T) {
	const agility, mongoose, sages, ungrouped, juju, scroll = 990401, 990402, 990403, 990404, 990405, 990406
	addToDatabase(&proto.SimDatabase{Consumables: []*proto.Consumable{
		{Id: agility, Type: proto.ConsumableType_ConsumableTypeElixir, GroupKey: "elixir 29/1"},
		{Id: mongoose, Type: proto.ConsumableType_ConsumableTypeElixir, GroupKey: "elixir 29/1"},
		{Id: sages, Type: proto.ConsumableType_ConsumableTypeElixir, GroupKey: "elixir 29/3"},
		{Id: ungrouped, Type: proto.ConsumableType_ConsumableTypeElixir},
		{Id: juju, GroupKey: "elixir 29/3"},
		{Id: scroll, Type: proto.ConsumableType_ConsumableTypeScroll, GroupKey: "scroll 29/1"},
	}})

	got := selectedConsumableIDs(&proto.ConsumesSpec{ElixirIds: []int32{agility, mongoose, sages, ungrouped}, ExtraIds: []int32{juju, scroll}}, false)
	if want := []int32{agility, sages, ungrouped, scroll}; !slices.Equal(got, want) {
		t.Errorf("selected %v, want %v", got, want)
	}
}

func TestHandWrittenConsumableEffectReplacesDataDriven(t *testing.T) {
	const majorHealingPotion = 13446
	unregister := func() {
		delete(consumableEffects, majorHealingPotion)
		delete(dataDrivenConsumableEffects, majorHealingPotion)
	}
	t.Cleanup(unregister)

	for _, handFirst := range []bool{true, false} {
		applied := ""
		hand := func() { NewConsumableEffect(majorHealingPotion, func(Agent) { applied = "hand" }) }
		dataDriven := func() { NewDataDrivenConsumableEffect(majorHealingPotion, func(Agent) { applied = "rows" }) }
		if handFirst {
			hand()
			dataDriven()
		} else {
			dataDriven()
			hand()
		}

		if !HasConsumableEffect(majorHealingPotion) {
			t.Fatalf("hand first %v: %d has no effect registered", handFirst, majorHealingPotion)
		}
		apply, _ := consumableEffect(majorHealingPotion)
		apply(nil)
		if applied != "hand" {
			t.Errorf("hand first %v: the %s effect applied, want the hand-written one", handFirst, applied)
		}
		unregister()
	}
}

func TestASecondEffectOfTheSameOriginPanics(t *testing.T) {
	const majorHealingPotion = 13446
	for name, register := range map[string]func(int32, ApplyEffect){
		"hand-written": NewConsumableEffect,
		"data-driven":  NewDataDrivenConsumableEffect,
	} {
		t.Run(name, func(t *testing.T) {
			t.Cleanup(func() {
				delete(consumableEffects, majorHealingPotion)
				delete(dataDrivenConsumableEffects, majorHealingPotion)
			})

			register(majorHealingPotion, func(Agent) {})
			defer func() {
				if recover() == nil {
					t.Errorf("a second %s effect for one consumable registered without a panic", name)
				}
			}()
			register(majorHealingPotion, func(Agent) {})
		})
	}
}

// ui/sim/proto/consumes_migration.ts reads the fields ConsumesSpec dropped in api version 2 by these
// names, and by these numbers where a binary payload can still carry them, so they must stay reserved.
func TestRemovedConsumesFieldsStayReserved(t *testing.T) {
	descriptor := (&proto.ConsumesSpec{}).ProtoReflect().Descriptor()
	for name, number := range map[protoreflect.Name]protoreflect.FieldNumber{
		"goblinSapper": 8, "bogling_root": 9, "scroll_agi": 12, "scroll_str": 13, "scroll_int": 14,
		"scroll_spi": 15, "scroll_arm": 16, "superSapper": 0, "drums_id": 0, "pet_scroll_agi": 0,
		"pet_scroll_str": 0, "nightmare_seed": 0, "bloodthistle": 0, "pet_food_id": 0, "pet_extra_ids": 20,
		"battle_elixir_id": 3, "guardian_elixir_id": 4,
	} {
		if !descriptor.ReservedNames().Has(name) || (number != 0 && !descriptor.ReservedRanges().Has(number)) {
			t.Errorf("ConsumesSpec does not reserve %s = %d", name, number)
		}
	}
}
