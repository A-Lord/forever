package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var arcaneBlastRank = spellData.ArcaneBlast.Highest()
var arcaneBlastDamage = arcaneBlastRank.DamageEffect()

func (mage *Mage) registerArcaneBlastSpell() {

	config := spelldata.SpellConfig(&mage.Unit, arcaneBlastRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellArcaneBlast

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, arcaneBlastDamage.Roll(sim, core.CharacterLevel), spell.OutcomeMagicHitAndCrit)
		if result.Landed() {
			mage.ArcaneChargesAura.Activate(sim)
			mage.ArcaneChargesAura.AddStack(sim)
		}
	}

	mage.ArcaneBlast = mage.RegisterSpell(config)
}
