package dbc

import (
	"maps"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// 22780 Biznicks 247x128 Accurascope, bound to bows, guns and crossbows, with its hit traded for an
// A_MOD_RATING of 20 points on mask, in the spell store and in the client tables alike.
func rangedWeaponRating(t *testing.T, mask int32) int {
	const id = 22780
	row := spelldata.Find(id)
	if row.EquipClass != dbcenums.ITEM_CLASS_WEAPON || row.EquipSubclass != 262156 {
		t.Fatalf("22780 is bound to class %d subclass %d, want bows, guns and crossbows", row.EquipClass, row.EquipSubclass)
	}
	effects := row.Effects
	row.Effects = []spelldata.Effect{{SpellID: id, Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_RATING, Misc: mask, BasePoints: 20,
		Target: [2]dbcenums.ImplicitTarget{dbcenums.TARGET_UNIT_CASTER}}}

	instance := dbcInstance
	dbcInstance = NewDBC()
	dbcInstance.Spells[id] = Spell{ID: id, EquippedItemClass: int32(ITEM_CLASS_WEAPON), EquippedItemSubclass: 262156}
	dbcInstance.SpellEffects[id] = map[int]SpellEffect{0: {SpellID: id, EffectType: dbcenums.E_APPLY_AURA, EffectAura: dbcenums.A_MOD_RATING,
		EffectBasePoints: 20, EffectMiscValues: []int{int(mask), 0}, ImplicitTargets: []int{int(dbcenums.TARGET_UNIT_CASTER), 0}}}
	dbcInstance.indexSpellEffects()

	t.Cleanup(func() {
		row.Effects = effects
		dbcInstance = instance
	})
	return id
}

// A rating bound to ranged weapons folds into an item and into an enchant as the ranged percent of each
// hit, crit and haste rating its mask names, the one spelldata.RangedRatingPercents reads and the parse
// applies, and as no rating stat.
func TestRangedWeaponRatingFolds(t *testing.T) {
	hit, crit, haste := 20/core.HitRatingPerHitPercent, 20/core.CritRatingPerCritPercent, 20/core.HasteRatingPerHastePercent
	for _, c := range []struct {
		name string
		mask int32
		want map[proto.PseudoStat]float64
	}{
		{"hit on every channel", 32 | 64 | 128, map[proto.PseudoStat]float64{proto.PseudoStat_PseudoStatRangedHitPercent: hit}},
		{"melee and ranged hit", 32 | 64, map[proto.PseudoStat]float64{proto.PseudoStat_PseudoStatRangedHitPercent: hit}},
		{"crit on every channel", 256 | 512 | 1024, map[proto.PseudoStat]float64{proto.PseudoStat_PseudoStatRangedCritPercent: crit}},
		{"melee, ranged and cast haste", 131072 | 262144 | 524288, map[proto.PseudoStat]float64{proto.PseudoStat_PseudoStatRangedHastePercent: haste}},
		{"defense, melee hit and spell crit", 2 | 32 | 1024, map[proto.PseudoStat]float64{
			proto.PseudoStat_PseudoStatRangedHitPercent:  hit,
			proto.PseudoStat_PseudoStatRangedCritPercent: crit,
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			id := rangedWeaponRating(t, c.mask)

			row := spelldata.Find(int32(id))
			parsed := map[proto.PseudoStat]float64{}
			for _, rating := range spelldata.RangedRatingPercents(row, &row.Effects[0], 20) {
				parsed[rangedPercentStats[rating.Stat]] = rating.Percent
			}
			if !maps.Equal(parsed, c.want) {
				t.Errorf("the parse reads %v, want %v", parsed, c.want)
			}

			var equipStats stats.Stats
			equipPseudoStats := make([]float64, stats.PseudoStatsLen)
			AddEquipSpellStats(&equipStats, equipPseudoStats, id)
			checkRangedRatingFold(t, "equip spell", equipStats, equipPseudoStats, c.want)

			if props := buildBaseStatScalingProps(id, id); len(props.Stats) != 0 {
				t.Errorf("item: stats %v, want none", props.Stats)
			}

			var enchantStats stats.Stats
			enchantPseudoStats := make([]float64, stats.PseudoStatsLen)
			processEnchantmentEffects([]int{ITEM_ENCHANTMENT_EQUIP_SPELL}, []int{id}, []int{0}, &enchantStats, enchantPseudoStats, true)
			checkRangedRatingFold(t, "enchant", enchantStats, enchantPseudoStats, c.want)
		})
	}
}

func checkRangedRatingFold(t *testing.T, what string, got stats.Stats, gotPseudoStats []float64, want map[proto.PseudoStat]float64) {
	t.Helper()
	if got != (stats.Stats{}) {
		t.Errorf("%s: stats %v, want none", what, got)
	}
	for pseudoStat, value := range gotPseudoStats {
		if value != want[proto.PseudoStat(pseudoStat)] {
			t.Errorf("%s: %s = %v, want %v", what, proto.PseudoStat(pseudoStat), value, want[proto.PseudoStat(pseudoStat)])
		}
	}
}
