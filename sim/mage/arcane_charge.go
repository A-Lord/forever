package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Arcane Blast's stacking aura (400573): +10% arcane damage and +175% Arcane Blast mana
// cost per stack, up to 4. The three modifiers are the row's own, so ParseEffects attaches
// them and follows the stacks.
var arcaneChargesRank = spellData.ArcaneBlastTriggered.Highest()

func (mage *Mage) registerArcaneCharges() {
	if !mage.Talents.ArcaneBlast {
		return
	}

	mage.ArcaneChargesAura = core.BlockPrepull(mage.GetOrRegisterAura(spelldata.AuraConfig(arcaneChargesRank,
		spelldata.Label("Arcane Blast Charges"))))
	spelldata.ParseEffects(&mage.Character, mage.ArcaneChargesAura, arcaneChargesRank)
}
