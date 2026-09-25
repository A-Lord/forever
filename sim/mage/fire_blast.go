package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// TODO: To be implemented. spellData.FireBlast holds the seven trainer ranks, 2136 to 10199.
// The client also carries 400616 to 400623, the copies the Season of Discovery rune passive
// Overheat (400615) swaps onto the action bar. Overheat is an Engrave grant with no place in
// Forever, and the generator drops its stand-ins.

var fireBlastRank = spellData.FireBlast.Highest()
var fireBlastBaseDamage = fireBlastRank.DamageEffect().Average(core.CharacterLevel)

func (mage *Mage) registerFireBlastSpell() {
	config := spelldata.SpellConfig(&mage.Unit, fireBlastRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellFireBlast

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, fireBlastBaseDamage, spell.OutcomeMagicHitAndCrit)
	}

	mage.RegisterSpell(config)
}
