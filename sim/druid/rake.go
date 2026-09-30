package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerRakeSpell() {
	rank := spellData.Rake.Highest()
	// TODO: Next game build -- Rake gains damage from attack power; this row states no coefficient, so
	// the hit and the ticks are the row's amounts until one is measured and added as an override.
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.Dot = spelldata.DotConfig(rank, rank.PeriodicDamageEffect(), bleedAuraOpt())
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		druid.breakProwl(sim)
		initial := rank.DamageEffect().AverageAt(druid.level()) * druid.rendAndTearMultiplier(target)
		result := spell.CalcAndDealDamage(sim, target, initial, spell.OutcomeMeleeSpecialHitAndCrit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
			druid.awardComboPoints(sim, spell, rank)
		}
		druid.refundOnMiss(sim, spell, result)
	}
	druid.Rake = druid.RegisterSpell(config)
	druid.Rake.ShortName = "Rake"
}

func (druid *Druid) CurrentRakeCost() float64 {
	return druid.Rake.Cost.GetCurrentCost()
}
