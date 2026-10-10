package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerHeal(_ int32, rank *spelldata.Spell) {
	heal := rank.HealEffect()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellHealing))
	config.ClassSpellMask = PriestSpellHeal
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		if target == nil || target.Type == core.EnemyUnit {
			target = &priest.Unit
		}
		priest.withRenewedHope(sim, target, spell, func() {
			spell.CalcAndDealHealing(sim, target, heal.Average(), spell.OutcomeHealingCrit)
		})
	}

	priest.RegisterSpell(config)
}
