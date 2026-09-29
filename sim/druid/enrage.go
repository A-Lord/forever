package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerEnrageSpell() {
	rank := spellData.Enrage.Highest()
	periodic := rank.PeriodicEffect()
	druid.EnrageAura = druid.RegisterAura(spelldata.AuraConfig(rank))
	druid.EnrageAura.AttachPeriodicAction(core.PeriodicActionOptions{
		Period:   periodic.Period(),
		NumTicks: int(rank.Duration() / periodic.Period()),
		OnAction: func(sim *core.Simulation) {
			druid.AddRage(sim, periodic.Tenths(), druid.enrageMetrics())
		},
	})
	if druid.BearFormAura != nil {
		druid.BearFormAura.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
			druid.EnrageAura.Deactivate(sim)
		})
	}
	if druid.DireBearFormAura != nil {
		druid.DireBearFormAura.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
			druid.EnrageAura.Deactivate(sim)
		})
	}
	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		instant := rank.EnergizeEffect()
		if instant != spelldata.NilEffect {
			druid.AddRage(sim, instant.Tenths(), druid.enrageMetrics())
		}
		druid.EnrageAura.Activate(sim)
	}
	config.RelatedSelfBuff = druid.EnrageAura
	druid.Enrage = druid.RegisterSpell(config)
}

func (druid *Druid) enrageMetrics() *core.ResourceMetrics {
	if druid.enrageMetric == nil {
		druid.enrageMetric = druid.NewRageMetrics(core.ActionID{SpellID: spellData.Enrage.Highest().ID})
	}
	return druid.enrageMetric
}
