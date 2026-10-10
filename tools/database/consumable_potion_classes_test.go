package database

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func TestPotionClassesFollowTheirEffectsAndStats(t *testing.T) {
	potion := func(id, spellID int32, granted stats.Stats) *proto.Consumable {
		return &proto.Consumable{Id: id, Type: proto.ConsumableType_ConsumableTypePotion, Stats: granted.ToProtoArray(),
			ItemEffects: []*proto.ItemEffect{{BuffId: spellID}}}
	}
	const (
		resource   = proto.PotionClass_PotionClassResource
		defensive  = proto.PotionClass_PotionClassDefensive
		throughput = proto.PotionClass_PotionClassThroughput
	)
	db := NewWowDatabase()
	db.Consumables[13444] = potion(13444, 17531, stats.Stats{})                                          // Major Mana Potion: E_ENERGIZE mana
	db.Consumables[18253] = potion(18253, 22729, stats.Stats{})                                          // Major Rejuvenation Potion: a heal and mana
	db.Consumables[13457] = potion(13457, 17543, stats.Stats{})                                          // Greater Fire Protection Potion: A_SCHOOL_ABSORB
	db.Consumables[13442] = potion(13442, 17528, stats.Stats{stats.Strength: 60})                        // Mighty Rage Potion: rage and Strength
	db.Consumables[13455] = potion(13455, 17540, stats.Stats{stats.Armor: 2000})                         // Greater Stoneshield Potion
	db.Consumables[250937] = potion(250937, 1251932, stats.Stats{stats.SpellDamage: 45})                 // Major Spellblasting Potion
	db.Consumables[13510] = &proto.Consumable{Id: 13510, Type: proto.ConsumableType_ConsumableTypeFlask, // Flask of the Titans
		Stats: stats.Stats{stats.Health: 400}.ToProtoArray(), ItemEffects: []*proto.ItemEffect{{BuffId: 17626}}}
	AddPotionClasses(db)

	for id, want := range map[int32][]proto.PotionClass{
		13444:  {resource},
		18253:  {resource, defensive},
		13457:  {defensive},
		13442:  {resource, throughput},
		13455:  {defensive},
		250937: {throughput},
		13510:  nil,
	} {
		if got := db.Consumables[id].PotionClasses; !slices.Equal(got, want) {
			t.Errorf("%d: classes %v, want %v", id, got, want)
		}
	}
}
