package database

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Adds to the stats db.json lists for a consumable, which the settings offer it by, what its on-use
// grants against creature types: Potion of Demonslaying 9224 (11406) lists none, and sim/common
// registers it by hand.
func AddVersusStats(db *WowDatabase) {
	for _, consumable := range db.Consumables {
		if len(consumable.ItemEffects) == 0 {
			continue
		}
		granted := shared.VersusStats(spelldata.Find(consumable.ItemEffects[0].BuffId), core.CharacterLevel)
		if granted != (stats.Stats{}) {
			consumable.Stats = stats.FromProtoArray(consumable.Stats).Add(granted).ToProtoArray()
		}
	}
}
