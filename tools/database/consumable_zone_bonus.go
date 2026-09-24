package database

import (
	"github.com/wowsims/forever/sim/common/forever"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Adds the zone bonus of each Natural flask db.json ships to its stats, which the settings offer a
// consumable by: the bonus sim/common/forever applies, at the rating a percent of each takes, attack
// and casting speed as haste.
func AddNaturalFlaskZoneBonusStats(db *WowDatabase) error {
	for _, id := range forever.NaturalFlasks {
		consumable, ok := db.Consumables[id]
		if !ok || len(consumable.ItemEffects) == 0 {
			continue
		}
		row := spelldata.MustFind(consumable.ItemEffects[0].BuffId)
		bonus, err := forever.NaturalFlaskZoneBonusOf(row, row.EffectN(2).Average(core.CharacterLevel))
		if err != nil {
			return err
		}
		var ratings stats.Stats
		ratings[stats.MeleeHitRating] = bonus.Stats[stats.PhysicalHitPercent] * core.PhysicalHitRatingPerHitPercent
		ratings[stats.SpellHitRating] = bonus.Stats[stats.SpellHitPercent] * core.SpellHitRatingPerHitPercent
		ratings[stats.MeleeCritRating] = bonus.Stats[stats.PhysicalCritPercent] * core.PhysicalCritRatingPerCritPercent
		ratings[stats.SpellCritRating] = bonus.Stats[stats.SpellCritPercent] * core.SpellCritRatingPerCritPercent
		ratings[stats.ExpertiseRating] = bonus.Stats[stats.ExpertisePercent] * core.ExpertiseRatingPerExpertisePercent
		ratings[stats.MeleeHasteRating] = bonus.AttackHastePercent * core.PhysicalHasteRatingPerHastePercent
		ratings[stats.SpellHasteRating] = bonus.CastHastePercent * core.SpellHasteRatingPerHastePercent
		consumable.Stats = stats.FromProtoArray(consumable.Stats).Add(ratings).ToProtoArray()
	}
	return nil
}
