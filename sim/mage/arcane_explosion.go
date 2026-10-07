package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (mage *Mage) registerArcaneExplosionSpell() {
	spellData.ArcaneExplosion.Each(mage.registerArcaneExplosionRank)
}

func (mage *Mage) registerArcaneExplosionRank(_ int32, rank *spelldata.Spell) {
	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellArcaneExplosion

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, rollDamage(rank.DamageEffect()))
	}

	mage.RegisterSpell(config)
}
