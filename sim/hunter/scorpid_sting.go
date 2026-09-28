package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var scorpidStingRank = spellData.ScorpidSting.Highest()

func (hunter *Hunter) registerScorpidSting() {
	auras := hunter.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		aura := target.RegisterAura(spelldata.AuraConfig(scorpidStingRank,
			spelldata.Label("Scorpid Sting-"+hunter.Label)))
		aura.Tag = "Sting"
		spelldata.ParseEffects(&hunter.Character, aura, scorpidStingRank)
		return aura
	})

	config := spelldata.SpellConfig(&hunter.Unit, scorpidStingRank, spelldata.Flags(core.SpellFlagAPL))
	// See Serpent Sting for the mask.
	config.ProcMask = core.ProcMaskEmpty

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeRangedHit)

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			if result.Landed() {
				aura := auras.Get(target)
				dropOtherSting(sim, aura)
				aura.Activate(sim)
			}
			spell.DealOutcome(sim, result)
		})
	}

	config.RelatedAuraArrays = auras.ToMap()

	hunter.RegisterSpell(config)
}
