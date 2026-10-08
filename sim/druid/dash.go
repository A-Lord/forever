package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerDashSpell() {
	rank := spellData.Dash.Highest()
	druid.DashAura = druid.RegisterAura(spelldata.AuraConfig(rank))
	druid.DashAura.NewActiveMovementSpeedEffect(rank.EffectN(1).Percent())
	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		druid.DashAura.Activate(sim)
	}
	config.RelatedSelfBuff = druid.DashAura
	druid.Dash = druid.RegisterSpell(config)
}
