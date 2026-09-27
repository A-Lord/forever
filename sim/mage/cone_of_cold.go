package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var coneOfColdRank = spellData.ConeOfCold.Highest()

// The slow (effect 1) does nothing to a boss.
func (mage *Mage) registerConeOfColdSpell() {
	// Binary: a spell with a non-damage effect (the slow) is resisted whole or not at all, which the
	// row does not state.
	config := spelldata.SpellConfig(&mage.Unit, coneOfColdRank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Flags(core.SpellFlagBinary))
	config.ClassSpellMask = MageSpellConeOfCold

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, rollDamage(coneOfColdRank.DamageEffect()))
	}

	mage.RegisterSpell(config)
}
