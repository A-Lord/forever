package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var frostboltRank = spellData.Frostbolt.Highest()
var frostboltBaseDamage = frostboltRank.EffectN(2).Average(core.CharacterLevel)

func (mage *Mage) registerFrostboltSpell() {
	config := spelldata.SpellConfig(&mage.Unit, frostboltRank, spelldata.Magic(core.ProcMaskSpellDamage))

	config.ThreatMultiplier = 1

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, frostboltBaseDamage, spell.OutcomeMagicHitAndCrit)

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
		})
	}

	mage.RegisterSpell(config)
}
