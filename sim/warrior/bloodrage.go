package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var bloodrageRank = spellData.Bloodrage.Highest()
var bloodrageOverTime = spellData.BloodrageTriggered.Highest()
var bloodrageOverTimeTick = bloodrageOverTime.PeriodicEffect()

func (warrior *Warrior) registerBloodrage() {
	rageMetrics := warrior.NewRageMetrics(core.ActionID{SpellID: bloodrageRank.ID})
	healthCost := warrior.GetBaseStats()[stats.Health] * float64(bloodrageRank.Powers[0].CostPct) / 100
	improvedBloodrage := spellData.ImprovedBloodrage.MultiplierAt(warrior.Talents.ImprovedBloodrage)
	instantRage := spellData.Bloodrage.EffectAt(1).TenthsAt(1) * improvedBloodrage
	ragePerTick := bloodrageOverTimeTick.Tenths() * improvedBloodrage

	config := spelldata.SpellConfig(&warrior.Unit, bloodrageRank)

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		warrior.AddRage(sim, instantRage, rageMetrics)
		warrior.RemoveHealth(sim, healthCost)

		core.StartPeriodicAction(sim, core.PeriodicActionOptions{
			NumTicks: int(bloodrageOverTime.Duration() / bloodrageOverTimeTick.Period()),
			Period:   bloodrageOverTimeTick.Period(),
			OnAction: func(sim *core.Simulation) {
				warrior.AddRage(sim, ragePerTick, rageMetrics)
			},
		})
	}

	spell := warrior.RegisterSpell(config)

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			return warrior.CurrentRage() < 70
		},
	})
}
