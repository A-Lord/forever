package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var fireballRank = spellData.Fireball.Highest()
var fireballDamage = fireballRank.DamageEffect()

func (mage *Mage) registerFireballSpell() {
	fireballTick := fireballRank.PeriodicEffect()

	config := spelldata.SpellConfig(&mage.Unit, fireballRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellFireball

	config.Dot = spelldata.DotConfig(fireballRank, fireballTick, spelldata.Label("FireballDoT"))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := fireballDamage.Roll(sim, core.CharacterLevel)
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
			}
		})
	}

	mage.RegisterSpell(config)
}
