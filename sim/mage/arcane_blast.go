package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (mage *Mage) registerArcaneBlastSpell() {
	spellData.ArcaneBlast.Each(mage.registerArcaneBlastRank)
}

func (mage *Mage) registerArcaneBlastRank(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()

	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellArcaneBlast

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
		if result.Landed() {
			mage.ArcaneChargesAura.Activate(sim)
			mage.ArcaneChargesAura.AddStack(sim)
		}
	}

	mage.RegisterSpell(config)
}
