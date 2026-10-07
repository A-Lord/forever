package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var iceLanceRank = spellData.IceLance.Highest()
var iceLanceDamage = iceLanceRank.DamageEffect()

// "Deals $s2% increased damage to Frozen targets": effect 2, 300.
var iceLanceFrozenMultiplier = 1 + iceLanceRank.EffectN(2).Percent()

// Ice Lance
// https://www.wowhead.com/forever/spell=1240047
//
// "Deals 146 Frost damage to an enemy target. Deals 300% increased damage to Frozen targets."
//
// The damage effect states no spell power coefficient; the store carries one as an override
// (tools/database/overrides). Effects 3 (20) and 4 (5) are server-side scripts the tooltip does not
// name, and are not modelled.
func (mage *Mage) registerIceLanceSpell() {
	config := spelldata.SpellConfig(&mage.Unit, iceLanceRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellIceLance

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		frozenMultiplier := 1.0
		if mage.isFrozen(target) {
			frozenMultiplier = iceLanceFrozenMultiplier
		}

		spell.DamageMultiplier *= frozenMultiplier
		result := spell.CalcDamage(sim, target, iceLanceDamage.Average(), spell.OutcomeMagicHitAndCrit)
		spell.DamageMultiplier /= frozenMultiplier

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
		})
	}

	mage.RegisterSpell(config)
}
