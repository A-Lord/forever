package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var arcaneExplosionRank = spellData.ArcaneExplosion.Highest()
var arcaneExplosionDamage = arcaneExplosionRank.DamageEffect()

func (mage *Mage) registerArcaneExplosionSpell() {
	config := spelldata.SpellConfig(&mage.Unit, arcaneExplosionRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellArcaneExplosion

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		baseDamage := arcaneExplosionDamage.Roll(sim, core.CharacterLevel)
		spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)
	}

	mage.RegisterSpell(config)
}
