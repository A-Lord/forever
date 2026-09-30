package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerFerociousBiteSpell() {
	// The trainer list stops at rank 4. Rank 5 (31018) is in the store.
	// TODO: Ingame research needed -- whether rank 5 spell 31018 can be learned.
	rank := spellData.FerociousBite.ByID(22829)
	damageEffect := rank.DamageEffect()
	perEnergy := rank.EffectN(2).BaseValue() / 100
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ExtraCastCondition = druid.requiresComboPoints(rank)
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		druid.breakProwl(sim)
		cp := float64(druid.ComboPoints())
		extra := druid.CurrentEnergy()
		damage := damageEffect.RollAt(sim, druid.level()) + float64(damageEffect.PointsPerResource)*cp
		damage += extra * perEnergy
		damage += damageEffect.APCoeff() * cp * spell.MeleeAttackPower(target)
		damage *= druid.rendAndTearMultiplier(target)
		// Combo points and the extra energy are spent on a landed bite only, before the damage is
		// dealt, so nothing the hit triggers changes what is spent.
		result := spell.CalcDamage(sim, target, damage, spell.OutcomeMeleeSpecialHitAndCrit)
		if result.Landed() {
			druid.spendComboPoints(sim, spell)
			if extra > 0 {
				druid.SpendEnergy(sim, extra, spell.EnergyMetrics())
			}
		}
		spell.DealDamage(sim, result)
		druid.refundOnMiss(sim, spell, result)
	}
	druid.FerociousBite = druid.RegisterSpell(config)
}

func (druid *Druid) CurrentFerociousBiteCost() float64 {
	return druid.FerociousBite.Cost.GetCurrentCost()
}
