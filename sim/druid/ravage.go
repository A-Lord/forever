package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerRavageSpell() {
	rank := spellData.Ravage.Highest()
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	// "Must be prowling and behind the target" is the tooltip's; the row states Cat Form only.
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return druid.ProwlAura != nil && druid.ProwlAura.IsActive() && druid.BehindTarget()
	}
	// No Active Defense (Attributes 0x200000): it misses or lands, never dodged, parried or blocked.
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, druid.weaponSpecialDamage(sim, spell, target, rank), spell.OutcomeMeleeSpecialNoBlockDodgeParry)
		if result.Landed() {
			druid.awardComboPoints(sim, spell, rank)
		}
		druid.refundOnMiss(sim, spell, result)
	}
	druid.Ravage = druid.RegisterSpell(config)
}
