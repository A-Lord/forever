package database

import (
	"slices"

	"github.com/wowsims/forever/sim/common/forever"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// One rating feeds the physical and the spell percent alike, so a consumable granting both is listed
// at the rating of the larger of the two, once. A consumable granting one of them is listed by that
// channel's pseudo stats: physical as melee and ranged, spell as spell.
var percentRatings = []struct {
	physical, spell stats.Stat
	rating          stats.Stat
	perPercent      float64
	melee, ranged   proto.PseudoStat
	spellPseudo     proto.PseudoStat
}{
	{stats.PhysicalHitPercent, stats.SpellHitPercent, stats.HitRating, core.HitRatingPerHitPercent,
		proto.PseudoStat_PseudoStatMeleeHitPercent, proto.PseudoStat_PseudoStatRangedHitPercent, proto.PseudoStat_PseudoStatSpellHitPercent},
	{stats.PhysicalCritPercent, stats.SpellCritPercent, stats.CritRating, core.CritRatingPerCritPercent,
		proto.PseudoStat_PseudoStatMeleeCritPercent, proto.PseudoStat_PseudoStatRangedCritPercent, proto.PseudoStat_PseudoStatSpellCritPercent},
}

// Adds to the stats and pseudo stats db.json lists for a consumable, which the settings offer it by,
// what its on-use grants that the stat parse leaves out: the stats against creature types (Potion of
// Demon Slaying 9224's 11406: 180 attack power against Demons), the hit and crit of one channel
// (Fire-toasted Bun 23327's 29332: A_MOD_HIT_CHANCE 2 as 2% melee and ranged hit), an all-channel crit
// (Elixir of the Mongoose 13452's 17538: A_MOD_CRIT_PCT 2 as 2% melee, ranged and spell crit), and a
// Natural flask's zone bonus at the rating one percent takes, attack and casting speed as haste.
func AddUnparsedStats(db *WowDatabase) error {
	for id, consumable := range db.Consumables {
		if len(consumable.ItemEffects) == 0 {
			continue
		}
		row := spelldata.Find(consumable.ItemEffects[0].BuffId)

		var granted, percents stats.Stats
		var allCrit float64
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
			case dbcenums.A_MOD_SPELL_HIT_CHANCE:
				percents[stats.SpellHitPercent] += e.Average(core.CharacterLevel)
			case dbcenums.A_MOD_WEAPON_CRIT_PERCENT:
				percents[stats.PhysicalCritPercent] += e.Average(core.CharacterLevel)
			case dbcenums.A_MOD_SPELL_CRIT_CHANCE:
				percents[stats.SpellCritPercent] += e.Average(core.CharacterLevel)
			case dbcenums.A_MOD_CRIT_PCT:
				allCrit += e.Average(core.CharacterLevel)
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
		granted[stats.ExpertiseRating] += percents[stats.ExpertisePercent] * core.ExpertiseRatingPerExpertisePercent

		pseudoStats := make([]float64, stats.PseudoStatsLen)
		for _, r := range percentRatings {
			physical, spell := percents[r.physical], percents[r.spell]
			switch {
			case physical > 0 && spell > 0:
				granted[r.rating] += max(physical, spell) * r.perPercent
			case physical > 0:
				pseudoStats[r.melee] += physical
				pseudoStats[r.ranged] += physical
			case spell > 0:
				pseudoStats[r.spellPseudo] += spell
			}
		}
		if allCrit != 0 {
			pseudoStats[proto.PseudoStat_PseudoStatMeleeCritPercent] += allCrit
			pseudoStats[proto.PseudoStat_PseudoStatRangedCritPercent] += allCrit
			pseudoStats[proto.PseudoStat_PseudoStatSpellCritPercent] += allCrit
		}

		if granted != (stats.Stats{}) {
			consumable.Stats = stats.FromProtoArray(consumable.Stats).Add(granted).ToProtoArray()
		}
		if slices.ContainsFunc(pseudoStats, func(v float64) bool { return v != 0 }) {
			consumable.PseudoStats = pseudoStats
		}
	}
	return nil
}
