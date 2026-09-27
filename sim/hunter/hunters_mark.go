package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var huntersMarkRank = spellData.HuntersMark.Highest()

// Effect 2: the ranged attack power every attacker gains against the marked enemy.
var huntersMarkRangedAttackPower = huntersMarkRank.Effect(dbcenums.A_RANGED_ATTACK_POWER_ATTACKER_BONUS, 0).BaseValue()

// The mark as the hunter's own cast. A_RANGED_ATTACK_POWER_ATTACKER_BONUS has no row in the parse
// table, so the bonus is attached by hand, exclusive with a raid's mark the way two marks are.
func (hunter *Hunter) registerHuntersMark() {
	auras := hunter.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		aura := target.RegisterAura(spelldata.AuraConfig(huntersMarkRank,
			spelldata.Label("Hunter's Mark-"+hunter.Label)))
		aura.Tag = "HuntersMark"
		aura.NewExclusiveEffect("HuntersMark", true, core.ExclusiveEffect{
			Priority: huntersMarkRangedAttackPower,
			OnGain: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
				ee.Aura.Unit.PseudoStats.BonusRangedAttackPower += ee.Priority
			},
			OnExpire: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
				ee.Aura.Unit.PseudoStats.BonusRangedAttackPower -= ee.Priority
			},
		})
		return aura
	})

	config := spelldata.SpellConfig(&hunter.Unit, huntersMarkRank, spelldata.Flags(core.SpellFlagAPL))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		auras.Get(target).Activate(sim)
	}
	config.RelatedAuraArrays = auras.ToMap()

	hunter.RegisterSpell(config)
}
