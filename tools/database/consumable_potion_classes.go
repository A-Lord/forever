package database

import (
	"slices"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Lists the classes of each potion, which the settings offer it under and the APL potion placeholders
// cast it by. A gain of mana, rage or energy is a resource potion, a heal or absorb a defensive one, and
// each stat it grants of the class the major cooldown manager files a buff to it under: armor,
// avoidance, health and resistances defensive, mana regen resource, anything else throughput. Runs after
// AddUnparsedStats, so the stats against creature types and the hit and crit percents count.
func AddPotionClasses(db *WowDatabase) {
	for _, consumable := range db.Consumables {
		consumable.PotionClasses = nil
		if consumable.Type != proto.ConsumableType_ConsumableTypePotion || len(consumable.ItemEffects) == 0 {
			continue
		}

		var classes []proto.PotionClass
		add := func(class proto.PotionClass) {
			if !slices.Contains(classes, class) {
				classes = append(classes, class)
			}
		}

		row := spelldata.Find(consumable.ItemEffects[0].BuffId)
		for i := range row.Effects {
			e := &row.Effects[i]
			switch {
			case e.Type == dbcenums.E_ENERGIZE, e.Type == dbcenums.E_ENERGIZE_PCT,
				e.Type == dbcenums.E_APPLY_AURA && e.Aura == dbcenums.A_PERIODIC_ENERGIZE:
				if shared.EnergizesABar(dbcenums.PowerType(e.Misc)) {
					add(proto.PotionClass_PotionClassResource)
				}
			case e.Type == dbcenums.E_HEAL, e.Type == dbcenums.E_HEAL_PCT,
				e.Type == dbcenums.E_APPLY_AURA && (e.Aura == dbcenums.A_PERIODIC_HEAL || e.Aura == dbcenums.A_SCHOOL_ABSORB):
				add(proto.PotionClass_PotionClassDefensive)
			}
		}

		for _, stat := range stats.FromProtoArray(consumable.Stats).GetBuffedStatTypes() {
			switch {
			case slices.Contains(core.SurvivalStats, stat):
				add(proto.PotionClass_PotionClassDefensive)
			case slices.Contains(core.ManaStats, stat):
				add(proto.PotionClass_PotionClassResource)
			default:
				add(proto.PotionClass_PotionClassThroughput)
			}
		}
		if slices.ContainsFunc(consumable.PseudoStats, func(v float64) bool { return v != 0 }) {
			add(proto.PotionClass_PotionClassThroughput)
		}

		slices.Sort(classes)
		consumable.PotionClasses = classes
	}
}
