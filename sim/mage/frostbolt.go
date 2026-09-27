package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var frostboltRank = spellData.Frostbolt.Highest()
var frostboltDamage = frostboltRank.EffectN(2)

func (mage *Mage) registerFrostboltSpell() {
	config := spelldata.SpellConfig(&mage.Unit, frostboltRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellFrostbolt

	config.ThreatMultiplier = 1

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, frostboltDamage.Roll(sim, core.CharacterLevel), spell.OutcomeMagicHitAndCrit)

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
		})
	}

	mage.RegisterSpell(config)
}
