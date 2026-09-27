package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var frostNovaRank = spellData.FrostNova.Highest()
var frostNovaDamage = frostNovaRank.DamageEffect()

// Frost Nova
// https://www.wowhead.com/forever/spell=10230
//
// Frost damage to every enemy around the caster, and a root. The root does nothing to a boss the
// sim fights and nothing reads a frozen target yet, so only the damage is modelled.
func (mage *Mage) registerFrostNovaSpell() {
	// Binary: a spell with a non-damage effect (the root) is resisted whole or not at all, which the
	// row does not state.
	config := spelldata.SpellConfig(&mage.Unit, frostNovaRank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Flags(core.SpellFlagBinary))
	config.ClassSpellMask = MageSpellFrostNova

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, func(sim *core.Simulation, _ *core.Spell) float64 {
			return frostNovaDamage.Roll(sim, core.CharacterLevel)
		})
	}

	mage.RegisterSpell(config)
}
