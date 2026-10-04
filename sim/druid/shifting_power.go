package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerShiftingPowerSpell() {
	rank := spellData.ShiftingPower.Highest()
	energy := rank.EnergizeEffect().Average()
	metrics := druid.NewEnergyMetrics(core.ActionID{SpellID: rank.ID})
	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return druid.Talents.ShiftingPower
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		druid.breakProwl(sim)
		druid.AddEnergy(sim, energy+druid.WolfsheadEnergyBonus, metrics)
	}
	druid.ShiftingPower = druid.RegisterSpell(config)
}
