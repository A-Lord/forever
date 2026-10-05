package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerRakeSpell() {
	rank := spellData.Rake.Highest()
	// TODO: Ingame research needed -- the direct hit's attack power share.
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	tick := rank.PeriodicDamageEffect()
	config.Dot = spelldata.DotConfig(rank, tick, bleedAuraOpt())
	// TODO: Ingame research needed -- whether Rake snapshots attack power.
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		damage := tick.AverageAt(druid.level()) + tick.APCoeff()*dot.Spell.MeleeAttackPower(target)
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, damage, rank.TickOutcome(dot))
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		initial := rank.DamageEffect().AverageAt(druid.level()) * druid.rendAndTearMultiplier(target)
		result := spell.CalcAndDealDamage(sim, target, initial, spell.OutcomeMeleeSpecialHitAndCrit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
			druid.awardComboPoints(sim, spell, rank)
		}
		druid.refundOnMiss(sim, spell, result)
	}
	druid.Rake = druid.RegisterSpell(config)
}

func (druid *Druid) CurrentRakeCost() float64 {
	return druid.Rake.Cost.GetCurrentCost()
}
