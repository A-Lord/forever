package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var coneOfColdRank = spellData.ConeOfCold.Highest()
var coneOfColdDamage = coneOfColdRank.DamageEffect()

func (mage *Mage) registerConeOfColdSpell() {
	config := spelldata.SpellConfig(&mage.Unit, coneOfColdRank, spelldata.Magic(core.ProcMaskSpellDamage))
	// Binary because of the chill the row's first effect applies.
	config.Flags |= core.SpellFlagBinary
	config.ClassSpellMask = MageSpellConeOfCold

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		baseDamage := coneOfColdDamage.Roll(sim, core.CharacterLevel)
		spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)
	}

	mage.RegisterSpell(config)
}
