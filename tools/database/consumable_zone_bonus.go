package database

import (
	"fmt"

	"github.com/wowsims/forever/sim/common/forever"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Adds the zone bonus of each Natural flask db.json ships to its stats, which the settings offer a
// consumable by: the percent of the flask's A_DUMMY at the rating a percent of what the auras after
// it name takes, attack and casting speed as haste. sim/common/forever applies the bonus by hand.
func AddNaturalFlaskZoneBonusStats(db *WowDatabase) error {
	for _, id := range forever.NaturalFlasks {
		consumable, ok := db.Consumables[id]
		if !ok || len(consumable.ItemEffects) == 0 {
			continue
		}
		row := spelldata.MustFind(consumable.ItemEffects[0].BuffId)
		percent := row.EffectN(2).Average(core.CharacterLevel)
		add := func(stat proto.Stat, perPercent float64) {
			for len(consumable.Stats) <= int(stat) {
				consumable.Stats = append(consumable.Stats, 0)
			}
			consumable.Stats[stat] += percent * perPercent
		}
		for _, e := range row.Effects[2:] {
			switch e.Aura {
			case dbcenums.A_MOD_HIT_CHANCE:
				add(proto.Stat_StatMeleeHitRating, core.PhysicalHitRatingPerHitPercent)
			case dbcenums.A_MOD_SPELL_HIT_CHANCE:
				add(proto.Stat_StatSpellHitRating, core.SpellHitRatingPerHitPercent)
			case dbcenums.A_MOD_CRIT_PCT:
				add(proto.Stat_StatMeleeCritRating, core.PhysicalCritRatingPerCritPercent)
				add(proto.Stat_StatSpellCritRating, core.SpellCritRatingPerCritPercent)
			case dbcenums.A_MOD_EXPERTISE:
				add(proto.Stat_StatExpertiseRating, 4*core.ExpertisePerQuarterPercentReduction)
			case dbcenums.A_MOD_MELEE_RANGED_HASTE_2:
				add(proto.Stat_StatMeleeHasteRating, core.PhysicalHasteRatingPerHastePercent)
			case dbcenums.A_MOD_CASTING_SPEED_NOT_STACK:
				add(proto.Stat_StatSpellHasteRating, core.SpellHasteRatingPerHastePercent)
			default:
				return fmt.Errorf("%d %s: no zone bonus stat for %v", row.ID, row.Name, e.Aura)
			}
		}
	}
	return nil
}
