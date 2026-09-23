package database

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Adds to the stats db.json lists for a consumable, which the settings offer it by, the attack power
// and spell damage its on-use grants against one creature type: Potion of Demonslaying 9224 (11406)
// lists none, and sim/common registers it by hand.
func AddVersusStats(db *WowDatabase) {
	for _, consumable := range db.Consumables {
		if len(consumable.ItemEffects) == 0 {
			continue
		}
		row := spelldata.Find(consumable.ItemEffects[0].BuffId)
		for i := range row.Effects {
			e := &row.Effects[i]
			var stat proto.Stat
			switch e.Aura {
			case dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS:
				stat = proto.Stat_StatAttackPower
			case dbcenums.A_MOD_FLAT_SPELL_DAMAGE_VERSUS:
				stat = proto.Stat_StatSpellDamage
			default:
				continue
			}
			for len(consumable.Stats) <= int(stat) {
				consumable.Stats = append(consumable.Stats, 0)
			}
			consumable.Stats[stat] += e.Average(core.CharacterLevel)
		}
	}
}
