package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// TODO: Swipe gains 3% of the druid's attack power in game; its rows do not state it.
func (druid *Druid) registerSwipeBearSpell() {
	rank := spellData.Swipe.Highest()
	damageEffect := rank.DamageEffect()
	// TODO: Ingame research needed -- whether Swipe gains damage from attack power; this row states no
	// coefficient.
	base := damageEffect.Average()
	targets := int32(damageEffect.ChainTargets)
	if targets < 1 {
		targets = int32(rank.MaxTargets)
	}
	var results core.SpellResultSlice
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		cleave(sim, target, spell, targets, &results, func(target *core.Unit) *core.SpellResult {
			return spell.CalcDamage(sim, target, base*druid.rendAndTearMultiplier(target), spell.OutcomeMeleeSpecialHitAndCrit)
		})
	}
	druid.Swipe = druid.RegisterSpell(config)
}
