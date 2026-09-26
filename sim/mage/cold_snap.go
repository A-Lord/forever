package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var coldSnapRank = spellData.ColdSnap.Highest()

func (mage *Mage) registerColdSnapSpell() {
	if !mage.Talents.ColdSnap {
		return
	}

	config := spelldata.SpellConfig(&mage.Unit, coldSnapRank)
	config.Flags |= core.SpellFlagNoOnCastComplete | core.SpellFlagAPL
	config.ClassSpellMask = MageSpellColdSnap

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		// The row is a dummy: it finishes the cooldown of every Frost spell.
		for _, spell := range mage.Spellbook {
			if spell.SpellSchool.Matches(core.SpellSchoolFrost) && spell.CD.Timer != nil {
				spell.CD.Reset()
			}
		}
	}

	mage.RegisterSpell(config)
}
