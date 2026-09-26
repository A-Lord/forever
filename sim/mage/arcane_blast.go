package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var arcaneBlastRank = spellData.ArcaneBlast.Highest()
var arcaneBlastDamage = arcaneBlastRank.DamageEffect()

// Forever makes Arcane Blast a talent (400574 and its ranks, the Season of Discovery
// version): each cast stacks the Arcane Blast aura registered in arcane_charge.go.
func (mage *Mage) registerArcaneBlastSpell() {
	if !mage.Talents.ArcaneBlast {
		return
	}

	config := spelldata.SpellConfig(&mage.Unit, arcaneBlastRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellArcaneBlast

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := arcaneBlastDamage.Roll(sim, core.CharacterLevel)
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		if result.Landed() {
			mage.ArcaneChargesAura.Activate(sim)
			mage.ArcaneChargesAura.AddStack(sim)
		}
	}

	mage.ArcaneBlast = mage.RegisterSpell(config)
}
