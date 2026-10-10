package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerPowerWordShield(_ int32, rank *spelldata.Spell) {
	baseAbsorb := rank.AbsorbEffect()
	weakenedSoul := spellData.PowerWordShieldTriggered.ByID(6788)
	if priest.weakenedSoulAuras == nil {
		priest.weakenedSoulAuras = priest.NewAllyAuraArray(func(target *core.Unit) *core.Aura {
			return target.GetOrRegisterAura(spelldata.AuraConfig(weakenedSoul))
		})
	}

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellHealing))
	config.ClassSpellMask = PriestSpellPowerWordShield
	config.RelatedAuraArrays = priest.weakenedSoulAuras.ToMap()
	config.ExtraCastCondition = func(_ *core.Simulation, target *core.Unit) bool {
		target = priest.healTarget(target)
		return !priest.weakenedSoulAuras.Get(target).IsActive()
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		target = priest.healTarget(target)
		priest.weakenedSoulAuras.Get(target).Activate(sim)
		spell.CalcAndDealHealing(sim, target, baseAbsorb.Average(), spell.OutcomeHealing)
	}

	priest.RegisterSpell(config)
}
