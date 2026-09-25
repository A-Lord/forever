package database

import (
	"slices"

	"github.com/wowsims/forever/sim/common/forever"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var percentRatings = []struct {
	percent, rating stats.Stat
	perPercent      float64
}{
	{stats.PhysicalHitPercent, stats.MeleeHitRating, core.PhysicalHitRatingPerHitPercent},
	{stats.SpellHitPercent, stats.SpellHitRating, core.SpellHitRatingPerHitPercent},
	{stats.PhysicalCritPercent, stats.MeleeCritRating, core.PhysicalCritRatingPerCritPercent},
	{stats.SpellCritPercent, stats.SpellCritRating, core.SpellCritRatingPerCritPercent},
	{stats.ExpertisePercent, stats.ExpertiseRating, core.ExpertiseRatingPerExpertisePercent},
}

// Adds to the stats db.json lists for a consumable, which the settings offer it by, what its on-use
// grants that the stat parse leaves out, a percent at the rating one percent takes: the stats against
// creature types (Potion of Demonslaying 9224's 11406: 185 attack power and 108 spell damage against
// Demons), A_MOD_HIT_CHANCE as physical hit (Fire-toasted Bun 23327's 29332: 2% as 20 melee hit
// rating), and a Natural flask's zone bonus, attack and casting speed as haste.
func AddUnparsedStats(db *WowDatabase) error {
	for id, consumable := range db.Consumables {
		if len(consumable.ItemEffects) == 0 {
			continue
		}
		row := spelldata.Find(consumable.ItemEffects[0].BuffId)

		var granted, percents stats.Stats
		for i := range row.Effects {
			e := &row.Effects[i]
			if e.Type != dbcenums.E_APPLY_AURA {
				continue
			}
			if stat, ok := spelldata.VersusStat(e.Aura); ok {
				granted[stat] += e.Average(core.CharacterLevel)
			}
			if e.Aura == dbcenums.A_MOD_HIT_CHANCE {
				percents[stats.PhysicalHitPercent] += e.Average(core.CharacterLevel)
			}
		}
		if slices.Contains(forever.NaturalFlasks, id) {
			bonus, err := forever.NaturalFlaskZoneBonusOf(row, row.EffectN(2).Average(core.CharacterLevel))
			if err != nil {
				return err
			}
			percents = percents.Add(bonus.Stats)
			granted[stats.MeleeHasteRating] += bonus.AttackHastePercent * core.PhysicalHasteRatingPerHastePercent
			granted[stats.SpellHasteRating] += bonus.CastHastePercent * core.SpellHasteRatingPerHastePercent
		}
		for _, r := range percentRatings {
			granted[r.rating] += percents[r.percent] * r.perPercent
		}

		if granted != (stats.Stats{}) {
			consumable.Stats = stats.FromProtoArray(consumable.Stats).Add(granted).ToProtoArray()
		}
	}
	return nil
}
