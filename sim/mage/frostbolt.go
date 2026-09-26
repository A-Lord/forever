package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var frostboltRank = spellData.Frostbolt.Highest()
var frostboltDamage = frostboltRank.DamageEffect()

func (mage *Mage) registerFrostboltSpell() {
	config := spelldata.SpellConfig(&mage.Unit, frostboltRank, spelldata.Magic(core.ProcMaskSpellDamage))
	// Binary because of the chill the row's first effect applies.
	config.Flags |= core.SpellFlagBinary
	config.ClassSpellMask = MageSpellFrostbolt

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := frostboltDamage.Roll(sim, core.CharacterLevel)
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
		})
	}

	mage.RegisterSpell(config)
}
