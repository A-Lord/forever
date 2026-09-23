package core

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
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
		PetFoodId:     33874,
	}

	got := selectedConsumableIDs(consumes, &proto.PartyBuffs{})
	want := []int32{13444, 13510, 238637, 18588, 12404, 20749, 13446, 7676}
	if !slices.Equal(got, want) {
		t.Errorf("selected %v, want %v", got, want)
	}

	underWindfury := selectedConsumableIDs(consumes, &proto.PartyBuffs{WindfuryTotem: proto.TristateEffect_TristateEffectRegular})
	if slices.Contains(underWindfury, 20749) {
		t.Errorf("selected the main-hand imbue under Windfury Totem: %v", underWindfury)
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
