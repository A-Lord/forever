package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerFerociousBiteSpell() {
	// The trainer list stops at rank 4. Rank 5 (31018) is in the store.
	// TODO: Ingame research needed -- whether rank 5 spell 31018 can be learned.
	rank := spellData.FerociousBite.ByID(22829)
	perEnergy := rank.EffectN(2).BaseValue() / 100
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ExtraCastCondition = druid.requiresComboPoints(rank)
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		druid.breakProwl(sim)
		cp := druid.spendComboPoints(sim, spell)
		extra := druid.CurrentEnergy()
		if extra > 0 {
			druid.SpendEnergy(sim, extra, spell.EnergyMetrics())
		}
		damage := rank.DamageEffect().RollAt(sim, druid.level()) + float64(rank.DamageEffect().PointsPerResource)*float64(cp)
		damage += extra * perEnergy
		damage += classicAPPerComboPointBite * float64(cp) * spell.MeleeAttackPower(target)
		damage *= druid.rendAndTearMultiplier(target)
		result := spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMeleeSpecialHitAndCrit)
		druid.refundOnMiss(sim, spell, result, rank)
	}
	druid.FerociousBite = druid.RegisterSpell(config)
}

func (druid *Druid) CurrentFerociousBiteCost() float64 {
	return druid.FerociousBite.Cost.GetCurrentCost()
}
