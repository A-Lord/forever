package core

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestSelectedConsumableIDs(t *testing.T) {
	const oil = 990410
	addToDatabase(&proto.SimDatabase{Consumables: []*proto.Consumable{
		{Id: oil, Type: proto.ConsumableType_ConsumableTypeImbue, TemporaryEnchant: &proto.SimEnchant{EffectId: 990420}},
	}})
	consumes := &proto.ConsumesSpec{
		PotId:         13444,
		Potions:       []int32{13444, 13446},
		FlaskId:       13510,
		FoodId:        20452,
		ExplosiveId:   18588,
		MhImbueId:     oil,
		OhImbueId:     12404,
		ConjuredItems: []int32{7676},
		ExtraIds:      []int32{10646, 10309, 10305, 5206},
	}

	got := selectedConsumableIDs(consumes, Item{TempEnchant: 990420})
	want := []int32{13444, 13510, 20452, 18588, 12404, oil, 10646, 10309, 10305, 5206, 13446, 7676}
	if !slices.Equal(got, want) {
		t.Errorf("selected %v, want %v", got, want)
	}

	bareMainHand := selectedConsumableIDs(consumes, Item{})
	if slices.Contains(bareMainHand, oil) {
		t.Errorf("selected the main-hand imbue while the main hand carries no enchant: %v", bareMainHand)
	}
}

func TestSelectedConsumableIDsTakeAStoneAndAPoisonOnOneHand(t *testing.T) {
	const stone, poison = 990411, 990412
	addToDatabase(&proto.SimDatabase{Consumables: []*proto.Consumable{
		{Id: stone, Type: proto.ConsumableType_ConsumableTypeImbue, TemporaryEnchant: &proto.SimEnchant{EffectId: 990421}},
		{Id: poison, Type: proto.ConsumableType_ConsumableTypeWeaponImbue, TemporaryEnchant: &proto.SimEnchant{EffectId: 990422}},
	}})
	consumes := &proto.ConsumesSpec{MhImbueId: stone, MhWeaponImbueId: poison}

	if got := selectedConsumableIDs(consumes, Item{TempEnchant: 990421, WeaponImbue: 990422}); !slices.Equal(got, []int32{stone, poison}) {
		t.Errorf("selected %v, want the stone and the poison", got)
	}
	if got := selectedConsumableIDs(consumes, Item{WeaponImbue: 990422}); !slices.Equal(got, []int32{poison}) {
		t.Errorf("selected %v, want the poison alone on a main hand the stone does not take", got)
	}
}

func TestSelectedConsumableIDsKeepTheFirstOfEachGroup(t *testing.T) {
	const agility, mongoose, sages, ungrouped, juju, scroll = 990401, 990402, 990403, 990404, 990405, 990406
	addToDatabase(&proto.SimDatabase{Consumables: []*proto.Consumable{
		{Id: agility, Type: proto.ConsumableType_ConsumableTypeElixir, GroupKey: "29/1"},
		{Id: mongoose, Type: proto.ConsumableType_ConsumableTypeElixir, GroupKey: "29/1"},
		{Id: sages, Type: proto.ConsumableType_ConsumableTypeElixir, GroupKey: "29/3"},
		{Id: ungrouped, Type: proto.ConsumableType_ConsumableTypeElixir},
		{Id: juju, GroupKey: "29/3"},
		{Id: scroll, Type: proto.ConsumableType_ConsumableTypeScroll, GroupKey: "29/2"},
	}})

	got := selectedConsumableIDs(&proto.ConsumesSpec{ElixirIds: []int32{agility, mongoose, sages, ungrouped}, ExtraIds: []int32{juju, scroll}}, Item{})
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
		dataDriven := func() { NewSpellDataConsumableEffect(majorHealingPotion, func(Agent) { applied = "rows" }) }
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
		"data-driven":  NewSpellDataConsumableEffect,
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
