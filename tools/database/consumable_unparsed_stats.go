package database

import (
	"slices"

	"github.com/wowsims/forever/sim/common/forever"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// One rating feeds the physical and the spell percent alike, so a consumable granting both is listed
// at the larger of the two, once.
var percentRatings = []struct {
	percents   []stats.Stat
	rating     stats.Stat
	perPercent float64
}{
	{[]stats.Stat{stats.PhysicalHitPercent, stats.SpellHitPercent}, stats.HitRating, core.HitRatingPerHitPercent},
	{[]stats.Stat{stats.PhysicalCritPercent, stats.SpellCritPercent}, stats.CritRating, core.CritRatingPerCritPercent},
	{[]stats.Stat{stats.ExpertisePercent}, stats.ExpertiseRating, core.ExpertiseRatingPerExpertisePercent},
}

// Adds to the stats db.json lists for a consumable, which the settings offer it by, what its on-use
// grants that the stat parse leaves out, a percent at the rating one percent takes: the stats against
// creature types (Potion of Demon Slaying 9224's 11406: 180 attack power against Demons),
// A_MOD_HIT_CHANCE as physical hit (Fire-toasted Bun 23327's 29332: 2% as 20 hit rating),
// A_MOD_SPELL_CRIT_CHANCE as spell crit (Scroll of Arcane Power I 211957's 430952: 1% as 14 crit
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
			switch e.Aura {
			case dbcenums.A_MOD_HIT_CHANCE:
				percents[stats.PhysicalHitPercent] += e.Average(core.CharacterLevel)
			case dbcenums.A_MOD_SPELL_CRIT_CHANCE:
				percents[stats.SpellCritPercent] += e.Average(core.CharacterLevel)
			}
		}
		if slices.Contains(forever.NaturalFlasks, id) {
			bonus, err := forever.NaturalFlaskZoneBonusOf(row, row.EffectN(2).Average(core.CharacterLevel))
			if err != nil {
				return err
			}
			percents = percents.Add(bonus.Stats)
			granted[stats.HasteRating] += max(bonus.AttackHastePercent, bonus.CastHastePercent) * core.HasteRatingPerHastePercent
		}
		for _, r := range percentRatings {
			var percent float64
			for _, stat := range r.percents {
				percent = max(percent, percents[stat])
			}
			granted[r.rating] += percent * r.perPercent
		}

		if granted != (stats.Stats{}) {
			consumable.Stats = stats.FromProtoArray(consumable.Stats).Add(granted).ToProtoArray()
		}
	}
	return nil
}
