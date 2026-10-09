package database

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func TestUnparsedStatsListVersusStatsChannelPercentsAndZoneBonus(t *testing.T) {
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
	db.Consumables[211954] = &proto.Consumable{Id: 211954, ItemEffects: []*proto.ItemEffect{{BuffId: 430949}}}
	db.Consumables[211957] = &proto.Consumable{Id: 211957, ItemEffects: []*proto.ItemEffect{{BuffId: 430952}}}
	db.Consumables[13452] = &proto.Consumable{Id: 13452, ItemEffects: []*proto.ItemEffect{{BuffId: 17538}}}
	db.Consumables[250337] = &proto.Consumable{Id: 250337, ItemEffects: []*proto.ItemEffect{{BuffId: 1250940}}}
	for id, spellID := range map[int32]int32{274273: 1293740, 274274: 1293741, 274275: 1293742, 274276: 1293743} {
		db.Consumables[id] = flask(id, spellID)
	}
	if err := AddUnparsedStats(db); err != nil {
		t.Fatal(err)
	}

	amount := func(spellID int32, aura dbcenums.EffectAuraType) float64 {
		return spelldata.MustFind(spellID).FirstAura(aura).Average()
	}
	zoneBonus := func(spellID int32) float64 {
		return spelldata.MustFind(spellID).EffectN(2).Average()
	}
	meleeAndRangedHit := func(spellID int32) map[proto.PseudoStat]float64 {
		hit := amount(spellID, dbcenums.A_MOD_HIT_CHANCE)
		return map[proto.PseudoStat]float64{proto.PseudoStat_PseudoStatMeleeHitPercent: hit, proto.PseudoStat_PseudoStatRangedHitPercent: hit}
	}
	allCrit := func(spellID int32) map[proto.PseudoStat]float64 {
		crit := amount(spellID, dbcenums.A_MOD_CRIT_PCT)
		return map[proto.PseudoStat]float64{
			proto.PseudoStat_PseudoStatMeleeCritPercent: crit, proto.PseudoStat_PseudoStatRangedCritPercent: crit, proto.PseudoStat_PseudoStatSpellCritPercent: crit,
		}
	}
	for id, want := range map[int32]struct {
		stats       map[proto.Stat]float64
		pseudoStats map[proto.PseudoStat]float64
	}{
		9224:   {stats: map[proto.Stat]float64{proto.Stat_StatAttackPower: amount(11406, dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS)}},
		23327:  {pseudoStats: meleeAndRangedHit(29332)},
		22237:  {pseudoStats: meleeAndRangedHit(27723)},
		211954: {pseudoStats: map[proto.PseudoStat]float64{proto.PseudoStat_PseudoStatSpellHitPercent: amount(430949, dbcenums.A_MOD_SPELL_HIT_CHANCE)}},
		211957: {pseudoStats: map[proto.PseudoStat]float64{proto.PseudoStat_PseudoStatSpellCritPercent: amount(430952, dbcenums.A_MOD_SPELL_CRIT_CHANCE)}},
		13452:  {pseudoStats: allCrit(17538)},
		250337: {pseudoStats: allCrit(1250940)},
		274273: {stats: map[proto.Stat]float64{proto.Stat_StatStamina: 60, proto.Stat_StatHitRating: zoneBonus(1293740) * core.HitRatingPerHitPercent}},
		274274: {stats: map[proto.Stat]float64{proto.Stat_StatStamina: 60, proto.Stat_StatCritRating: zoneBonus(1293741) * core.CritRatingPerCritPercent}},
		274275: {stats: map[proto.Stat]float64{proto.Stat_StatStamina: 60, proto.Stat_StatExpertiseRating: zoneBonus(1293742) * core.ExpertiseRatingPerExpertisePercent}},
		274276: {stats: map[proto.Stat]float64{proto.Stat_StatStamina: 60, proto.Stat_StatHasteRating: zoneBonus(1293743) * core.HasteRatingPerHastePercent}},
	} {
		consumable := db.Consumables[id]
		if (len(consumable.Stats) == 0) != (len(want.stats) == 0) {
			t.Errorf("%d lists stats %v, want %v", id, consumable.Stats, want.stats)
		}
		for stat, value := range consumable.Stats {
			if value != want.stats[proto.Stat(stat)] {
				t.Errorf("%d lists %v %v, want %v", id, proto.Stat(stat), value, want.stats[proto.Stat(stat)])
			}
		}
		if (len(consumable.PseudoStats) == 0) != (len(want.pseudoStats) == 0) {
			t.Errorf("%d lists pseudo stats %v, want %v", id, consumable.PseudoStats, want.pseudoStats)
		}
		for stat, value := range consumable.PseudoStats {
			if value != want.pseudoStats[proto.PseudoStat(stat)] {
				t.Errorf("%d lists %v %v, want %v", id, proto.PseudoStat(stat), value, want.pseudoStats[proto.PseudoStat(stat)])
			}
		}
	}
	if titans := db.Consumables[13510]; len(titans.Stats) != 0 || len(titans.PseudoStats) != 0 {
		t.Errorf("Flask of the Titans lists %v %v, want nothing added", titans.Stats, titans.PseudoStats)
	}
}
