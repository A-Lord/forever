package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var pyroblastRank = spellData.Pyroblast.Highest()
var pyroblastDamage = pyroblastRank.DamageEffect()

func (mage *Mage) registerPyroblastSpell() {
	if !mage.Talents.Pyroblast {
		return
	}

	config := spelldata.SpellConfig(&mage.Unit, pyroblastRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellPyroblast

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := pyroblastDamage.Roll(sim, core.CharacterLevel)
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
			if result.Landed() {
				spell.RelatedDotSpell.Cast(sim, target)
			}
		})
	}

	mage.Pyroblast = mage.RegisterSpell(config)

	// The dot rides its own spell so its ticks report under their own tag.
	dotConfig := spelldata.SpellConfig(&mage.Unit, pyroblastRank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc(), spelldata.Tag(1))
	dotConfig.ClassSpellMask = MageSpellPyroblastDot
	dotConfig.BonusCoefficient = 0 // the direct hit's share; the dot's own is on the DotConfig

	dotConfig.Dot = spelldata.DotConfig(pyroblastRank, pyroblastRank.PeriodicEffect(),
		spelldata.Label("PyroblastDoT"))

	dotConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.Dot(target).Apply(sim)
	}

	mage.Pyroblast.RelatedDotSpell = mage.RegisterSpell(dotConfig)
}
