package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerProwlSpell() {
	rank := spellData.Prowl.Highest()
	druid.ProwlAura = druid.RegisterAura(spelldata.AuraConfig(rank))
	// The row's attribute word 0 cancels auto attack (0x100000), which core does not read: no swing
	// lands while prowling, and the opener that ends it starts them again. The row's not-in-combat bit
	// keeps the cast before the pull, so the pause cannot run past the encounter's clock.
	druid.ProwlAura.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) {
		druid.AutoAttacks.PauseMeleeBy(sim, core.NeverExpires)
	})
	druid.ProwlAura.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
		druid.AutoAttacks.ResumeMeleeAt(sim, 0, 0)
	})
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
