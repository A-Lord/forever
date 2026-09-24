package database

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestHitFoodListsItsHitChance(t *testing.T) {
	db := NewWowDatabase()
	db.Consumables[23327] = &proto.Consumable{Id: 23327, ItemEffects: []*proto.ItemEffect{{BuffId: 29332}}}
	db.Consumables[22237] = &proto.Consumable{Id: 22237, ItemEffects: []*proto.ItemEffect{{BuffId: 27723}}}
	db.Consumables[13510] = &proto.Consumable{Id: 13510, ItemEffects: []*proto.ItemEffect{{BuffId: 17626}}}
	AddHitChanceStats(db)

	for _, id := range []int32{23327, 22237} {
		for stat, value := range db.Consumables[id].Stats {
			want := 0.0
			if proto.Stat(stat) == proto.Stat_StatMeleeHitRating {
				want = 20
			}
			if value != want {
				t.Errorf("%d lists %v %v, want %v", id, proto.Stat(stat), value, want)
			}
		}
		if len(db.Consumables[id].Stats) == 0 {
			t.Errorf("%d lists no stats, want 20 melee hit rating", id)
		}
	}
	if len(db.Consumables[13510].Stats) != 0 {
		t.Errorf("Flask of the Titans lists %v, want no hit", db.Consumables[13510].Stats)
	}
}
