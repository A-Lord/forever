package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerProwlSpell() {
	rank := spellData.Prowl.Highest()
	druid.ProwlAura = druid.RegisterAura(spelldata.AuraConfig(rank))
	if druid.CatFormAura != nil {
		druid.CatFormAura.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
			druid.breakProwl(sim)
		})
	}
	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return !druid.ProwlAura.IsActive()
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		druid.ProwlAura.Activate(sim)
	}
	druid.Prowl = druid.RegisterSpell(config)
}
