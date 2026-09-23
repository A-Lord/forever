package database

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestOnUseCategoryFallsBackToTheSpell(t *testing.T) {
	onUse := func(spellID, categoryID int32) *proto.Consumable {
		return &proto.Consumable{ItemEffects: []*proto.ItemEffect{{BuffId: spellID,
			Effect: &proto.ItemEffect_OnUse{OnUse: &proto.OnUseEffect{CategoryId: categoryID}}}}}
	}

	for _, tc := range []struct {
		name       string
		consumable *proto.Consumable
		want       int32
	}{
		{"Major Mana Potion 13444, category 4 on the item", onUse(17531, 4), PotionCategory},
		{"Thistle Tea 7676, category 1153 on the item over its spell's 60", onUse(9512, 1153), ConjuredCategory},
		{"Minor Spellblasting Potion 250932, category 4 on its spell 1251919 alone", onUse(1251919, 0), PotionCategory},
		{"a consumable with no on-use row", &proto.Consumable{}, 0},
	} {
		if got := OnUseCategory(tc.consumable); got != tc.want {
			t.Errorf("%s: category %d, want %d", tc.name, got, tc.want)
		}
	}
}
