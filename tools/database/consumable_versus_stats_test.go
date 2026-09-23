package database

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestPotionOfDemonslayingListsItsVersusStats(t *testing.T) {
	db := NewWowDatabase()
	db.Consumables[9224] = &proto.Consumable{Id: 9224, ItemEffects: []*proto.ItemEffect{{BuffId: 11406}}}
	db.Consumables[13510] = &proto.Consumable{Id: 13510, ItemEffects: []*proto.ItemEffect{{BuffId: 17626}}}
	AddVersusStats(db)

	stats := db.Consumables[9224].Stats
	if stats[proto.Stat_StatAttackPower] != 185 || stats[proto.Stat_StatSpellDamage] != 108 {
		t.Errorf("9224 lists %v attack power and %v spell damage, want 185 and 108",
			stats[proto.Stat_StatAttackPower], stats[proto.Stat_StatSpellDamage])
	}
	if len(db.Consumables[13510].Stats) != 0 {
		t.Errorf("Flask of the Titans lists %v, want no versus stats", db.Consumables[13510].Stats)
	}
}
