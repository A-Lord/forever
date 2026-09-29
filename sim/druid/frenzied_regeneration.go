package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerFrenziedRegenerationSpell() {
	rank := spellData.FrenziedRegeneration.Highest()
	// The periodic trigger names no spell. The aura is what the rotation activates.
	druid.FrenziedRegenerationAura = druid.RegisterAura(spelldata.AuraConfig(rank))
	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		druid.FrenziedRegenerationAura.Activate(sim)
	}
	config.RelatedSelfBuff = druid.FrenziedRegenerationAura
	druid.FrenziedRegeneration = druid.RegisterSpell(config)
}
