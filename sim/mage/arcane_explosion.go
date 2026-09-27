package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var arcaneExplosionRank = spellData.ArcaneExplosion.Highest()

func (mage *Mage) registerArcaneExplosionSpell() {
	config := spelldata.SpellConfig(&mage.Unit, arcaneExplosionRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellArcaneExplosion

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, rollDamage(arcaneExplosionRank.DamageEffect()))
	}

	mage.RegisterSpell(config)
}
