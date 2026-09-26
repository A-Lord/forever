package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Frostfire Bolt is content Forever added; ranks 401502, 1237312, 1237313.
var frostfireBoltRank = spellData.FrostfireBolt.Highest()
var frostfireBoltDamage = frostfireBoltRank.DamageEffect()

func (mage *Mage) registerFrostfireBolt() {
	config := spelldata.SpellConfig(&mage.Unit, frostfireBoltRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellFrostfireBolt

	config.Dot = spelldata.DotConfig(frostfireBoltRank, frostfireBoltRank.PeriodicEffect(),
		spelldata.Label("FrostfireBoltDoT"))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := frostfireBoltDamage.Roll(sim, core.CharacterLevel)
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
