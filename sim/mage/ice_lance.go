package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var iceLanceRank = spellData.IceLance.Highest()
var iceLanceDamage = iceLanceRank.DamageEffect()

// The frozen-target multiplier is the row's second effect, a dummy holding 300 (percent).
var iceLanceFrozenMultiplier = iceLanceRank.EffectN(2).Percent()

func (mage *Mage) registerIceLanceSpell() {
	if !mage.Talents.IceLance {
		return
	}

	config := spelldata.SpellConfig(&mage.Unit, iceLanceRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellIceLance
	// The row states no spell power share; the client's coefficient is 0.1428571428571429 (1/7).
	config.BonusCoefficient = 0.1428571428571429

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		frozen := mage.FingersOfFrostAura != nil && mage.FingersOfFrostAura.IsActive()
		if frozen {
			spell.DamageMultiplier *= iceLanceFrozenMultiplier
		}

		baseDamage := iceLanceDamage.Roll(sim, core.CharacterLevel)
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

		if frozen {
			spell.DamageMultiplier /= iceLanceFrozenMultiplier
			mage.FingersOfFrostAura.RemoveStack(sim)
		}

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
		})
	}

	mage.RegisterSpell(config)
}
