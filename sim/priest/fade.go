package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerFade(rank *spelldata.Spell) {
	threatDrop := rank.EffectN(1).Average()
	var spell *core.Spell

	// A_MOD_TOTAL_THREAT lowers the priest's threat on every enemy while the aura is up and gives it
	// back when it ends. The sim keeps threat as each spell's running total, so the drop is booked
	// against Fade on gain and returned on expire.
	auraConfig := spelldata.AuraConfig(rank, spelldata.Label(rank.Name+" "+rank.Rank))
	auraConfig.Tag = rank.Name
	auraConfig.OnGain = func(_ *core.Aura, _ *core.Simulation) {
		spell.ApplyAOEThreatIgnoreMultipliers(threatDrop)
	}
	auraConfig.OnExpire = func(_ *core.Aura, _ *core.Simulation) {
		spell.ApplyAOEThreatIgnoreMultipliers(-threatDrop)
	}
	aura := priest.RegisterAura(auraConfig)

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ClassSpellMask = PriestSpellFade
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		activateRank(sim, aura)
	}
	config.RelatedSelfBuff = aura

	spell = priest.RegisterSpell(config)
}
