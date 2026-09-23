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
