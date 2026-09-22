package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var demoralizingShoutRank = spellData.DemoralizingShout.Highest()

// TODO: The core Demoralizing Shout aura still takes Booming Voice and Improved Demoralizing
// Shout points. In the client Booming Voice (12321) widens the radius only, the improved talent
// does not exist, and the shout states -205 attack power for 45 seconds (11556). Pending the
// shared shout aura rework, both are passed as 0.
func (warrior *Warrior) registerDemoralizingShout() {
	warrior.DemoralizingShoutAuras = warrior.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.DemoralizingShoutAura(target, 0, 0)
	})

	config := spelldata.SpellConfig(&warrior.Unit, demoralizingShoutRank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ThreatMultiplier = 1
	// TODO: Ingame research needed if this adds flat threat
	config.FlatThreatBonus = 0

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
			result := spell.CalcAndDealOutcome(sim, aoeTarget, spell.OutcomeMagicHit)
			if result.Landed() {
				warrior.DemoralizingShoutAuras.Get(aoeTarget).Activate(sim)
			}
		}
	}

	config.RelatedAuraArrays = warrior.DemoralizingShoutAuras.ToMap()

	warrior.DemoralizingShout = warrior.RegisterSpell(config)
}
