package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var rendRank = spellData.Rend.Highest()

// TODO: Ingame testing needed if Rend has a coef
func (warrior *Warrior) registerRend() {
	tick := rendRank.PeriodicEffect()

	config := spelldata.SpellConfig(&warrior.Unit, rendRank,
		spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.Flags(core.SpellFlagNoOnCastComplete))

	config.Dot = spelldata.DotConfig(rendRank, tick)

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		} else {
			spell.IssueRefund(sim)
		}
	}

	warrior.Rend = warrior.RegisterSpell(config)
}
