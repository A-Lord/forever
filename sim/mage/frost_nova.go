package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var frostNovaRank = spellData.FrostNova.Highest()
var frostNovaDamage = frostNovaRank.DamageEffect()

func (mage *Mage) registerFrostNovaSpell() {
	config := spelldata.SpellConfig(&mage.Unit, frostNovaRank, spelldata.Magic(core.ProcMaskSpellDamage))
	// Binary because of the root the row's second effect applies; the root itself does nothing
	// to a raid boss, so only the damage is dealt.
	config.Flags |= core.SpellFlagBinary
	config.ClassSpellMask = MageSpellFrostNova

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		baseDamage := frostNovaDamage.Roll(sim, core.CharacterLevel)
		spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)
	}

	mage.RegisterSpell(config)
}
