package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var coldSnapRank = spellData.ColdSnap.Highest()

// Cold Snap
// https://www.wowhead.com/forever/spell=12472
//
// "Finishes the remaining cooldown on all your other Frost spells." The row's effect is a dummy, so
// "Frost spells" is read as the school: Frost Nova, Cone of Cold, Ice Block and Ice Barrier.
func (mage *Mage) registerColdSnapSpell() {
	config := spelldata.SpellConfig(&mage.Unit, coldSnapRank, spelldata.Flags(core.SpellFlagAPL))
	config.ClassSpellMask = MageSpellColdSnap

	config.ApplyEffects = func(_ *core.Simulation, _ *core.Unit, coldSnap *core.Spell) {
		for _, spell := range mage.Spellbook {
			if spell != coldSnap && spell.SpellSchool.Matches(core.SpellSchoolFrost) && spell.CD.Timer != nil {
				spell.CD.Reset()
			}
		}
	}

	mage.RegisterSpell(config)
}
