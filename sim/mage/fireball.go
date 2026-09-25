package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var fireballRank = spellData.Fireball.Highest()
var fireballBaseDamage = fireballRank.DamageEffect().Average(core.CharacterLevel)

func (mage *Mage) registerFireballSpell() {
	spellConfig := spelldata.SpellConfig(&mage.Unit, fireballRank, spelldata.Magic(core.ProcMaskSpellDamage))
	spellConfig.ClassSpellMask = MageSpellFireball

	spellConfig.Dot = spelldata.DotConfig(fireballRank, fireballRank.PeriodicEffect())

	spellConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, fireballBaseDamage, spell.OutcomeMagicHitAndCrit)
		spell.WaitTravelTime(sim, func(s *core.Simulation) {
			spell.DealDamage(sim, result)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
			}
		})
	}

	mage.RegisterSpell(spellConfig)
}
