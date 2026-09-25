package database

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestUnparsedStatsListVersusStatsHitChanceAndZoneBonus(t *testing.T) {
	flask := func(id, spellID int32) *proto.Consumable {
		stats := make([]float64, len(proto.Stat_name))
		stats[proto.Stat_StatStamina] = 60
		return &proto.Consumable{Id: id, Stats: stats, ItemEffects: []*proto.ItemEffect{{BuffId: spellID}}}
	}
	db := NewWowDatabase()
	db.Consumables[9224] = &proto.Consumable{Id: 9224, ItemEffects: []*proto.ItemEffect{{BuffId: 11406}}}
	db.Consumables[23327] = &proto.Consumable{Id: 23327, ItemEffects: []*proto.ItemEffect{{BuffId: 29332}}}
	db.Consumables[22237] = &proto.Consumable{Id: 22237, ItemEffects: []*proto.ItemEffect{{BuffId: 27723}}}
	db.Consumables[13510] = &proto.Consumable{Id: 13510, ItemEffects: []*proto.ItemEffect{{BuffId: 17626}}}
	for id, spellID := range map[int32]int32{274273: 1293740, 274274: 1293741, 274275: 1293742, 274276: 1293743} {
		db.Consumables[id] = flask(id, spellID)
	}
	if err := AddUnparsedStats(db); err != nil {
		t.Fatal(err)
	}

	for id, want := range map[int32]map[proto.Stat]float64{
		9224:   {proto.Stat_StatAttackPower: 185, proto.Stat_StatSpellDamage: 108},
		23327:  {proto.Stat_StatMeleeHitRating: 20},
		22237:  {proto.Stat_StatMeleeHitRating: 20},
		274273: {proto.Stat_StatStamina: 60, proto.Stat_StatMeleeHitRating: 50, proto.Stat_StatSpellHitRating: 50},
		274274: {proto.Stat_StatStamina: 60, proto.Stat_StatMeleeCritRating: 56, proto.Stat_StatSpellCritRating: 56},
		274275: {proto.Stat_StatStamina: 60, proto.Stat_StatExpertiseRating: 50},
		274276: {proto.Stat_StatStamina: 60, proto.Stat_StatMeleeHasteRating: 50, proto.Stat_StatSpellHasteRating: 50},
	} {
		listed := db.Consumables[id].Stats
		if len(listed) == 0 {
			t.Errorf("%d lists no stats, want %v", id, want)
		}
		for stat, value := range listed {
			if value != want[proto.Stat(stat)] {
				t.Errorf("%d lists %v %v, want %v", id, proto.Stat(stat), value, want[proto.Stat(stat)])
			}
		}
	}
	if len(db.Consumables[13510].Stats) != 0 {
		t.Errorf("Flask of the Titans lists %v, want nothing added", db.Consumables[13510].Stats)
	}
}
