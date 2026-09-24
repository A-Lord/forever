package database

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Adds to the stats db.json lists for a consumable, which the settings offer it by, the hit chance its
// on-use applies (A_MOD_HIT_CHANCE), which the sim grants as physical hit percent, at the melee hit
// rating a percent takes: Fire-toasted Bun 23327 (29332) and Dark Desire 22237 (27723) list 2% as 20.
func AddHitChanceStats(db *WowDatabase) {
	for _, consumable := range db.Consumables {
		if len(consumable.ItemEffects) == 0 {
			continue
		}
		row := spelldata.Find(consumable.ItemEffects[0].BuffId)
		if row == nil {
			continue
		}
		var hit stats.Stats
		for i := range row.Effects {
			e := &row.Effects[i]
			if e.Type == dbcenums.E_APPLY_AURA && e.Aura == dbcenums.A_MOD_HIT_CHANCE {
				hit[stats.MeleeHitRating] += e.Average(core.CharacterLevel) * core.PhysicalHitRatingPerHitPercent
			}
		}
		if hit != (stats.Stats{}) {
			consumable.Stats = stats.FromProtoArray(consumable.Stats).Add(hit).ToProtoArray()
		}
	}
}
