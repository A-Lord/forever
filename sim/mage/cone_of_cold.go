package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var coneOfColdRank = spellData.ConeOfCold.Highest()

// The slow (effect 1) does nothing to a boss.
func (mage *Mage) registerConeOfColdSpell() {
	config := spelldata.SpellConfig(&mage.Unit, coneOfColdRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellConeOfCold

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, rollDamage(coneOfColdRank.DamageEffect()))
	}

	mage.RegisterSpell(config)
}
