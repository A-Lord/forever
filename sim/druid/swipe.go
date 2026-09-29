package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// TODO: Swipe gains 3% of the druid's attack power in game; its rows do not state it.
func (druid *Druid) registerSwipeBearSpell() {
	rank := spellData.Swipe.Highest()
	damageEffect := rank.DamageEffect()
	// The row states no attack power coefficient. Swipe does not scale with attack power.
	base := damageEffect.Average()
	targets := int32(damageEffect.ChainTargets)
	if targets < 1 {
		targets = int32(rank.MaxTargets)
	}
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		results := spell.CalcCleaveDamage(sim, target, targets, base*druid.rendAndTearMultiplier(target), spell.OutcomeMeleeSpecialHitAndCrit)
		for _, result := range results {
			spell.DealDamage(sim, result)
		}
	}
	druid.Swipe = druid.RegisterSpell(config)
}
