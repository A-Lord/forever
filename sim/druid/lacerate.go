package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerLacerateSpell() {
	rank := spellData.Lacerate.Highest()
	tick := rank.PeriodicDamageEffect()
	perStack := rank.EffectN(2).Percent()
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.Dot = spelldata.DotConfig(rank, tick, bleedAuraOpt())
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		damage := tick.AverageAt(druid.level()) + classicAPCoefficientLacerate*dot.Spell.MeleeAttackPower(target)
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, damage, rank.TickOutcome(dot))
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		dot := spell.Dot(target)
		stacks := int32(0)
		if dot.IsActive() {
			stacks = dot.GetStacks()
		}
		if maxStacks := dot.MaxStacks; maxStacks == 0 || stacks < maxStacks {
			stacks++
		}
		// The direct hit is weapon damage times the dummy percent per stack. The dot stays the row's amount.
		weapon := spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
		damage := weapon*perStack*float64(stacks) + classicAPCoefficientLacerate*spell.MeleeAttackPower(target)
		damage *= druid.rendAndTearMultiplier(target)
		result := spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		if result.Landed() {
			dot.Apply(sim)
			dot.SetStacks(sim, stacks)
		}
		druid.refundOnMiss(sim, spell, result, rank)
	}
	druid.Lacerate = druid.RegisterSpell(config)
	druid.Lacerate.ShortName = "Lacerate"
}
