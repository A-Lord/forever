package database

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestNaturalFlasksListTheirZoneBonus(t *testing.T) {
	flask := func(id, spellID int32) *proto.Consumable {
		stats := make([]float64, len(proto.Stat_name))
		stats[proto.Stat_StatStamina] = 60
		return &proto.Consumable{Id: id, Stats: stats, ItemEffects: []*proto.ItemEffect{{BuffId: spellID}}}
	}
	db := NewWowDatabase()
	for id, spellID := range map[int32]int32{274273: 1293740, 274274: 1293741, 274275: 1293742, 274276: 1293743} {
		db.Consumables[id] = flask(id, spellID)
	}
	if err := AddNaturalFlaskZoneBonusStats(db); err != nil {
		t.Fatal(err)
	}

	for id, want := range map[int32]map[proto.Stat]float64{
		274273: {proto.Stat_StatMeleeHitRating: 50, proto.Stat_StatSpellHitRating: 50},
		274274: {proto.Stat_StatMeleeCritRating: 56, proto.Stat_StatSpellCritRating: 56},
		274275: {proto.Stat_StatExpertiseRating: 50},
		274276: {proto.Stat_StatMeleeHasteRating: 50, proto.Stat_StatSpellHasteRating: 50},
	} {
		want[proto.Stat_StatStamina] = 60
		for stat, value := range db.Consumables[id].Stats {
			if value != want[proto.Stat(stat)] {
				t.Errorf("%d lists %v %v, want %v", id, proto.Stat(stat), value, want[proto.Stat(stat)])
			}
		}
	}
}
