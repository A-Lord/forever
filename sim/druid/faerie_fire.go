package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// TODO: Faerie Fire does not reset the swing timer.
func (druid *Druid) registerFaerieFireSpell() {
	rank := spellData.FaerieFire.Highest()
	druid.FaerieFireAuras = druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataFaerieFireAura(target, true, 0)
	})
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	// Cost, cooldown and GCD in cat or bear come from the form passive's spell mods.
	config.ExtraCastCondition = func(_ *core.Simulation, target *core.Unit) bool {
		return druid.FaerieFireAuras == nil || !druid.FaerieFireAuras.Get(target).IsActive()
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
		if result.Landed() {
			druid.FaerieFireAuras.Get(target).Activate(sim)
		}
	}
	config.RelatedAuraArrays = druid.FaerieFireAuras.ToMap()
	druid.FaerieFire = druid.RegisterSpell(config)
}

func (druid *Druid) registerFaerieFireFeralSpell() {
}
