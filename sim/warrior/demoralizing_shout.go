package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var demoralizingShoutRank = spellData.DemoralizingShout.Highest()

func (warrior *Warrior) registerDemoralizingShout() {
	warrior.DemoralizingShoutAuras = warrior.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		// Nothing in the warrior tree prices this shout: Forever has no Improved
		// Demoralizing Shout, and Booming Voice widens the radius only, so the
		// aura is the client's attack power reduction for 45 seconds.
		return buffs.DemoralizingShoutAura(target, true, 0)
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
