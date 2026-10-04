package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerPrimalBiteSpell() {
	rank := spellData.PrimalBite.Highest()
	var results core.SpellResultSlice
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return druid.Talents.PrimalBite
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		hit := func(target *core.Unit) *core.SpellResult {
			return spell.CalcDamage(sim, target, druid.weaponSpecialDamage(sim, spell, target, rank), spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		}
		result := cleave(sim, target, spell, max(1, druid.berserkTargets()), &results, hit)
		druid.refundOnMiss(sim, spell, result)
	}
	druid.PrimalBite = druid.RegisterSpell(config)
}
