package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerDesperatePrayer(rank *spelldata.Spell) {
	heal := rank.HealEffect()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellHealing))
	config.ClassSpellMask = PriestSpellDesperatePrayer
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealHealing(sim, &priest.Unit, heal.Roll(sim), spell.OutcomeHealingCrit)
	}

	priest.RegisterSpell(config)
}
