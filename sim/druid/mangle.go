package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerMangleAuras() {
}

// TODO: Primal Bite generates roughly twice the threat its rows state; measure the threat multiplier in game.
func (druid *Druid) registerPrimalBiteSpell() {
	rank := spellData.PrimalBite.Highest()
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return druid.Talents.PrimalBite
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		base := druid.weaponSpecialDamage(sim, spell, target, rank)
		extra := druid.berserkExtraTargets()
		if extra <= 0 {
			result := spell.CalcAndDealDamage(sim, target, base, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			druid.refundOnMiss(sim, spell, result, rank)
			return
		}
		results := spell.CalcCleaveDamage(sim, target, 1+extra, base, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		for _, result := range results {
			spell.DealDamage(sim, result)
		}
		if len(results) > 0 {
			druid.refundOnMiss(sim, spell, results[0], rank)
		}
	}
	druid.PrimalBite = druid.RegisterSpell(config)
	druid.MangleBear = druid.PrimalBite
}
