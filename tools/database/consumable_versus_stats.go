package database

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Adds to the stats db.json lists for a consumable, which the settings offer it by, what its on-use
// grants against creature types, which the stat parse leaves out: Potion of Demonslaying 9224 (11406)
// lists its 185 attack power and 108 spell damage against Demons.
func AddVersusStats(db *WowDatabase) {
	for _, consumable := range db.Consumables {
		if len(consumable.ItemEffects) == 0 {
			continue
		}
		granted := versusStats(spelldata.Find(consumable.ItemEffects[0].BuffId), core.CharacterLevel)
		if granted != (stats.Stats{}) {
			consumable.Stats = stats.FromProtoArray(consumable.Stats).Add(granted).ToProtoArray()
		}
	}
}

// The stats the row's versus auras grant, whatever creature types they name.
func versusStats(row *spelldata.Spell, level int32) stats.Stats {
	var granted stats.Stats
	for i := range row.Effects {
		e := &row.Effects[i]
		if stat, ok := spelldata.VersusStat(e.Aura); ok && e.Type == dbcenums.E_APPLY_AURA {
			granted[stat] += e.Average(level)
		}
	}
	return granted
}
